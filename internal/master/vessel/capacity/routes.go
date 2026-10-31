package vesselcapacity

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) *Handler {
	// Initialize audit service
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)

	// Initialize vessel capacity module with audit
	repo := NewRepository(db)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	// All routes nested under /vessel/:id/capacity
	router.POST("/:id/capacity", handler.CreateVesselCapacity)
	router.GET("/:id/capacity", handler.GetVesselCapacity)
	router.PATCH("/:id/capacity", handler.UpdateVesselCapacity)
	router.DELETE("/:id/capacity", handler.DeleteVesselCapacity)

	return handler
}
