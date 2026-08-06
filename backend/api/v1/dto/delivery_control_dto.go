package dto

import (
	"time"
)

// DeliveryControlStatusDTO is the admin-facing status of the global outbound
// delivery pause. Explicit states distinguish "paused" from "propagating" and
// from "active with uncertain attempts".
type DeliveryControlStatusDTO struct {
	State           string     `json:"state"`                     // active | pause_requested | paused | resume_requested | active_with_uncertain_attempts
	Mode            string     `json:"mode"`                      // immediate | drain
	Reason          string     `json:"reason,omitempty"`          // required when paused
	PausedBy        string     `json:"pausedBy,omitempty"`        // actor identity
	PausedAt        *time.Time `json:"pausedAt,omitempty"`
	EffectiveAt     *time.Time `json:"effectiveAt,omitempty"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`       // optional auto-resume deadline
	ResumedBy       string     `json:"resumedBy,omitempty"`
	ResumedAt       *time.Time `json:"resumedAt,omitempty"`
	Version         int64      `json:"version"`                   // monotonic generation
	HeldCount       int64      `json:"heldCount"`                 // deliveries frozen by the pause
	RetryingHeld    int64      `json:"retryingHeld"`              // retries frozen (preserved budget)
	UncertainCount  int64      `json:"uncertainCount"`            // in-flight attempts with unknown outcome
	ActiveAttemptCount int64   `json:"activeAttemptCount"`        // currently executing attempts
	CanPause        bool       `json:"canPause"`
	CanResume       bool       `json:"canResume"`
	LastUpdatedAt   time.Time  `json:"lastUpdatedAt"`
}

// DeliveryControlEventDTO is one pause/resume audit event.
type DeliveryControlEventDTO struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	Mode      string    `json:"mode,omitempty"`
	FromState string    `json:"fromState,omitempty"`
	ToState   string    `json:"toState,omitempty"`
	Version   int64     `json:"version"`
	RequestID string    `json:"requestId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// PauseRequestDTO is the payload to pause outbound deliveries.
type PauseRequestDTO struct {
	Mode            string     `json:"mode"`                // immediate | drain
	Reason          string     `json:"reason"`              // mandatory
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"` // optional auto-resume deadline
	ExpectedVersion int64      `json:"expectedVersion,omitempty"` // caller's last-known version (stale -> 409)
}

// ResumeRequestDTO is the payload to resume outbound deliveries.
type ResumeRequestDTO struct {
	Reason          string `json:"reason,omitempty"`
	ExpectedVersion int64  `json:"expectedVersion,omitempty"` // caller's last-known version (stale -> 409)
}
