package config

import (
	"log"
	"os"

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
		MQTTBroker:    getEnv("MQTT_BROKER", "192.168.18.88"),
		MQTTPort:      getEnv("MQTT_PORT", "1883"),
		MQTT2Broker:   getEnv("MQTT2_BROKER", ""),
		MQTT2Port:     getEnv("MQTT2_PORT", "8883"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
