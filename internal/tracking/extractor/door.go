package extractor

import "github.com/Alugara-Inovasi-Utama/emerti/model"

// Door extracts door status (magnetic sensor)
// Location (bd/er) determines Top/Bot, sensorNumber determines which door (1 or 2)
//
// Message format:
//
//	{
//	  "value": 1,
//	  "timestamp": "2026-02-03 15:57:17",
//	  "sensor_code": "MH23"
//	}
//
// Topics: hcml/vs001/bd/door/1, hcml/vs001/bd/door/2
//
//	hcml/vs001/er/door/1, hcml/vs001/er/door/2
//
// Mapping:
//   - bd + door/1 → TopDoor1
//   - bd + door/2 → TopDoor2
//   - er + door/1 → BotDoor1
//   - er + door/2 → BotDoor2
func Door(t *model.Tracking, v map[string]interface{}, p map[string]bool, isTop, isBot bool, sensorNumber int) {
	var boolVal bool
	var found bool

	// Check "value" field (number: 0/1)
	if val, ok := v["value"].(float64); ok {
		boolVal = val != 0 // 0 = closed (false), 1 = open (true)
		found = true
	} else if val, ok := v["value"].(bool); ok {
		boolVal = val
		found = true
	}

	if found {
		// Map to appropriate door field based on location and sensor number
		if isTop {
			if sensorNumber == 1 && !p["top_door1"] {
				t.TopDoor1 = &boolVal
				p["top_door1"] = true
			} else if sensorNumber == 2 && !p["top_door2"] {
				t.TopDoor2 = &boolVal
				p["top_door2"] = true
			}
		} else if isBot {
			if sensorNumber == 1 && !p["bot_door1"] {
				t.BotDoor1 = &boolVal
				p["bot_door1"] = true
			} else if sensorNumber == 2 && !p["bot_door2"] {
				t.BotDoor2 = &boolVal
				p["bot_door2"] = true
			}
		}
	}
}
