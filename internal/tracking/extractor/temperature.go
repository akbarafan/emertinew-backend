package extractor

import (
	"strconv"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

// parseNumericField extracts a float64 from the first matching key in v,
// accepting JSON numbers (float64) as well as numeric strings.
func parseNumericField(v map[string]interface{}, keys ...string) (float64, bool) {
	for _, k := range keys {
		raw, exists := v[k]
		if !exists || raw == nil {
			continue
		}
		switch val := raw.(type) {
		case float64:
			return val, true
		case string:
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

// Temperature extracts temperature and humidity data for top and bottom MCU
// Supports combined sensor format: {"timestamp": "2026-11-09 16:53:35", "sensor_code": "TMP", "temp": 33.9, "humidity": 69.1}
func Temperature(t *model.Tracking, v map[string]interface{}, p map[string]bool, isTop, isBot bool) {
	if temp, ok := parseNumericField(v, "temp", "temperature"); ok {
		if isTop && !p["top_temp"] {
			t.TopTemp = &temp
			p["top_temp"] = true
		} else if isBot && !p["bot_temp"] {
			t.BotTemp = &temp
			p["bot_temp"] = true
		}
	}
	// Combined payload: also extract humidity if published in the same temp packet
	Humidity(t, v, p, isTop, isBot)
}
