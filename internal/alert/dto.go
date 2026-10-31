package alert

import "time"

// UpsertEventRequest is one alert observed by the command center.
//
// The client only sends a row on the absent -> present transition, never on
// every poll: flow sensors publish every 5 seconds, so a naive "send what you
// see" would write roughly 17k duplicate rows per vessel per day.
type UpsertEventRequest struct {
	EventKey   string `json:"event_key" binding:"required,max=120" example:"temp-alert-12"`
	VesselID   uint   `json:"vessel_id" binding:"required" example:"1"`
	VesselName string `json:"vessel_name" binding:"required,max=150" example:"KM Nusantara 01"`
	Type       string `json:"type" binding:"required,max=40" example:"temp_alert"`
	Severity   string `json:"severity" binding:"required,oneof=info warning error" example:"warning"`
	Message    string `json:"message" binding:"required" example:"High temperature — Engine Rm 71°C > 60°C"`

	// DetectedAt must be the tracking timestamp that showed the condition.
	// Using the poll time here is what made every row in the feed show the same
	// clock and made the sort by time meaningless.
	DetectedAt time.Time `json:"detected_at" binding:"required"`
	// ResolvedAt is set when a later poll shows the condition cleared.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// UpsertEventsRequest is the bulk form the client uses each poll.
type UpsertEventsRequest struct {
	Events []UpsertEventRequest `json:"events"`
	// Complete states that `events` is the FULL set of conditions the client can
	// currently see, not a partial batch. Only then may the server close alerts
	// the batch omitted.
	//
	// Without this distinction an empty batch is ambiguous: it could mean "the
	// fleet is clear" or "this client has not finished loading yet". Treating
	// the second as the first would wipe the audit trail every time a browser
	// tab started up.
	Complete bool `json:"complete" example:"true"`
}

// AckEventRequest acknowledges one alert without resolving it.
type AckEventRequest struct {
	// Note is optional free text kept alongside the acknowledgement.
	Note string `json:"note,omitempty"`
}

// ListEventsQuery filters the history feed.
//
// `OpenOnly` is what the widget polls: unresolved rows drive the live badge.
// The rest are for the operator-facing history view.
type ListEventsQuery struct {
	VesselID uint   `form:"vessel_id"`
	Type     string `form:"type"`
	Severity string `form:"severity"`
	From     string `form:"from" example:"2026-04-30"`
	To       string `form:"to" example:"2026-05-01"`
	OpenOnly bool   `form:"open_only"`
	Page     int    `form:"page" example:"1"`
	Limit    int    `form:"limit" example:"50"`

	// FromTime and ToTime are the parsed, WIB-anchored form of From and To,
	// filled in by the service. Comparing a `timestamptz` column against the raw
	// string relies on Postgres guessing the type, which silently changes
	// meaning between "2026-04-30" and "2026-04-30 00:00:00". `To` is exclusive.
	FromTime time.Time `form:"-" json:"-"`
	ToTime   time.Time `form:"-" json:"-"`
}

// ListEventsResponse is a page of history plus the counters the widget header
// needs, so the client does not have to fetch everything to render "3 open".
type ListEventsResponse struct {
	Events    []EventDTO `json:"events"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
	OpenCount int        `json:"open_count"`
}

// EventDTO is the wire shape of one alert row.
type EventDTO struct {
	ID         uint       `json:"id"`
	EventKey   string     `json:"event_key"`
	VesselID   uint       `json:"vessel_id"`
	VesselName string     `json:"vessel_name"`
	Type       string     `json:"type"`
	Severity   string     `json:"severity"`
	Message    string     `json:"message"`
	DetectedAt time.Time  `json:"detected_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	AckAt      *time.Time `json:"ack_at,omitempty"`
	IsOpen     bool       `json:"is_open"`
}
