package extractor

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

// AC extracts AC electrical data (voltage, current, frequency, power) for top and bottom MCU
// Topic: hcml/vs001/bd/ac/1 or hcml/vs001/er/ac/1
// Format: {"voltage": 239.2, "current": 3080.192, "frequency": 50.5, "power": 288358.4, "timestamp": "..."}
func AC(t *model.Tracking, v map[string]interface{}, p map[string]bool, isTop, isBot bool) {
	// Extract voltage
	if volt, ok := v["voltage"].(float64); ok {
		if isTop && !p["top_ac_volt"] {
			t.TopACVolt = &volt
			p["top_ac_volt"] = true
		} else if isBot && !p["bot_ac_volt"] {
			t.BotACVolt = &volt
			p["bot_ac_volt"] = true
		}
	}

	// Extract current (amp)
	if amp, ok := v["current"].(float64); ok {
		if isTop && !p["top_ac_amp"] {
			t.TopACAmp = &amp
			p["top_ac_amp"] = true
		} else if isBot && !p["bot_ac_amp"] {
			t.BotACAmp = &amp
			p["bot_ac_amp"] = true
		}
	}

	// Extract frequency
	if freq, ok := v["frequency"].(float64); ok {
		if isTop && !p["top_ac_freq"] {
			t.TopACFreq = &freq
			p["top_ac_freq"] = true
		} else if isBot && !p["bot_ac_freq"] {
			t.BotACFreq = &freq
			p["bot_ac_freq"] = true
		}
	}

	// Extract power - handle both float64 and int
	if powerVal, exists := v["power"]; exists {
		var power float64
		var validType bool

		switch val := powerVal.(type) {
		case float64:
			power = val
			validType = true
		case int:
			power = float64(val)
			validType = true
		case int64:
			power = float64(val)
			validType = true
		default:
			log.Printf("[AC] Warning: power field has unexpected type %T with value %v", powerVal, powerVal)
		}

		if validType {
			if isTop && !p["top_ac_power"] {
				t.TopACPower = &power
				p["top_ac_power"] = true
				log.Printf("[AC] Set top_ac_power = %.2f", power)
			} else if isBot && !p["bot_ac_power"] {
				t.BotACPower = &power
				p["bot_ac_power"] = true
				log.Printf("[AC] Set bot_ac_power = %.2f", power)
			}
		}
	} else {
		log.Printf("[AC] Warning: power field not found in sensor data")
	}
}
