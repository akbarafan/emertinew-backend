package tracking

import (
	"testing"
	"time"
)

func TestRealtimeStateExpiry(t *testing.T) {
	s := &service{}
	rpmVal := 1200
	tempVal := 38.5
	voltVal := 220.0

	// Create state with fresh and expired sensors
	now := time.Now()
	state := &RealtimeTrackingState{
		VesselID:    1,
		VesselName:  "Test Vessel",
		RPM1:        &rpmVal,
		LastRPM1:    now.Add(-70 * time.Second), // expired (>60s)
		TopTemp:     &tempVal,
		LastTopTemp: now.Add(-10 * time.Second), // fresh (<60s)
		TopACVolt:   &voltVal,
		LastTopAC:   now.Add(-5 * time.Second),  // fresh (<60s)
	}

	dto := s.convertRealtimeStateToDTO(state)
	if dto == nil {
		t.Fatal("expected non-nil dto")
	}

	// RPM1 should be nil because it expired (>60s)
	if dto.RPM1 != nil {
		t.Errorf("expected RPM1 to be nil after expiry, got %d", *dto.RPM1)
	}

	// TopTemp should be preserved because it is fresh (<60s)
	if dto.TopTemp == nil || *dto.TopTemp != tempVal {
		t.Errorf("expected TopTemp to be preserved, got %v", dto.TopTemp)
	}

	// TopACVolt should be preserved because it is fresh (<60s)
	if dto.TopACVolt == nil || *dto.TopACVolt != voltVal {
		t.Errorf("expected TopACVolt to be preserved, got %v", dto.TopACVolt)
	}
}
