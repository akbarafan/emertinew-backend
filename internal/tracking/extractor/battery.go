package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// Battery extracts battery data (volt, amp) for top and bottom MCU
// Note: Battery data from DC channels format is already handled by DC extractor (channel 2)
// This function handles standalone battery sensor format
// Format: {"device": "top", "bat_volt": 12.5, "bat_amp": 2.3}
func Battery(t *model.Tracking, v map[string]interface{}, p map[string]bool, isTop, isBot bool) {
	// Skip if channels format (handled by DC extractor)
	if _, hasChannels := v["channels"]; hasChannels {
		return
	}

	if volt, ok := v["bat_volt"].(float64); ok {
		if isTop && !p["top_bat_volt"] {
			t.TopBatVolt = &volt
			p["top_bat_volt"] = true
		} else if isBot && !p["bot_bat_volt"] {
			t.BotBatVolt = &volt
			p["bot_bat_volt"] = true
		}
	}
	if amp, ok := v["bat_amp"].(float64); ok {
		if isTop && !p["top_bat_amp"] {
			t.TopBatAmp = &amp
			p["top_bat_amp"] = true
		} else if isBot && !p["bot_bat_amp"] {
			t.BotBatAmp = &amp
			p["bot_bat_amp"] = true
		}
	}
}
