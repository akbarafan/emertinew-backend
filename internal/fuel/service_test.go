package fuel

import (
	"math"
	"testing"
	"time"
)

// stubRepo returns fixed directions and aggregates, so the consumption rule can
// be tested without a database.
type stubRepo struct {
	directions []FlowSensorDirection
	results    []FuelAggregateResult
}

func (s *stubRepo) GetFuelTotalByDate(date time.Time) ([]FuelAggregateResult, error) {
	return s.results, nil
}

func (s *stubRepo) GetFuelTotalByRange(start, end time.Time) ([]FuelAggregateResult, error) {
	return s.results, nil
}

func (s *stubRepo) GetFuelTotalByVesselAndDate(vesselID uint, date time.Time) (*FuelAggregateResult, error) {
	if len(s.results) == 0 {
		return &FuelAggregateResult{VesselID: vesselID}, nil
	}
	return &s.results[0], nil
}

func (s *stubRepo) GetLifetimeFuelByVessel(vesselID uint) (*FuelAggregateResult, error) {
	if len(s.results) == 0 {
		return &FuelAggregateResult{VesselID: vesselID}, nil
	}
	return &s.results[0], nil
}

func (s *stubRepo) GetFlowSensorDirections() ([]FlowSensorDirection, error) {
	return s.directions, nil
}

func TestBuildEngineBreakdownConsumptionRule(t *testing.T) {
	// The rule under test, from the engine's point of view (marine standard):
	//   in  300 L/h  ->  fuel supplied into the engine
	//   out  50 L/h  ->  fuel returning to the tank
	//   consumed     ->  300 - 50 = 250 L/h
	cases := []struct {
		name       string
		directions []FlowSensorDirection
		flows      map[string]float64
		wantByName map[string]float64
		wantTotal  float64
	}{
		{
			name: "directed pair is the difference",
			directions: []FlowSensorDirection{
				{VesselID: 1, Slot: "flow1", Direction: DirectionIn, Engine: "Port"},
				{VesselID: 1, Slot: "flow2", Direction: DirectionOut, Engine: "Port"},
			},
			flows:      map[string]float64{"flow1": 300, "flow2": 50},
			wantByName: map[string]float64{"Port": 250},
			wantTotal:  250,
		},
		{
			name: "in only means the meter is on the supply side",
			directions: []FlowSensorDirection{
				{VesselID: 1, Slot: "flow1", Direction: DirectionIn, Engine: "Port"},
			},
			flows:      map[string]float64{"flow1": 300},
			wantByName: map[string]float64{"Port": 300},
			wantTotal:  300,
		},
		{
			name: "single meter is the consumption",
			directions: []FlowSensorDirection{
				{VesselID: 1, Slot: "flow1", Direction: DirectionSingle, Engine: "Port"},
			},
			flows:      map[string]float64{"flow1": 120},
			wantByName: map[string]float64{"Port": 120},
			wantTotal:  120,
		},
		{
			name: "out only cannot measure consumption",
			directions: []FlowSensorDirection{
				{VesselID: 1, Slot: "flow1", Direction: DirectionOut, Engine: "Port"},
			},
			flows:      map[string]float64{"flow1": 300},
			wantByName: map[string]float64{"Port": NotMeasured},
			wantTotal:  NotMeasured,
		},
		{
			name: "two engines are kept apart",
			directions: []FlowSensorDirection{
				{VesselID: 1, Slot: "flow1", Direction: DirectionIn, Engine: "Port"},
				{VesselID: 1, Slot: "flow2", Direction: DirectionOut, Engine: "Port"},
				{VesselID: 1, Slot: "flow3", Direction: DirectionIn, Engine: "Starboard"},
				{VesselID: 1, Slot: "flow4", Direction: DirectionOut, Engine: "Starboard"},
			},
			flows: map[string]float64{"flow1": 70.7, "flow2": 5.7, "flow3": 67.4, "flow4": 5.4},
			wantByName: map[string]float64{
				"Port":      65.0,
				"Starboard": 62.0,
			},
			wantTotal: 127.0,
		},
		{
			name: "returned more than drawn never goes negative",
			directions: []FlowSensorDirection{
				{VesselID: 1, Slot: "flow1", Direction: DirectionIn, Engine: "Port"},
				{VesselID: 1, Slot: "flow2", Direction: DirectionOut, Engine: "Port"},
			},
			flows:      map[string]float64{"flow1": 30, "flow2": 80},
			wantByName: map[string]float64{"Port": 0},
			wantTotal:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &service{repo: &stubRepo{directions: tc.directions}}
			total, returned, stockChange, engines := svc.buildEngineBreakdown(1, tc.flows, tc.directions)

			if total != tc.wantTotal {
				t.Errorf("consumed = %v, want %v", total, tc.wantTotal)
			}
			if tc.wantTotal != NotMeasured {
				if stockChange != returned-total {
					t.Errorf("stock change = %v, want %v", stockChange, returned-total)
				}
			}
			for name, want := range tc.wantByName {
				found := false
				for _, eng := range engines {
					if eng.EngineName != name {
						continue
					}
					found = true
					if math.Abs(eng.Consumed-want) > 1e-6 {
						t.Errorf("%s consumed = %v, want %v", name, eng.Consumed, want)
					}
					// The deprecated alias must not drift from the real figure.
					if eng.NetTotal != eng.Consumed {
						t.Errorf("%s net_total = %v, want %v", name, eng.NetTotal, eng.Consumed)
					}
				}
				if !found {
					t.Errorf("engine %q missing from the breakdown", name)
				}
			}
		})
	}
}

