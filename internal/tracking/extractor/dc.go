package extractor

import (
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

// DC extracts DC electrical data (voltage, current) for each DC sensor
// Location (bd/er) determines Top/Bot fields
// Source field determines DC_Kapal vs DC_Baterai
//
// Message format:
//
//	{
//	  "source": "DC_Kapal",
//	  "voltage": 12.695,
//	  "current": 3.276,
//	  "timestamp": "2026-01-28 14:28:16"
//	}
//
// Topics: hcml/vs001/bd/dc_kapal/1, hcml/vs001/bd/dc_baterai/1
//
//	hcml/vs001/er/dc_kapal/1, hcml/vs001/er/dc_baterai/1
//
// Mapping:
//   - source: "DC_Kapal" + bd → TopDCVolt/TopDCAmp
//   - source: "DC_Kapal" + er → BotDCVolt/BotDCAmp
//   - source: "DC_Baterai" + bd → TopBatVolt/TopBatAmp
//   - source: "DC_Baterai" + er → BotBatVolt/BotBatAmp
func DC(t *model.Tracking, v map[string]interface{}, p map[string]bool, isTop, isBot bool) {
	// Get source to determine DC_Kapal or DC_Baterai
	source, _ := v["source"].(string)
	sourceLower := strings.ToLower(source)

	isDCKapal := strings.Contains(sourceLower, "kapal") || strings.Contains(sourceLower, "dc_kapal")
	isDCBaterai := strings.Contains(sourceLower, "baterai") || strings.Contains(sourceLower, "dc_baterai")

	// Extract voltage and current
	voltage, hasVoltage := v["voltage"].(float64)
	current, hasCurrent := v["current"].(float64)

	if isDCKapal {
		// DC Kapal → dc_volt, dc_amp
		if hasVoltage {
			if isTop && !p["top_dc_volt"] {
				t.TopDCVolt = &voltage
				p["top_dc_volt"] = true
			} else if isBot && !p["bot_dc_volt"] {
				t.BotDCVolt = &voltage
				p["bot_dc_volt"] = true
			}
		}
		if hasCurrent {
			if isTop && !p["top_dc_amp"] {
				t.TopDCAmp = &current
				p["top_dc_amp"] = true
			} else if isBot && !p["bot_dc_amp"] {
				t.BotDCAmp = &current
				p["bot_dc_amp"] = true
			}
		}
	} else if isDCBaterai {
		// DC Baterai → bat_volt, bat_amp
		if hasVoltage {
			if isTop && !p["top_bat_volt"] {
				t.TopBatVolt = &voltage
				p["top_bat_volt"] = true
			} else if isBot && !p["bot_bat_volt"] {
				t.BotBatVolt = &voltage
				p["bot_bat_volt"] = true
			}
		}
		if hasCurrent {
			if isTop && !p["top_bat_amp"] {
				t.TopBatAmp = &current
				p["top_bat_amp"] = true
			} else if isBot && !p["bot_bat_amp"] {
				t.BotBatAmp = &current
				p["bot_bat_amp"] = true
			}
		}
	}
}
