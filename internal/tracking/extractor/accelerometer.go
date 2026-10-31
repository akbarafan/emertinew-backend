package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// Accelerometer extracts accelerometer data (x, y, z)
// Format: {"x": 0.1, "y": 0.2, "z": 9.8}
func Accelerometer(t *model.Tracking, v map[string]interface{}, p map[string]bool) {
	if x, ok := v["x"].(float64); ok && !p["accel_x"] {
		t.AccelX = &x
		p["accel_x"] = true
	}
	if y, ok := v["y"].(float64); ok && !p["accel_y"] {
		t.AccelY = &y
		p["accel_y"] = true
	}
	if z, ok := v["z"].(float64); ok && !p["accel_z"] {
		t.AccelZ = &z
		p["accel_z"] = true
	}
}
