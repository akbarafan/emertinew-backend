package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// Temperature extracts temperature data for top and bottom MCU
// Format: {"device": "top", "temp": 25.5}
func Temperature(t *model.Tracking, v map[string]interface{}, p map[string]bool, isTop, isBot bool) {
	if temp, ok := v["temp"].(float64); ok {
		if isTop && !p["top_temp"] {
			t.TopTemp = &temp
			p["top_temp"] = true
		} else if isBot && !p["bot_temp"] {
			t.BotTemp = &temp
			p["bot_temp"] = true
		}
	}
}
