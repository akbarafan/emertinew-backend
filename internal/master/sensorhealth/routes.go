package sensorhealth

import (
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(masterRouter *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(db, repo)
	handler := NewHandler(service)

	masterRouter.GET("/sensor-health", middleware.RequireHardwareAdmin(), handler.GetSensorHealth)
}
