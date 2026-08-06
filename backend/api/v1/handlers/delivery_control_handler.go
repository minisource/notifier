package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"
)

// DeliveryControlHandler exposes the Global Outbound Delivery Pause / Emergency
// Freeze controls to authorized admins. Pause/Resume are high-risk actions;
// the backend enforces the pause regardless of UI state.
type DeliveryControlHandler struct {
	deliveryControl *service.DeliveryControlService
	notifRepo       repository.NotificationRepository
}

// NewDeliveryControlHandler creates the handler.
func NewDeliveryControlHandler(
	deliveryControl *service.DeliveryControlService,
	notifRepo repository.NotificationRepository,
) *DeliveryControlHandler {
	return &DeliveryControlHandler{deliveryControl: deliveryControl, notifRepo: notifRepo}
}

// Status godoc
// @Summary Get global delivery control status
// @Description Return the current global outbound delivery pause state with held/uncertain counts
// @Tags DeliveryControl
// @Security BearerAuth
// @Success 200 {object} dto.DeliveryControlStatusDTO
// @Router /admin/delivery-control/status [get]
func (h *DeliveryControlHandler) Status(c *fiber.Ctx) error {
	state, err := h.deliveryControl.CurrentState(c.Context())
	if err != nil {
		return response.InternalError(c, "Failed to read delivery control state")
	}
	return response.OK(c, h.buildStatus(c, state))
}

// Pause godoc
// @Summary Pause all outbound delivery
// @Description Freeze all outbound provider execution (SMS/email/push/webhook/retries/fallbacks). Reason is mandatory. Supports Idempotency-Key and ExpectedVersion (stale -> 409).
// @Tags DeliveryControl
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key (replays return the original result)"
// @Param body body dto.PauseRequestDTO true "Pause request"
// @Success 200 {object} dto.DeliveryControlStatusDTO
// @Router /admin/delivery-control/pause [post]
func (h *DeliveryControlHandler) Pause(c *fiber.Ctx) error {
	req := new(dto.PauseRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid pause payload")
	}
	actor := actorIdentity(c)
	idemKey := c.Get("Idempotency-Key")
	requestHash := pauseRequestHash(req)

	// Replay protection: same actor + same key + same payload returns the
	// original result; same key + different payload is rejected.
	if resultJSON, replay, err := h.deliveryControl.ReplayIdempotency(c.Context(), actor, idemKey, "pause", requestHash); err != nil {
		return response.Conflict(c, "IDEMPOTENCY_CONFLICT: "+err.Error())
	} else if replay {
		var cached dto.DeliveryControlStatusDTO
		if json.Unmarshal([]byte(resultJSON), &cached) == nil {
			return response.OK(c, cached)
		}
		// Stored JSON malformed — fall through and recompute below.
	}

	// Stale-version protection: if the caller submits a version it saw and the
	// state changed meanwhile, reject with 409 (no last-write-wins).
	if req.ExpectedVersion > 0 {
		if err := h.deliveryControl.ValidateExpectedVersion(c.Context(), req.ExpectedVersion); err != nil {
			if errors.Is(err, service.ErrControlStateConflict) {
				return response.Conflict(c, "CONTROL_STATE_CONFLICT: another operator changed the delivery control state; reload and confirm again")
			}
			return response.InternalError(c, "Failed to validate delivery control version")
		}
	}

	state, err := h.deliveryControl.RequestPause(c.Context(), actor, req.Mode, req.Reason, req.ExpiresAt)
	if err != nil {
		return response.BadRequest(c, "PAUSE_REJECTED", err.Error())
	}

	status := h.buildStatus(c, state)
	if b, err := json.Marshal(status); err == nil {
		h.deliveryControl.RecordIdempotency(c.Context(), actor, idemKey, "pause", requestHash, string(b), state.State, state.Version)
	}
	return response.OK(c, status)
}

// Resume godoc
// @Summary Resume outbound delivery
// @Description Resume frozen outbound delivery; held work is released gradually. Supports Idempotency-Key and ExpectedVersion (stale -> 409).
// @Tags DeliveryControl
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key (replays return the original result)"
// @Param body body dto.ResumeRequestDTO false "Resume request"
// @Success 200 {object} dto.DeliveryControlStatusDTO
// @Router /admin/delivery-control/resume [post]
func (h *DeliveryControlHandler) Resume(c *fiber.Ctx) error {
	req := new(dto.ResumeRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid resume payload")
	}
	actor := actorIdentity(c)
	idemKey := c.Get("Idempotency-Key")
	requestHash := resumeRequestHash(req)

	if resultJSON, replay, err := h.deliveryControl.ReplayIdempotency(c.Context(), actor, idemKey, "resume", requestHash); err != nil {
		return response.Conflict(c, "IDEMPOTENCY_CONFLICT: "+err.Error())
	} else if replay {
		var cached dto.DeliveryControlStatusDTO
		if json.Unmarshal([]byte(resultJSON), &cached) == nil {
			return response.OK(c, cached)
		}
	}

	if req.ExpectedVersion > 0 {
		if err := h.deliveryControl.ValidateExpectedVersion(c.Context(), req.ExpectedVersion); err != nil {
			if errors.Is(err, service.ErrControlStateConflict) {
				return response.Conflict(c, "CONTROL_STATE_CONFLICT: another operator changed the delivery control state; reload and confirm again")
			}
			return response.InternalError(c, "Failed to validate delivery control version")
		}
	}

	state, err := h.deliveryControl.RequestResume(c.Context(), actor, req.Reason)
	if err != nil {
		return response.BadRequest(c, "RESUME_REJECTED", err.Error())
	}

	status := h.buildStatus(c, state)
	if b, err := json.Marshal(status); err == nil {
		h.deliveryControl.RecordIdempotency(c.Context(), actor, idemKey, "resume", requestHash, string(b), state.State, state.Version)
	}
	return response.OK(c, status)
}

