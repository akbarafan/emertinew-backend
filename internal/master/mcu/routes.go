package mcu

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)

	repo := NewRepository(db)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	router.POST("/mcu", middleware.RequireHardwareAdmin(), handler.CreateMCU)
	router.GET("/mcu", middleware.RequireHardwareAdmin(), handler.GetAllMCUs)
	router.GET("/mcu/:id", middleware.RequireHardwareAdmin(), handler.GetMCU)
	router.PATCH("/mcu/:id", middleware.RequireHardwareAdmin(), handler.UpdateMCU)
	router.DELETE("/mcu/:id", middleware.RequireHardwareAdmin(), handler.DeleteMCU)

	// Bulk sensor topic conversion
	router.POST("/mcu/:id/sensors/preview-convert", middleware.RequireHardwareAdmin(), handler.PreviewConvertSensors)
	router.POST("/mcu/:id/sensors/apply-convert", middleware.RequireHardwareAdmin(), handler.ApplyConvertSensors)
}
