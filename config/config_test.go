package config

import "testing"

func TestActiveMQTTBrokers_PrimaryOnly(t *testing.T) {
	cfg := &Config{
		MQTTBroker:  "103.150.194.254",
		MQTTPort:    "1883",
		MQTT2Broker: "",
		MQTT2Port:   "1883",
	}

	got := cfg.ActiveMQTTBrokers()
	if len(got) != 1 {
		t.Fatalf("expected 1 active broker, got %d (%+v)", len(got), got)
	}
	if got[0].Label != "PRIMARY" || got[0].Host != "103.150.194.254" || got[0].Port != "1883" {
		t.Fatalf("unexpected primary broker: %+v", got[0])
	}
}

func TestActiveMQTTBrokers_SecondaryOnly(t *testing.T) {
	cfg := &Config{
		MQTTBroker:  "   ",
		MQTTPort:    "1883",
		MQTT2Broker: "103.150.194.254",
		MQTT2Port:   "",
	}

	got := cfg.ActiveMQTTBrokers()
	if len(got) != 1 {
		t.Fatalf("expected 1 active broker when primary is blank, got %d (%+v)", len(got), got)
	}
	if got[0].Label != "SECONDARY" || got[0].Host != "103.150.194.254" || got[0].Port != "1883" {
		t.Fatalf("unexpected secondary broker: %+v", got[0])
	}
}

func TestActiveMQTTBrokers_BothOrEmptyOrDuplicate(t *testing.T) {
	both := (&Config{
		MQTTBroker:  "103.103.20.19",
		MQTTPort:    "1883",
		MQTT2Broker: "103.150.194.254",
		MQTT2Port:   "1883",
	}).ActiveMQTTBrokers()
	if len(both) != 2 {
		t.Fatalf("expected 2 brokers when both configured, got %d", len(both))
	}

	dup := (&Config{
		MQTTBroker:  "103.150.194.254",
		MQTTPort:    "1883",
		MQTT2Broker: "103.150.194.254",
		MQTT2Port:   "1883",
	}).ActiveMQTTBrokers()
	if len(dup) != 1 {
		t.Fatalf("expected duplicate endpoint to be deduplicated to 1, got %d", len(dup))
	}

	none := (&Config{
		MQTTBroker:  "",
		MQTT2Broker: "",
	}).ActiveMQTTBrokers()
	if len(none) != 0 {
		t.Fatalf("expected 0 brokers when both empty, got %d", len(none))
	}
}