// History godoc
// @Summary List delivery control audit events
// @Description Return pause/resume audit history newest-first
// @Tags DeliveryControl
// @Security BearerAuth
// @Param limit query int false "Max events (default 50)"
// @Success 200 {array} dto.DeliveryControlEventDTO
// @Router /admin/delivery-control/history [get]
func (h *DeliveryControlHandler) History(c *fiber.Ctx) error {
	limit := 50
	if v, err := strconv.Atoi(c.Query("limit", "50")); err == nil && v > 0 && v <= 200 {
		limit = v
	}
	events, err := h.deliveryControl.ListEvents(c.Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to list delivery control history")
	}
	out := make([]*dto.DeliveryControlEventDTO, 0, len(events))
	for _, e := range events {
		out = append(out, &dto.DeliveryControlEventDTO{
			ID:        e.ID.String(),
			Action:    e.Action,
			Actor:     e.Actor,
			Reason:    e.Reason,
			Mode:      e.Mode,
			FromState: e.FromState,
			ToState:   e.ToState,
			Version:   e.Version,
			RequestID: e.RequestID,
			CreatedAt: e.CreatedAt,
		})
	}
	return response.OK(c, out)
}

// Held godoc
// @Summary List held deliveries
// @Description Return deliveries currently frozen by the global pause (oldest first)
// @Tags DeliveryControl
// @Security BearerAuth
// @Param page query int false "Page (default 1)"
// @Param pageSize query int false "Page size (default 20, max 100)"
// @Success 200 {object} map[string]interface{}
// @Router /admin/delivery-control/held [get]
func (h *DeliveryControlHandler) Held(c *fiber.Ctx) error {
	page := parsePage(c.Query("page"))
	pageSize := parsePageSize(c.Query("pageSize"))
	held, total, err := h.notifRepo.ListHeld(c.Context(), pageSize, (page-1)*pageSize)
	if err != nil {
		return response.InternalError(c, "Failed to list held deliveries")
	}
	return response.OK(c, fiber.Map{
		"items":     held,
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
	})
}

func (h *DeliveryControlHandler) buildStatus(c *fiber.Ctx, state *models.DeliveryControlState) *dto.DeliveryControlStatusDTO {
	held, _ := h.deliveryControl.HeldSummary(c.Context())

	status := &dto.DeliveryControlStatusDTO{
		State:          state.State,
		Mode:           state.Mode,
		Reason:         state.Reason,
		PausedBy:       state.PausedBy,
		PausedAt:       state.PausedAt,
		EffectiveAt:    state.EffectiveAt,
		ExpiresAt:      state.ExpiresAt,
		ResumedBy:      state.ResumedBy,
		ResumedAt:      state.ResumedAt,
		Version:        state.Version,
		HeldCount:      held,
		RetryingHeld:   h.deliveryControl.RetryingHeldCount(c.Context()),
		ActiveAttemptCount: h.deliveryControl.ActiveCount(c.Context()),
		// In-flight cancellation + outcome_unknown tracking is not implemented
		// in the current synchronous execution model, so this is honestly 0
		// (never fabricated). See TODO: outcome_unknown reconciliation.
		UncertainCount: 0,
		LastUpdatedAt:  state.UpdatedAt,
	}
	switch state.State {
	case models.DeliveryControlActive, models.DeliveryControlActiveWithUncertain:
		status.CanPause = true
		status.CanResume = false
	case models.DeliveryControlPaused, models.DeliveryControlPauseRequested:
		status.CanPause = false
		status.CanResume = true
	case models.DeliveryControlResumeRequested:
		status.CanPause = true
		status.CanResume = false
	}
	return status
}

// pauseRequestHash produces a stable hash of the normalized pause payload so
// replay detection can compare requests deterministically.
func pauseRequestHash(req *dto.PauseRequestDTO) string {
	var expires string
	if req.ExpiresAt != nil {
		expires = req.ExpiresAt.UTC().Format(time.RFC3339)
	}
	sum := sha256.Sum256([]byte(req.Mode + "\n" + req.Reason + "\n" + expires))
	return hex.EncodeToString(sum[:])
}

// resumeRequestHash produces a stable hash of the normalized resume payload.
func resumeRequestHash(req *dto.ResumeRequestDTO) string {
	sum := sha256.Sum256([]byte(req.Reason))
	return hex.EncodeToString(sum[:])
}

func actorIdentity(c *fiber.Ctx) string {
	if v := c.Locals("userId"); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if v := c.Locals("userEmail"); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "admin"
}

func parsePage(v string) int {
	p, err := strconv.Atoi(v)
	if err != nil || p < 1 {
		return 1
	}
	return p
}

func parsePageSize(v string) int {
	p, err := strconv.Atoi(v)
	if err != nil || p < 1 {
		return 20
	}
	if p > 100 {
		return 100
	}
	return p
}
