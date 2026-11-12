package routes

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/alert"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/auth"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/fuel"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/geo"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/geofence"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/manifest"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/classification"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/company"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/crew"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/fleet"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/fueltype"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/materialhull"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/mcu"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/offshore"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/operationstatus"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/port"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/sensor"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/sensorhealth"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/squad"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/vessel"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/master/vesseltype"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/telegram"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/ticket"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/tracking"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/user"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/voyage"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// MQTTBrokerHealth describes runtime status of a configured MQTT broker.
type MQTTBrokerHealth struct {
	Label     string `json:"label"`
	Broker    string `json:"broker"`
	Connected bool   `json:"connected"`
}

type Routes struct {
	SensorService   sensor.Service
	TrackingService tracking.Service
	TelegramService telegram.Service
	TicketService   ticket.Service
	mqttStatusFn    func() []MQTTBrokerHealth
}

// SetMQTTStatusProvider registers a callback returning live status for each configured broker.
func (rt *Routes) SetMQTTStatusProvider(fn func() []MQTTBrokerHealth) {
	rt.mqttStatusFn = fn
}

func SetupRoutes(r *gin.Engine, db *gorm.DB, tg telegram.Service) *Routes {
	// Serve static uploads
	r.Static("/uploads", "./assets/uploads")

	rt := &Routes{}

	// Health check route
	r.GET("/", func(c *gin.Context) {
		var brokers []MQTTBrokerHealth
		if rt.mqttStatusFn != nil {
			brokers = rt.mqttStatusFn()
		}

		mqttStatus := "standby"
		primaryBroker := ""
		if len(brokers) > 0 {
			primaryBroker = brokers[0].Broker
			anyConnected := false
			for _, b := range brokers {
				if b.Connected {
					anyConnected = true
					break
				}
			}
			if anyConnected {
				mqttStatus = "connected"
			} else {
				mqttStatus = "reconnecting"
			}
		}

		c.JSON(200, gin.H{
			"message":      "Emerti Backend API",
			"mqtt_status":  mqttStatus,
			"mqtt_broker":  primaryBroker,
			"mqtt_brokers": brokers,
			"note":         "Using dynamic sensor subscriptions from master data",
		})
	})

	// Swagger documentation
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API group
	api := r.Group("/api")

	// JWT middleware
	jwtMiddleware := middleware.JWTMiddleware()

	// Register auth routes
	auth.RegisterRoutes(api, db, jwtMiddleware)

	// Initialize WebSocket for tracking (before any routes)
	tracking.InitializeWebSocket()

	// Register public routes (no authentication)
	tracking.RegisterPublicRoutes(api)

	// Register protected routes
	protected := api.Group("/")
	protected.Use(jwtMiddleware)

	// Master data routes
	master := protected.Group("/master")
	var sensorService sensor.Service
	sensorService = sensor.RegisterRoutes(protected, master, db)
	mcu.RegisterRoutes(master, db)
	vessel.RegisterRoutes(master, db)
	fueltype.RegisterRoutes(master, db)
	offshore.RegisterRoutes(master, db)
	port.RegisterRoutes(master, db)
	crew.RegisterRoutes(master, db)
	classification.RegisterRoutes(master, db)
	vesseltype.RegisterRoutes(master, db)
	operationstatus.RegisterRoutes(master, db)
	materialhull.RegisterRoutes(master, db)
	company.RegisterRoutes(master, db)
	fleet.RegisterRoutes(master, db)
	squad.RegisterRoutes(master, db)
	sensorhealth.RegisterRoutes(master, db)

	// Other protected routes
	user.RegisterRoutes(protected, db)
	audit.RegisterRoutes(protected, db)
	geofence.RegisterRoutes(protected, db)
	geo.RegisterRoutes(protected, db).Start()
	manifest.RegisterRoutes(protected, db)
	trackingService := tracking.RegisterRoutes(protected, db)
	trackingService.CleanupHistoricalSpikes()
	fuel.RegisterRoutes(protected, db)
	alert.RegisterRoutes(protected, db)
	voyage.RegisterRoutes(protected, db)
	ticketService := ticket.RegisterRoutes(api, protected, db, tg)

	rt.SensorService = sensorService
	rt.TrackingService = trackingService
	rt.TelegramService = tg
	rt.TicketService = ticketService
	return rt
}
