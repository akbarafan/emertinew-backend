package alert

import (
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

type Service interface {
	SyncEvents(events []UpsertEventRequest, keysStillPresent map[string]bool, complete bool) (SyncResult, error)
	ListEvents(query ListEventsQuery) (*ListEventsResponse, error)
	AckEvent(id uint, note string) (*EventDTO, error)
}

// SyncResult reports what one poll's sync changed, so the client can log the
// transition count without a second round trip.
type SyncResult struct {
	Created  int `json:"created"`
	Received int `json:"received"`
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// SyncEvents records one poll's worth of transitions.
//
// `events` is the set of conditions currently true and `keysStillPresent` is
// the same set expressed as keys. When `complete` is set, anything open on the
// server but absent from that map is closed, which is how a `stale_data` alert
// ends when a vessel drops off the network — no client ever sends an explicit
// "all clear".
func (s *service) SyncEvents(events []UpsertEventRequest, keysStillPresent map[string]bool, complete bool) (SyncResult, error) {
	if err := s.validate(events); err != nil {
		return SyncResult{}, err
	}

	created, _, err := s.repo.UpsertRecords(events, keysStillPresent, complete)
	if err != nil {
		return SyncResult{}, err
	}
	return SyncResult{Created: created, Received: len(events)}, nil
}

func (s *service) ListEvents(query ListEventsQuery) (*ListEventsResponse, error) {
	if query.Severity != "" {
		switch query.Severity {
		case "info", "warning", "error":
		default:
			return nil, ErrInvalidSeverity
		}
	}

	from, err := parseDay(query.From)
	if err != nil {
		return nil, err
	}
	to, err := parseDayEnd(query.To)
	if err != nil {
		return nil, err
	}
	query.FromTime = from
	query.ToTime = to

	rows, total, openCount, err := s.repo.ListEvents(query)
	if err != nil {
		return nil, err
	}

	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}

	events := make([]EventDTO, 0, len(rows))
	for _, row := range rows {
		events = append(events, toDTO(row))
	}

	return &ListEventsResponse{
		Events:    events,
		Total:     total,
		Page:      page,
		Limit:     limit,
		OpenCount: openCount,
	}, nil
}

func (s *service) AckEvent(id uint, note string) (*EventDTO, error) {
	row, err := s.repo.AckEvent(id, time.Now())
	if err != nil {
		return nil, err
	}
	dto := toDTO(*row)
	return &dto, nil
}

func (s *service) validate(events []UpsertEventRequest) error {
	seen := make(map[string]struct{}, len(events))
	for _, e := range events {
		if e.EventKey == "" {
			return ErrMissingEventKey
		}
		if _, dup := seen[e.EventKey]; dup {
			// A duplicate within one batch would race the upsert against itself.
			return ErrDuplicateEventKey
		}
		seen[e.EventKey] = struct{}{}

		if e.DetectedAt.IsZero() {
			return ErrMissingDetectedAt
		}
		if e.DetectedAt.After(time.Now().Add(5 * time.Minute)) {
			// Guards against a client clock hours out, which would quietly
			// scramble the ordering of the entire feed.
			return ErrDetectedAtInFuture
		}
		switch e.Severity {
		case "info", "warning", "error":
		default:
			return ErrInvalidSeverity
		}
	}
	return nil
}

// parseDay anchors a YYYY-MM-DD filter at WIB midnight, matching how the fleet
// reports dates elsewhere (`/fuel` uses `base.WIB` for the same reason).
func parseDay(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, base.WIB)
	if err != nil {
		return time.Time{}, ErrInvalidDate
	}
	return t, nil
}

// parseDayEnd is `parseDay` for the exclusive upper bound, so `to=2026-05-01`
// includes the whole of that day rather than nothing at all.
func parseDayEnd(value string) (time.Time, error) {
	t, err := parseDay(value)
	if err != nil || t.IsZero() {
		return t, err
	}
	return t.Add(24 * time.Hour), nil
}

func toDTO(row model.MAlertEvent) EventDTO {
	return EventDTO{
		ID:         row.ID,
		EventKey:   row.EventKey,
		VesselID:   row.VesselID,
		VesselName: row.VesselName,
		Type:       row.Type,
		Severity:   row.Severity,
		Message:    row.Message,
		DetectedAt: row.DetectedAt,
		ResolvedAt: row.ResolvedAt,
		AckAt:      row.AckAt,
		IsOpen:     row.IsOpen(),
	}
}
