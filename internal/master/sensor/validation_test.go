package sensor

import (
	"encoding/json"
	"testing"
)

func TestValidateFlowConfigRequiresCalibration(t *testing.T) {
	// A flow meter without a k_factor has every reading discarded by the
	// extractor, so it would look healthy and report nothing.
	zero := 0.0
	if err := validateSensorConfig("hcml/vs001/er/flow/1", &zero, nil, nil); err == nil {
		t.Error("expected an error for k_factor = 0")
	}
	if err := validateSensorConfig("hcml/vs001/er/flow/1", nil, nil, nil); err == nil {
		t.Error("expected an error for a missing k_factor")
	}

	good := 36.34
	if err := validateSensorConfig("hcml/vs001/er/flow/1", &good, nil, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFlowConfigRejectsUnknownDirection(t *testing.T) {
	k := 36.34
	in := "in"
	out := "out"
	sideways := "sideways"

	if err := validateSensorConfig("hcml/vs001/er/flow/1", &k, nil, &in); err != nil {
		t.Errorf("unexpected error for in: %v", err)
	}
	if err := validateSensorConfig("hcml/vs001/er/flow/1", &k, nil, &out); err != nil {
		t.Errorf("unexpected error for out: %v", err)
	}
	if err := validateSensorConfig("hcml/vs001/er/flow/1", &k, nil, &sideways); err == nil {
		t.Error("expected an error for an unknown direction")
	}
}

func TestValidateSensorConfigRejectsFlowConfigOnOtherSensors(t *testing.T) {
	k := 36.34
	direction := "out"
	if err := validateSensorConfig("hcml/vs001/er/rpm/1", &k, nil, &direction); err == nil {
		t.Error("expected an error when a direction is set on a non-flow sensor")
	}
}

func TestValidateSensorConfigEngineLabelScope(t *testing.T) {
	k := 36.34
	engine := "Port Main Engine"

	// The flow meters and the tachometers belong to an engine.
	if err := validateSensorConfig("hcml/vs001/er/flow/1", &k, &engine, nil); err != nil {
		t.Errorf("unexpected error for a flow sensor: %v", err)
	}
	if err := validateSensorConfig("hcml/vs001/er/rpm/1", &k, &engine, nil); err != nil {
		t.Errorf("unexpected error for an rpm sensor: %v", err)
	}

	// The panel sensors describe the vessel, not a machine.
	if err := validateSensorConfig("hcml/vs001/er/door/1", nil, &engine, nil); err == nil {
		t.Error("expected an error when an engine label is set on a panel sensor")
	}
	if err := validateSensorConfig("hcml/vs001/bd/gps/1", nil, &engine, nil); err == nil {
		t.Error("expected an error when an engine label is set on a gps sensor")
	}
}

func TestOptionalStringDistinguishesNullFromAbsent(t *testing.T) {
	// Picking "Single" in MCU Management sends an explicit null. Without this
	// distinction the previous direction would survive in the database and the
	// operator could never clear it.
	var req UpdateSensorRequest
	if err := json.Unmarshal([]byte(`{"topic":"hcml/vs001/er/flow/1","name":"x","flow_direction":null}`), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !req.FlowDirection.Present {
		t.Error("an explicit null must count as present")
	}
	if req.FlowDirection.Value != nil {
		t.Errorf("value = %v, want nil", *req.FlowDirection.Value)
	}

	var req2 UpdateSensorRequest
	if err := json.Unmarshal([]byte(`{"topic":"hcml/vs001/er/flow/1","name":"x"}`), &req2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req2.FlowDirection.Present {
		t.Error("an absent field must not count as present")
	}
}

func TestNormaliseOptionalString(t *testing.T) {
	empty := ""
	spaced := "  Port Engine  "
	kept := "in"

	if got := normaliseOptionalString(&empty); got != nil {
		t.Errorf("empty string should normalise to nil, got %v", *got)
	}
	if got := normaliseOptionalString(&spaced); got == nil || *got != "Port Engine" {
		t.Errorf("value should be trimmed, got %v", got)
	}
	if got := normaliseOptionalString(&kept); got == nil || *got != "in" {
		t.Errorf("value should be kept, got %v", got)
	}
	if got := normaliseOptionalString(nil); got != nil {
		t.Error("nil should stay nil")
	}
}

func TestIsFlowTopic(t *testing.T) {
	cases := map[string]bool{
		"hcml/vs001/er/flow/1":                 true,
		"hcml/flt001/sqd001/vs001/brd/fuel/01": true,
		"hcml/vs001/er/rpm/1":                  false,
		"hcml/vs001/bd/gps/1":                  false,
		"nonsense":                             false,
	}
	for topic, want := range cases {
		if got := isFlowTopic(topic); got != want {
			t.Errorf("isFlowTopic(%q) = %v, want %v", topic, got, want)
		}
	}
}
