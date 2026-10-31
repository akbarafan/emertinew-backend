package sensor

import (
	"encoding/json"
	"errors"
	"strings"
)

// Flow direction values accepted by MCU Management.
const (
	FlowDirectionIn   = "in"
	FlowDirectionOut  = "out"
	FlowDirectionNone = ""
	engineNameMaxLen  = 100
)

// OptionalString is a string field that remembers whether it was sent at all.
//
// encoding/json cannot tell "absent" from "explicit null" for a *string: both
// end up as nil. MCU Management needs that difference — picking "Single" sends
// an explicit null, and without this type the previous direction would stay in
// the database and the operator could never clear it.
type OptionalString struct {
	// Present is true when the field was in the request body.
	Present bool
	// Value is nil when the field was explicitly null.
	Value *string
}

// UnmarshalJSON records presence and decodes the value.
func (o *OptionalString) UnmarshalJSON(data []byte) error {
	o.Present = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	o.Value = &s
	return nil
}

// normaliseOptionalString trims the value and turns an empty string into nil,
// so that "no engine" and "no direction" are stored as NULL rather than "".
func normaliseOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// isFlowTopic reports whether a topic addresses a flow (or fuel) meter.
//
// Topic format: <company>/<vessel>/<location>/<group>/<number> (5 parts) or
// <company>/<flight>/<squad>/<vessel>/<board>/<group>/<number> (7 parts).
func isFlowTopic(topic string) bool {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(topic)), "/")
	var group string
	switch len(parts) {
	case 5:
		group = parts[3]
	case 7:
		group = parts[5]
	default:
		return false
	}
	return group == "flow" || group == "fuel"
}

// validateSensorConfig guards the configuration that only makes sense on
// certain sensors.
//
//   - A flow meter without a calibration factor. `extractor.Flow` skips every
//     reading whose k_factor is 0, so the meter looks healthy and silently
//     reports nothing. Every sensor in the fleet shipped with k_factor = 0.
//   - A direction that is neither "in" nor "out". It would fall through to
//     the single-meter branch and be counted as consumption.
//   - An engine label on a sensor that has no engine. Only the flow meters and
//     the tachometers belong to an engine; the panel sensors (voltage, doors,
//     GPS) describe the vessel, not a machine.
//   - A direction on a sensor that is not a flow meter.
func validateSensorConfig(topic string, kFactor *float64, engineName *string, flowDirection *string) error {
	if isFlowTopic(topic) {
		if kFactor == nil || *kFactor <= 0 {
			return errors.New("k_factor must be greater than 0 for flow sensors: a reading is discarded when the factor is 0")
		}
		if flowDirection != nil {
			switch *flowDirection {
			case FlowDirectionIn, FlowDirectionOut:
			default:
				return errors.New("flow_direction must be \"in\", \"out\" or empty (single)")
			}
		}
	} else {
		if flowDirection != nil && *flowDirection != "" {
			return errors.New("flow_direction can only be set on a flow sensor")
		}
	}

	if engineName != nil {
		if len(*engineName) > engineNameMaxLen {
			return errors.New("engine_name is too long")
		}
		if *engineName != "" && !canBelongToEngine(topic) {
			return errors.New("engine_name can only be set on a flow or rpm sensor")
		}
	}

	return nil
}

// canBelongToEngine reports whether a sensor may carry an engine label.
//
// Topic format: <company>/<vessel>/<location>/<group>/<number> (5 parts) or
// <company>/<flight>/<squad>/<vessel>/<board>/<group>/<number> (7 parts).
func canBelongToEngine(topic string) bool {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(topic)), "/")
	if len(parts) < 5 {
		return false
	}
	group := parts[3]
	if len(parts) >= 7 {
		group = parts[5]
	}
	return group == "flow" || group == "rpm"
}
