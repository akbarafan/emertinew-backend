package extractor

import (
	"math"
	"testing"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

func TestTemperatureExtractsCombinedTopAndBotPayload(t *testing.T) {
	top := &model.Tracking{}
	Temperature(top, map[string]interface{}{"temp": 33.9, "humidity": 69.1}, map[string]bool{}, true, false)

	if top.TopTemp == nil || math.Abs(*top.TopTemp-33.9) > 1e-9 {
		t.Fatalf("TopTemp = %v, want 33.9", top.TopTemp)
	}
	if top.TopHumidity == nil || math.Abs(*top.TopHumidity-69.1) > 1e-9 {
		t.Fatalf("TopHumidity = %v, want 69.1", top.TopHumidity)
	}
	if top.BotTemp != nil || top.BotHumidity != nil {
		t.Fatal("Bot fields should remain nil for top sensor")
	}

	bot := &model.Tracking{}
	Temperature(bot, map[string]interface{}{"temperature": "41.5", "hum": "72"}, map[string]bool{}, false, true)
	if bot.BotTemp == nil || math.Abs(*bot.BotTemp-41.5) > 1e-9 {
		t.Fatalf("BotTemp = %v, want 41.5", bot.BotTemp)
	}
	if bot.BotHumidity == nil || math.Abs(*bot.BotHumidity-72.0) > 1e-9 {
		t.Fatalf("BotHumidity = %v, want 72", bot.BotHumidity)
	}
}