func TestBuildEngineBreakdownUnassignedBucket(t *testing.T) {
	// Channels without an engine label are still metered, so they belong in the
	// vessel total even though they cannot be attributed to an engine.
	directions := []FlowSensorDirection{
		{VesselID: 1, Slot: "flow1", Direction: DirectionSingle},
		{VesselID: 1, Slot: "flow2", Direction: DirectionSingle, Engine: "Port"},
	}
	svc := &service{repo: &stubRepo{}}
	total, _, _, engines := svc.buildEngineBreakdown(1, map[string]float64{"flow1": 40, "flow2": 60}, directions)

	if total != 100 {
		t.Errorf("consumed = %v, want 100", total)
	}
	if len(engines) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(engines))
	}
	if engines[0].EngineName != UnassignedEngineName && engines[1].EngineName != UnassignedEngineName {
		t.Error("the unlabelled channel should land in the Unassigned bucket")
	}
}

func TestGetFuelTotalByVesselFillsConsumption(t *testing.T) {
	repo := &stubRepo{
		directions: []FlowSensorDirection{
			{VesselID: 1, Slot: "flow1", Direction: DirectionIn, Engine: "Port"},
			{VesselID: 1, Slot: "flow2", Direction: DirectionOut, Engine: "Port"},
		},
		results: []FuelAggregateResult{{
			VesselID:   1,
			VesselName: "AHT. LIMIN",
			Flow1Total: 300,
			Flow2Total: 50,
		}},
	}
	svc := NewService(repo)

	dto, err := svc.GetFuelTotalByVessel(1, time.Now())
	if err != nil {
		t.Fatalf("GetFuelTotalByVessel: %v", err)
	}
	if dto.Consumed != 250 {
		t.Errorf("consumed = %v, want 250", dto.Consumed)
	}
	if dto.FuelReturned != 50 {
		t.Errorf("fuel returned = %v, want 50", dto.FuelReturned)
	}
	if dto.StockChange != -200 {
		t.Errorf("stock change = %v, want -200", dto.StockChange)
	}
	if dto.FuelTotal != 350 {
		t.Errorf("metered total = %v, want 350", dto.FuelTotal)
	}
}
