package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	AppPort       string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisPrefix   string
	MQTTBroker    string
	MQTTPort      string
	MQTT2Broker   string
	MQTT2Port     string

	// Telegram Customer Support
	TelegramBotToken      string
	TelegramSupportChatID string
	TelegramWebhookURL    string
	AppBaseURL            string
	WebPortalURL          string
}

func LoadConfig() *Config {
	// Load base .env first
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	// Load optional .env.local to override values locally
	if err := godotenv.Overload(".env.local"); err == nil {
		log.Println("Loaded .env.local overrides")
	}

	return &Config{
		// .env name, "Default if in .env ga ada"
		AppEnv:        getEnv("APP_ENV", "development"),
		AppPort:       getEnv("APP_PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "password"),
		DBName:        getEnv("DB_NAME", "dbname"),
		JWTSecret:     getEnv("JWT_SECRET", "secret"),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,
		RedisPrefix:   getEnv("REDIS_PREFIX", "emerti:session:"),
		MQTTBroker:    strings.TrimSpace(getEnv("MQTT_BROKER", "")),
		MQTTPort:      strings.TrimSpace(getEnv("MQTT_PORT", "1883")),
		MQTT2Broker:   strings.TrimSpace(getEnv("MQTT2_BROKER", "")),
		MQTT2Port:     strings.TrimSpace(getEnv("MQTT2_PORT", "1883")),

		// Telegram Bot configurations
		TelegramBotToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramSupportChatID: getEnv("TELEGRAM_SUPPORT_CHAT_ID", ""),
		TelegramWebhookURL:    getEnv("TELEGRAM_WEBHOOK_URL", ""),
		AppBaseURL:            getEnv("APP_BASE_URL", "http://localhost:6969"),
		WebPortalURL:          getEnv("WEB_PORTAL_URL", "http://localhost:3000"),
	}
}

// MQTTBrokerEndpoint describes an active MQTT broker connection.
type MQTTBrokerEndpoint struct {
	Label string
	Host  string
	Port  string
}

// ActiveMQTTBrokers returns all configured, non-empty MQTT broker endpoints.
// Leaving either MQTT_BROKER or MQTT2_BROKER empty cleanly disables that slot
// without warnings or background reconnect retries.
func (c *Config) ActiveMQTTBrokers() []MQTTBrokerEndpoint {
	if c == nil {
		return nil
	}
	var endpoints []MQTTBrokerEndpoint
	seen := make(map[string]bool)

	add := func(label, host, port string) {
		h := strings.TrimSpace(host)
		if h == "" {
			return
		}
		p := strings.TrimSpace(port)
		if p == "" {
			p = "1883"
		}
		key := h + ":" + p
		if seen[key] {
			return
		}
		seen[key] = true
		endpoints = append(endpoints, MQTTBrokerEndpoint{
			Label: label,
			Host:  h,
			Port:  p,
		})
	}

	add("PRIMARY", c.MQTTBroker, c.MQTTPort)
	add("SECONDARY", c.MQTT2Broker, c.MQTT2Port)
	return endpoints
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
