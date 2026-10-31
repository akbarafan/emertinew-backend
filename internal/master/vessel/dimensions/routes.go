package vesseldimensions

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) *Handler {
	// Initialize audit service
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)

	// Initialize vessel dimensions module with audit
	repo := NewRepository(db)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	// All routes nested under /vessel/:id/dimensions
	router.POST("/:id/dimensions", handler.CreateVesselDimensions)
	router.GET("/:id/dimensions", handler.GetVesselDimensions)
	router.PATCH("/:id/dimensions", handler.UpdateVesselDimensions)
	router.DELETE("/:id/dimensions", handler.DeleteVesselDimensions)

	return handler
}
