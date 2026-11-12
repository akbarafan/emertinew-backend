package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/config"
	_ "github.com/Alugara-Inovasi-Utama/emerti/docs" // Import swagger docs
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/sensor"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/telegram"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/tracking"
	"github.com/Alugara-Inovasi-Utama/emerti/routes"
	"github.com/Alugara-Inovasi-Utama/emerti/seed"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// @title           Emerti Backend API
// @version         1.0
// @description     Production-ready backend API untuk sistem Emerti dengan MQTT integration, JWT authentication, dan PostgreSQL database.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@emerti.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:6969
// @BasePath  /api

// @tag.name Authentication
// @tag.description Endpoint untuk login dan logout user

// @tag.name User Management
// @tag.description Endpoint untuk manajemen user (CRUD)

// @tag.name Master Data - Sensor
// @tag.description Endpoint untuk manajemen master data sensor

// @tag.name Master Data - Device
// @tag.description Endpoint untuk manajemen master data device

// @tag.name Master Data - Vessel
// @tag.description Endpoint untuk manajemen master data vessel/kapal

// @tag.name Sensor Logs
// @tag.description Endpoint untuk melihat logs data sensor dari MQTT

// @tag.name 7. Tracking
// @tag.description Endpoint untuk tracking realtime, per-minute, dan sensor health status

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Set timezone to Asia/Jakarta (GMT+7)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Println("[CONFIG] Failed to load Asia/Jakarta timezone, using UTC")
	} else {
		time.Local = loc
		log.Println("[CONFIG] Timezone set to Asia/Jakarta (GMT+7)")
	}

	// Parse command line flags
	seedOnly := flag.Bool("seed", false, "Run database seeding only (master data)")
	seedTrack := flag.Bool("seedtrack", false, "Run tracking data seeding only")
	runMigrate := flag.Bool("migrate", false, "Run SQL migrations from migrations folder")
	flag.Parse()

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	config.ConnectDatabase(*cfg)

	// Run SQL migrations if -migrate flag is provided
	if *runMigrate {
		runMigrations()
		return
	}

	// Initialize Redis
	if err := config.InitRedis(cfg); err != nil {
		log.Fatalf("[DB] Failed to connect to Redis: %v", err)
	}
	defer config.CloseRedis()

	// Run database seeding only if -seed flag is provided
	if *seedOnly {
		seed.RunAll(config.DB)
		log.Println("[SEED] Completed. Exiting...")
		return
	}

	// Run tracking seed only if -seedtrack flag is provided
	if *seedTrack {
		seed.RunTrackingSeed(config.DB)
		log.Println("[SEED] Tracking seed completed. Exiting...")
		return
	}

	// Setup Gin Router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("[SERVER] Running in production mode")
	} else {
		log.Println("[SERVER] Running in development mode")
	}
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Setup all routes (master routes)
	// Initialize Telegram Bot Service for CS & SLA Helpdesk
	dashboardBase := cfg.WebPortalURL
	if dashboardBase == "" {
		dashboardBase = cfg.AppBaseURL
	}
	telegramService := telegram.NewService(
		cfg.TelegramBotToken,
		cfg.TelegramSupportChatID,
		cfg.TelegramWebhookURL,
		dashboardBase,
	)

	allRoutes := routes.SetupRoutes(r, config.DB, telegramService)

	// Start Telegram polling if configured and enabled
	if telegramService.IsEnabled() {
		log.Printf("[TELEGRAM] ✅ Bot enabled and active: ChatID=%s", cfg.TelegramSupportChatID)
		telegramService.StartPolling()
		defer telegramService.StopPolling()
	} else {
		log.Printf("[TELEGRAM] ⚠️ Telegram Bot in standby: TokenLen=%d ChatID='%s'. Set TELEGRAM_BOT_TOKEN and TELEGRAM_SUPPORT_CHAT_ID in .env to activate.", len(cfg.TelegramBotToken), cfg.TelegramSupportChatID)
	}

	// Start tracking scheduler (processes newest data first for fast catch-up)
	trackingScheduler := tracking.NewTrackingScheduler(config.DB, "ALL")
	trackingScheduler.Start()
	defer trackingScheduler.Stop()

	// ── MQTT Brokers (single or dual, empty slot cleanly skipped) ────────────
	activeBrokers := cfg.ActiveMQTTBrokers()
	type activeMQTTClient struct {
		endpoint config.MQTTBrokerEndpoint
		service  *sensor.MQTTService
	}
	var mqttClients []activeMQTTClient

	if len(activeBrokers) == 0 {
		log.Println("[MQTT] Standby: MQTT_BROKER and MQTT2_BROKER are empty — MQTT ingestion disabled")
	} else {
		for _, ep := range activeBrokers {
			log.Printf("[CONFIG] Starting MQTT-%s client for %s:%s", ep.Label, ep.Host, ep.Port)
			svc := sensor.NewMQTTService(ep.Host, ep.Port, allRoutes.SensorService, ep.Label)
			svc.SetTrackingService(allRoutes.TrackingService)
			allRoutes.SensorService.AddMQTTNotifier(svc)
			connectMQTT(svc, ep.Label)
			mqttClients = append(mqttClients, activeMQTTClient{endpoint: ep, service: svc})
		}
		defer func() {
			for _, c := range mqttClients {
				c.service.Disconnect()
			}
		}()
	}

	allRoutes.SetMQTTStatusProvider(func() []routes.MQTTBrokerHealth {
		out := make([]routes.MQTTBrokerHealth, 0, len(mqttClients))
		for _, c := range mqttClients {
			out = append(out, routes.MQTTBrokerHealth{
				Label:     c.endpoint.Label,
				Broker:    c.endpoint.Host + ":" + c.endpoint.Port,
				Connected: c.service.IsConnected(),
			})
		}
		return out
	})

	log.Printf("[SERVER] Starting on :%s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

// connectMQTT attempts to connect and subscribe. Retries in background on failure.
func connectMQTT(mqttService *sensor.MQTTService, label string) {
	if err := mqttService.Connect(); err != nil {
		log.Printf("[MQTT-%s] Warning: Failed to connect to broker: %v", label, err)
		log.Printf("[MQTT-%s] Will keep trying to reconnect in the background...", label)

		go func() {
			retryInterval := 10 * time.Second
			for {
				time.Sleep(retryInterval)
				log.Printf("[MQTT-%s] Attempting to reconnect...", label)
				if err := mqttService.Connect(); err != nil {
					log.Printf("[MQTT-%s] Reconnection failed: %v", label, err)
				} else {
					log.Printf("[MQTT-%s] Reconnected successfully", label)
					if err := mqttService.SubscribeAll(); err != nil {
						log.Printf("[MQTT-%s] Warning: Failed to subscribe to sensor topics: %v", label, err)
					}
					break
				}
			}
		}()
	} else {
		if err := mqttService.SubscribeAll(); err != nil {
			log.Printf("[MQTT-%s] Warning: Failed to subscribe to sensor topics: %v", label, err)
		}
	}
}

// runMigrations executes SQL migration files from the migrations folder
func runMigrations() {
	log.Println("[DB] Running SQL migrations...")

	migrationsDir := "migrations"

	// Get all .sql files in migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("[DB] Failed to read migrations directory: %v", err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		log.Println("[DB] No SQL migration files found")
		return
	}

	log.Printf("📁 Found %d migration files", len(sqlFiles))

	// Execute each migration file
	successCount := 0
	for _, fileName := range sqlFiles {
		filePath := filepath.Join(migrationsDir, fileName)
		log.Printf("📄 Running migration: %s", fileName)

		// Read SQL file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("[DB] Failed to read %s: %v", fileName, err)
			continue
		}

		// Execute SQL
		sqlContent := string(content)
		if err := config.DB.Exec(sqlContent).Error; err != nil {
			log.Printf("[DB] Failed to execute %s: %v", fileName, err)
			continue
		}

		log.Printf("[DB] Migration completed: %s", fileName)
		successCount++
	}

	fmt.Println()
	log.Printf("🏁 Migrations finished: %d/%d successful", successCount, len(sqlFiles))
}
