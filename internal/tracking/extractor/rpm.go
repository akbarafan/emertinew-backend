package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// RPM extracts RPM data from sensor value
// Each RPM sensor has its own topic: hcml/vs001/er/rpm/1, /rpm/2, /rpm/3, /rpm/4
// Format: {"rpm": 1389, "timestamp": "2025-12-11 22:42:36"}
// sensorNumber determines which RPM slot to fill (1-4)
func RPM(t *model.Tracking, v map[string]interface{}, p map[string]bool, sensorNumber int) {
	// Get RPM value from payload
	var rpm int
	var hasRPM bool

	if val, ok := v["rpm"].(float64); ok {
		rpm = int(val)
		hasRPM = true
	} else if val, ok := v["rpm"].(int); ok {
		rpm = val
		hasRPM = true
	}

	if !hasRPM {
		return
	}

	// Assign to correct RPM slot based on sensor number from topic
	switch sensorNumber {
	case 1:
		if !p["rpm1"] {
			t.RPM1 = &rpm
			p["rpm1"] = true
		}
	case 2:
		if !p["rpm2"] {
			t.RPM2 = &rpm
			p["rpm2"] = true
		}
	case 3:
		if !p["rpm3"] {
			t.RPM3 = &rpm
			p["rpm3"] = true
		}
	case 4:
		if !p["rpm4"] {
			t.RPM4 = &rpm
			p["rpm4"] = true
		}
	}
}
