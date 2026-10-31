package fleet

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)

	repo := NewRepository(db)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	fleets := r.Group("/fleet")
	{
		fleets.POST("", middleware.RequireCompanyManager(), handler.CreateFleet)
		fleets.GET("", middleware.RequireGuest(), handler.GetAllFleets)
		fleets.GET("/:id", middleware.RequireGuest(), handler.GetFleetByID)
		fleets.PATCH("/:id", middleware.RequireCompanyManager(), handler.UpdateFleet)
		fleets.DELETE("/:id", middleware.RequireCompanyManager(), handler.DeleteFleet)
	}
}
