package extractor

import (
	"math"
	"testing"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

func TestFlowKeepsTheFirstSampleOfABatch(t *testing.T) {
	// Within one batch the first sample wins. The scheduler walks the samples
	// of a minute newest first, so "first" is the latest reading of that
	// minute.
	tr := &model.Tracking{}
	p := map[string]bool{}

	Flow(tr, map[string]interface{}{"pulse": 300.0}, p, 1, 200)
	Flow(tr, map[string]interface{}{"pulse": 100.0}, p, 1, 200)

	if tr.Flow1 == nil {
		t.Fatal("flow1 was not set")
	}
	if math.Abs(*tr.Flow1-1.5) > 1e-9 {
		t.Errorf("flow1 = %v, want 1.5 (the first sample of the batch)", *tr.Flow1)
	}
}

func TestFlowOverwritesWhenEachMessageHasItsOwnMap(t *testing.T) {
	// The realtime path builds a fresh map per MQTT message, so the newest
	// message replaces the previous reading.
	tr := &model.Tracking{}

	Flow(tr, map[string]interface{}{"pulse": 100.0}, map[string]bool{}, 1, 200)
	Flow(tr, map[string]interface{}{"pulse": 300.0}, map[string]bool{}, 1, 200)

	if tr.Flow1 == nil {
		t.Fatal("flow1 was not set")
	}
	if math.Abs(*tr.Flow1-1.5) > 1e-9 {
		t.Errorf("flow1 = %v, want 1.5 (the newest message)", *tr.Flow1)
	}
}

func TestFlowAccumulateSumsTheMinute(t *testing.T) {
	// The scheduler builds one row per minute and has to add up every sample
	// the hardware published during it.
	tr := &model.Tracking{}
	p := map[string]bool{}

	// Twelve samples of one minute, 100 pulses each at 200 pulse/L.
	for i := 0; i < 12; i++ {
		FlowAccumulate(tr, map[string]interface{}{"pulse": 100.0}, p, 1, 200)
	}

	if tr.Flow1 == nil {
		t.Fatal("flow1 was not set")
	}
	// 12 * (100 / 200) = 6 litres in that minute.
	if math.Abs(*tr.Flow1-6.0) > 1e-9 {
		t.Errorf("flow1 = %v, want 6 litres per minute", *tr.Flow1)
	}
}

func TestFlowAccumulateKeepsChannelsApart(t *testing.T) {
	tr := &model.Tracking{}
	p := map[string]bool{}

	FlowAccumulate(tr, map[string]interface{}{"pulse": 100.0}, p, 1, 200)  // 0.5 L
	FlowAccumulate(tr, map[string]interface{}{"pulse": 57.0}, p, 2, 33.79) // 1.687 L

	if tr.Flow1 == nil || math.Abs(*tr.Flow1-0.5) > 1e-9 {
		t.Errorf("flow1 = %v, want 0.5", tr.Flow1)
	}
	if tr.Flow2 == nil || math.Abs(*tr.Flow2-1.687) > 1e-3 {
		t.Errorf("flow2 = %v, want ~1.687", tr.Flow2)
	}
	if tr.Flow3 != nil || tr.Flow4 != nil {
		t.Error("untouched channels must stay nil")
	}
}

func TestFlowSkipsUncalibratedSensors(t *testing.T) {
	// A sensor with k_factor 0 is not calibrated. The reading is discarded
	// rather than reported as zero, so a broken meter cannot look like a dry
	// tank.
	for _, kFactor := range []float64{0, -1} {
		tr := &model.Tracking{}
		p := map[string]bool{}
		Flow(tr, map[string]interface{}{"pulse": 100.0}, p, 1, kFactor)
		if tr.Flow1 != nil {
			t.Errorf("k_factor %v: flow1 = %v, want nil", kFactor, *tr.Flow1)
		}

		tr2 := &model.Tracking{}
		p2 := map[string]bool{}
		FlowAccumulate(tr2, map[string]interface{}{"pulse": 100.0}, p2, 1, kFactor)
		if tr2.Flow1 != nil {
			t.Errorf("k_factor %v: accumulated flow1 = %v, want nil", kFactor, *tr2.Flow1)
		}
	}
}

func TestFlowIgnoresPayloadsWithoutPulse(t *testing.T) {
	tr := &model.Tracking{}
	p := map[string]bool{}
	Flow(tr, map[string]interface{}{"liter": 5.0}, p, 1, 200)
	if tr.Flow1 != nil {
		t.Errorf("flow1 = %v, want nil", *tr.Flow1)
	}
}
