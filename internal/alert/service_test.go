package alert

import (
	"testing"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

// stubRepo records what the service asked for, so the validation rules can be
// tested without standing up Postgres. The repository's own SQL is exercised by
// the integration path, not here.
type stubRepo struct {
	rows []model.MAlertEvent
}

func (s *stubRepo) UpsertRecords(events []UpsertEventRequest, keys map[string]bool, complete bool) (int, int, error) {
	return 0, 0, nil
}

func (s *stubRepo) ListEvents(q ListEventsQuery) ([]model.MAlertEvent, int, int, error) {
	return s.rows, len(s.rows), 0, nil
}

func (s *stubRepo) AckEvent(id uint, at time.Time) (*model.MAlertEvent, error) {
	return &model.MAlertEvent{ID: id, AckAt: &at}, nil
}

func (s *stubRepo) GetEventByID(id uint) (*model.MAlertEvent, error) {
	return &model.MAlertEvent{ID: id}, nil
}

// captureRepo remembers the filter the service handed to the repository.
type captureRepo struct {
	stubRepo
	capture         *ListEventsQuery
	captureComplete *bool
}

func (c *captureRepo) UpsertRecords(events []UpsertEventRequest, keys map[string]bool, complete bool) (int, int, error) {
	if c.captureComplete != nil {
		*c.captureComplete = complete
	}
	return c.stubRepo.UpsertRecords(events, keys, complete)
}

func (c *captureRepo) ListEvents(q ListEventsQuery) ([]model.MAlertEvent, int, int, error) {
	if c.capture != nil {
		*c.capture = q
	}
	return c.stubRepo.ListEvents(q)
}

// event builds a valid request, then applies whichever fields the case set.
// An empty field means "leave the default", which keeps each test case to one
// line.
func event(over UpsertEventRequest) UpsertEventRequest {
	base := UpsertEventRequest{
		EventKey:   "temp-alert-1",
		VesselID:   1,
		VesselName: "KM Test 01",
		Type:       "temp_alert",
		Severity:   "warning",
		Message:    "High temperature",
		DetectedAt: time.Now().Add(-5 * time.Minute),
	}
	if over.EventKey != "" {
		base.EventKey = over.EventKey
	}
	if over.VesselID != 0 {
		base.VesselID = over.VesselID
	}
	if over.VesselName != "" {
		base.VesselName = over.VesselName
	}
	if over.Type != "" {
		base.Type = over.Type
	}
	if over.Severity != "" {
		base.Severity = over.Severity
	}
	if over.Message != "" {
		base.Message = over.Message
	}
	if !over.DetectedAt.IsZero() {
		base.DetectedAt = over.DetectedAt
	}
	return base
}

func TestSyncEventsAcceptsAWellFormedBatch(t *testing.T) {
	svc := NewService(&stubRepo{})

	events := []UpsertEventRequest{
		event(UpsertEventRequest{}),
		event(UpsertEventRequest{EventKey: "door-open-1", Type: "door_open"}),
	}
	keys := map[string]bool{"temp-alert-1": true, "door-open-1": true}

	got, err := svc.SyncEvents(events, keys, true)
	if err != nil {
		t.Fatalf("SyncEvents: %v", err)
	}
	if got.Received != 2 {
		t.Errorf("Received = %d, want 2", got.Received)
	}
}

func TestSyncEventsRejectsBadInput(t *testing.T) {
	cases := []struct {
		name   string
		events []UpsertEventRequest
		want   error
	}{
		{
			// A duplicate key inside one batch would race the upsert against
			// itself: two CREATEs for the same unique event_key.
			name: "duplicate key in one batch",
			events: []UpsertEventRequest{
				event(UpsertEventRequest{}),
				event(UpsertEventRequest{EventKey: "temp-alert-1"}),
			},
			want: ErrDuplicateEventKey,
		},
		{
			// A browser clock hours out would scramble the ordering of the whole
			// feed, so it is rejected rather than stored.
			name:   "detected_at far in the future",
			events: []UpsertEventRequest{event(UpsertEventRequest{DetectedAt: time.Now().Add(48 * time.Hour)})},
			want:   ErrDetectedAtInFuture,
		},
		{
			name:   "unknown severity",
			events: []UpsertEventRequest{event(UpsertEventRequest{Severity: "fatal"})},
			want:   ErrInvalidSeverity,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(&stubRepo{})
			_, err := svc.SyncEvents(tc.events, nil, true)
			if err != tc.want {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestSyncEventsRejectsMissingRequiredFields(t *testing.T) {
	// Covered separately because the shared `event` helper treats a zero value
	// as "use the default", so it cannot express these two cases.
	svc := NewService(&stubRepo{})

	valid := func() UpsertEventRequest {
		return event(UpsertEventRequest{})
	}

	noKey := valid()
	noKey.EventKey = ""
	if _, err := svc.SyncEvents([]UpsertEventRequest{noKey}, nil, true); err != ErrMissingEventKey {
		t.Errorf("missing key: error = %v, want %v", err, ErrMissingEventKey)
	}

	noTime := valid()
	noTime.DetectedAt = time.Time{}
	if _, err := svc.SyncEvents([]UpsertEventRequest{noTime}, nil, true); err != ErrMissingDetectedAt {
		t.Errorf("zero detected_at: error = %v, want %v", err, ErrMissingDetectedAt)
	}
}

func TestSyncEventsPassesCompleteThrough(t *testing.T) {
	// The resolve-everything-omitted behaviour is gated on `complete`, so the
	// service has to forward it untouched.
	var captured bool
	repo := &captureRepo{captureComplete: &captured}
	svc := NewService(repo)

	if _, err := svc.SyncEvents(nil, map[string]bool{}, true); err != nil {
		t.Fatalf("SyncEvents: %v", err)
	}
	if !captured {
		t.Error("complete = true was not forwarded to the repository")
	}

	captured = true
	if _, err := svc.SyncEvents(nil, nil, false); err != nil {
		t.Fatalf("SyncEvents: %v", err)
	}
	if captured {
		t.Error("complete = false was not forwarded to the repository")
	}
}

func TestListEventsAnchorsDatesToWIB(t *testing.T) {
	var captured ListEventsQuery
	svc := NewService(&captureRepo{capture: &captured})

	if _, err := svc.ListEvents(ListEventsQuery{From: "2026-04-30", To: "2026-05-01"}); err != nil {
		t.Fatalf("ListEvents: %v", err)
	}

	if captured.FromTime.IsZero() || captured.ToTime.IsZero() {
		t.Fatal("date filters were not parsed into FromTime/ToTime")
	}
	// `to` is inclusive of the whole day, so the exclusive upper bound is the
	// following midnight. Getting this wrong silently returns an empty day.
	if got := captured.ToTime.Sub(captured.FromTime); got != 48*time.Hour {
		t.Errorf("date span = %v, want 48h", got)
	}
}

func TestListEventsRejectsBadFilters(t *testing.T) {
	svc := NewService(&stubRepo{})

	if _, err := svc.ListEvents(ListEventsQuery{Severity: "nope"}); err != ErrInvalidSeverity {
		t.Errorf("severity error = %v, want %v", err, ErrInvalidSeverity)
	}
	if _, err := svc.ListEvents(ListEventsQuery{From: "30-04-2026"}); err != ErrInvalidDate {
		t.Errorf("from error = %v, want %v", err, ErrInvalidDate)
	}
	if _, err := svc.ListEvents(ListEventsQuery{To: "tomorrow"}); err != ErrInvalidDate {
		t.Errorf("to error = %v, want %v", err, ErrInvalidDate)
	}
}

func TestListEventsAppliesPaginationDefaults(t *testing.T) {
	svc := NewService(&stubRepo{})

	got, err := svc.ListEvents(ListEventsQuery{})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if got.Page != 1 {
		t.Errorf("Page = %d, want 1", got.Page)
	}
	if got.Limit != 50 {
		t.Errorf("Limit = %d, want 50", got.Limit)
	}
}

func TestToDTOCarriesOpenState(t *testing.T) {
	now := time.Now()

	open := toDTO(model.MAlertEvent{
		ID: 1, EventKey: "temp-alert-1", VesselID: 1, VesselName: "KM A",
		Type: "temp_alert", Severity: "error", Message: "hot", DetectedAt: now,
	})
	if !open.IsOpen {
		t.Error("event with no resolved_at should be open")
	}

	resolvedAt := now.Add(time.Minute)
	resolved := toDTO(model.MAlertEvent{
		ID: 2, EventKey: "door-open-1", VesselID: 2, VesselName: "KM B",
		Type: "door_open", Severity: "warning", Message: "open",
		DetectedAt: now, ResolvedAt: &resolvedAt,
	})
	if resolved.IsOpen {
		t.Error("event with resolved_at should not be open")
	}
}
