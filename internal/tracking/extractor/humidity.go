package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// Humidity extracts humidity data for top and bottom MCU
// Format: {"device": "top", "humidity": 65.5}
func Humidity(t *model.Tracking, v map[string]interface{}, p map[string]bool, isTop, isBot bool) {
	if humidity, ok := v["humidity"].(float64); ok {
		if isTop && !p["top_humidity"] {
			t.TopHumidity = &humidity
			p["top_humidity"] = true
		} else if isBot && !p["bot_humidity"] {
			t.BotHumidity = &humidity
			p["bot_humidity"] = true
		}
	}
}
