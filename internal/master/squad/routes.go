package squad

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

	squads := r.Group("/squad")
	{
		squads.POST("", middleware.RequireCompanyManager(), handler.CreateSquad)
		squads.GET("", middleware.RequireGuest(), handler.GetAllSquads)
		squads.GET("/by-fleet", middleware.RequireGuest(), handler.GetSquadsByFleetID)
		squads.GET("/:id", middleware.RequireGuest(), handler.GetSquadByID)
		squads.PATCH("/:id", middleware.RequireCompanyManager(), handler.UpdateSquad)
		squads.DELETE("/:id", middleware.RequireCompanyManager(), handler.DeleteSquad)
	}
}
