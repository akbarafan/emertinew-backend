package sensor

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(protectedRouter *gin.RouterGroup, masterRouter *gin.RouterGroup, db *gorm.DB) Service {
	// Initialize sensor module with audit
	repo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	// Sensor Logs endpoints - tidak di dalam /master
	protectedRouter.GET("/sensor-data", handler.GetSensorData)
	protectedRouter.GET("/sensor-data/:sensor_id", handler.GetSensorDataBySensorID)

	// Sensor Master CRUD endpoints - di dalam /master
	masterRouter.POST("/sensor", middleware.RequireSuperadmin(), handler.CreateSensor)
	masterRouter.GET("/sensor", middleware.RequireSuperadmin(), handler.GetAllSensors)
	masterRouter.GET("/sensor/:id", middleware.RequireSuperadmin(), handler.GetSensor)
	masterRouter.PATCH("/sensor/:id", middleware.RequireSuperadmin(), handler.UpdateSensor)
	masterRouter.DELETE("/sensor/:id", middleware.RequireSuperadmin(), handler.DeleteSensor)
	masterRouter.PATCH("/sensor/:id/visibility", middleware.RequireSuperadmin(), handler.ToggleVisibility)
	masterRouter.GET("/vessels/:vessel_id/sensor-visibility", handler.GetVesselSensorVisibility)
	masterRouter.GET("/vessels/:vessel_id/engine-mapping", handler.GetEngineMapping)

	return service
}
