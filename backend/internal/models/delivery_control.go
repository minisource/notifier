package models

import (
	"time"

	"github.com/google/uuid"
)

// Global Notifier Outbound Delivery Pause / Emergency Freeze domain.
//
// The authoritative pause state lives in a single-row table
// (delivery_control_state, ID = "global"). Worker/provider boundaries read
// this state (through a short-TTL cache) and refuse to start any new outbound
// provider operation while it is not "active". Audit events are append-only.

// DeliveryControlState values.
const (
	DeliveryControlActive                = "active"
	DeliveryControlPauseRequested        = "pause_requested"
	DeliveryControlPaused                = "paused"
	DeliveryControlResumeRequested       = "resume_requested"
	DeliveryControlActiveWithUncertain   = "active_with_uncertain_attempts"
)

// DeliveryControlMode values.
const (
	DeliveryControlModeImmediate = "immediate"
	DeliveryControlModeDrain     = "drain"
)

// DeliveryControlEvent actions (audit log).
const (
	DeliveryControlActionPauseRequested  = "pause_requested"
	DeliveryControlActionPauseEffective  = "pause_effective"
	DeliveryControlActionResumeRequested = "resume_requested"
	DeliveryControlActionResumeEffective = "resume_effective"
	DeliveryControlActionAutoResume      = "auto_resume"
	DeliveryControlActionPauseExpired    = "pause_expired"
)

// DeliveryControlGlobalID is the fixed scope key of the global pause row.
const DeliveryControlGlobalID = "global"

// DeliveryControlState is the single authoritative row for the global
// outbound delivery pause. There is exactly one row (ID = "global"); the
// Version field is a monotonic generation used to invalidate worker caches.
type DeliveryControlState struct {
	ID          string     `gorm:"type:varchar(50);primary_key" json:"id"`
	State       string     `gorm:"type:varchar(30);not null;default:active" json:"state"`
	Mode        string     `gorm:"type:varchar(20);not null;default:immediate" json:"mode"`
	Reason      string     `gorm:"type:text" json:"reason,omitempty"`
	PausedBy    string     `gorm:"type:varchar(255)" json:"pausedBy,omitempty"`
	PausedAt    *time.Time `json:"pausedAt,omitempty"`
	EffectiveAt *time.Time `json:"effectiveAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"` // optional auto-resume deadline
	ResumedBy   string     `gorm:"type:varchar(255)" json:"resumedBy,omitempty"`
	ResumedAt   *time.Time `json:"resumedAt,omitempty"`
	Version     int64      `gorm:"not null;default:1" json:"version"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"not null;default:now()" json:"updatedAt"`
}

// TableName specifies the table name.
func (DeliveryControlState) TableName() string { return "delivery_control_state" }

// DeliveryControlEvent is an append-only audit record for pause/resume actions.
type DeliveryControlEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Action    string    `gorm:"type:varchar(50);not null;index" json:"action"`
	Actor     string    `gorm:"type:varchar(255)" json:"actor,omitempty"`
	Reason    string    `gorm:"type:text" json:"reason,omitempty"`
	Mode      string    `gorm:"type:varchar(20)" json:"mode,omitempty"`
	FromState string    `gorm:"type:varchar(30)" json:"fromState,omitempty"`
	ToState   string    `gorm:"type:varchar(30)" json:"toState,omitempty"`
	Version   int64     `gorm:"not null;default:1" json:"version"`
	RequestID string    `gorm:"type:varchar(64)" json:"requestId,omitempty"`
	CreatedAt time.Time `gorm:"not null;default:now();index" json:"createdAt"`
}

// TableName specifies the table name.
func (DeliveryControlEvent) TableName() string { return "delivery_control_events" }

// DeliveryControlIdempotency records one processed Pause/Resume request keyed
// by (actor, idempotency key) so that replayed or browser-retried requests
// return the original result instead of creating a duplicate transition or
// audit event. Rows are short-lived (configurable retention). The composite
// UNIQUE (actor, idempotency_key) makes concurrent identical requests safe:
// exactly one row can exist, so a duplicate insert is a replay, never a
// second transition.
type DeliveryControlIdempotency struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Actor         string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_dc_idem_actor_key" json:"actor"`
	IdempotencyKey string   `gorm:"type:varchar(128);not null;uniqueIndex:idx_dc_idem_actor_key" json:"idempotencyKey"`
	Operation     string    `gorm:"type:varchar(20);not null" json:"operation"` // pause | resume
	RequestHash   string    `gorm:"type:varchar(128);not null" json:"requestHash"`
	State         string    `gorm:"type:varchar(30)" json:"state,omitempty"`
	Version       int64     `json:"version,omitempty"`
	ResultJSON    string    `gorm:"type:text" json:"resultJson,omitempty"` // serialized status DTO returned to the caller
	CreatedAt     time.Time `gorm:"not null;default:now();index" json:"createdAt"`
	ExpiresAt     time.Time `gorm:"not null;default:now()" json:"expiresAt"`
}

// TableName specifies the table name.
func (DeliveryControlIdempotency) TableName() string { return "delivery_control_idempotency" }
