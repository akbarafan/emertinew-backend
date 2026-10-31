package model

import "time"

// MAlertEvent is one row of the alert audit trail.
//
// The detection rules live in the command-center frontend, which derives them
// from polling telemetry and POSTs an event the first time a condition is seen.
// `EventKey` is supplied by that client so a retry is an upsert rather than a
// duplicate, and `DetectedAt` is the *tracking* timestamp that showed the
// condition — stamping it with the poll time is what previously made the feed
// unreadable, since every row then carried roughly the same clock.
type MAlertEvent struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	EventKey   string `gorm:"column:event_key;type:varchar(120);not null;uniqueIndex" json:"event_key"`
	VesselID   uint   `gorm:"column:vessel_id;not null" json:"vessel_id"`
	VesselName string `gorm:"column:vessel_name;type:varchar(150);not null" json:"vessel_name"`
	Type       string `gorm:"column:type;type:varchar(40);not null" json:"type"`
	Severity   string `gorm:"column:severity;type:varchar(10);not null" json:"severity"`
	Message    string `gorm:"column:message;not null" json:"message"`

	DetectedAt time.Time  `gorm:"column:detected_at;not null" json:"detected_at"`
	ResolvedAt *time.Time `gorm:"column:resolved_at" json:"resolved_at,omitempty"`
	AckAt      *time.Time `gorm:"column:ack_at" json:"ack_at,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// Alert severities, mirroring the `SEVERITY` map in `cc-alerts-list.tsx`.
const (
	AlertSeverityInfo    = "info"
	AlertSeverityWarning = "warning"
	AlertSeverityError   = "error"
)

func (MAlertEvent) TableName() string {
	return "m_alert_events"
}

// IsOpen reports whether the condition is still present. An open row with no
// `ResolvedAt` is what the live badge counts.
func (a MAlertEvent) IsOpen() bool {
	return a.ResolvedAt == nil
}
