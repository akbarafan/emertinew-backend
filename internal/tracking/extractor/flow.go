package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// Flow extracts flow data from sensor value
// Each Flow sensor has its own topic: hcml/vs001/er/flow/1, /flow/2, /flow/3, /flow/4
// Format: {"timestamp": "2026-04-18 14:30:00", "pulse": 1000}
// sensorNumber determines which Flow slot to fill (1-4)
// kFactor is the calibration factor in pulse/L. Volume(L) = pulse / k_factor.
// kFactor must be > 0 (sensor must be calibrated), otherwise the reading is skipped.
//
// Unit contract: the value produced here is the volume of ONE SAMPLE, i.e. the
// litres that passed the meter during the hardware's 5-second window. It is not
// an hourly rate and not a per-minute volume.
//
// Two consumers exist and they need different things:
//
//   - Flow is used by the realtime broadcast path, which wants the latest
//     sample so the UI can show a current rate.
//   - FlowAccumulate is used by the scheduler, which builds one `trackings` row
//     per minute and therefore has to add up every sample that arrived during
//     that minute. What lands in `trackings.flow1..4` is litres per minute.
func Flow(t *model.Tracking, v map[string]interface{}, p map[string]bool, sensorNumber int, kFactor float64) {
	pulse, ok := v["pulse"].(float64)
	if !ok {
		return
	}

	if kFactor <= 0 {
		// Sensor not calibrated, skip
		return
	}

	flowValue := pulse / kFactor

	// Assign to correct Flow slot based on sensor number from topic
	switch sensorNumber {
	case 1:
		if !p["flow1"] {
			t.Flow1 = &flowValue
			p["flow1"] = true
		}
	case 2:
		if !p["flow2"] {
			t.Flow2 = &flowValue
			p["flow2"] = true
		}
	case 3:
		if !p["flow3"] {
			t.Flow3 = &flowValue
			p["flow3"] = true
		}
	case 4:
		if !p["flow4"] {
			t.Flow4 = &flowValue
			p["flow4"] = true
		}
	}
}

// FlowAccumulate adds one flow sample to the running total for its slot.
//
// The scheduler calls this once per sample found for the minute, so the stored
// value is the exact number of litres that passed the meter during that
// minute — no scaling factor and no averaging assumption involved.
//
// The `p` map is deliberately not used as a guard: every sample has to be
// added. A sensor that has already been seen in this minute still contributes
// again, because the hardware publishes a fresh sample every 5 seconds.
func FlowAccumulate(t *model.Tracking, v map[string]interface{}, p map[string]bool, sensorNumber int, kFactor float64) {
	pulse, ok := v["pulse"].(float64)
	if !ok {
		return
	}

	if kFactor <= 0 {
		// Sensor not calibrated, skip
		return
	}

	litres := pulse / kFactor

	switch sensorNumber {
	case 1:
		if t.Flow1 == nil {
			t.Flow1 = &litres
		} else {
			*t.Flow1 += litres
		}
	case 2:
		if t.Flow2 == nil {
			t.Flow2 = &litres
		} else {
			*t.Flow2 += litres
		}
	case 3:
		if t.Flow3 == nil {
			t.Flow3 = &litres
		} else {
			*t.Flow3 += litres
		}
	case 4:
		if t.Flow4 == nil {
			t.Flow4 = &litres
		} else {
			*t.Flow4 += litres
		}
	}
}
