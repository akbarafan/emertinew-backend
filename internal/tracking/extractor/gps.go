package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// GPS extracts GPS data (lat, lon) from sensor value
// Format: {"lat": -7.123, "lon": 112.456}
func GPS(t *model.Tracking, v map[string]interface{}, p map[string]bool) {
	if !p["lat"] {
		if lat, ok := v["lat"].(float64); ok {
			t.Lat = &lat
			p["lat"] = true
		}
	}
	if !p["lon"] {
		if lon, ok := v["lon"].(float64); ok {
			t.Lon = &lon
			p["lon"] = true
		}
	}
}
