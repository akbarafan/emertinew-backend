package fueltype

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

	fuelTypes := r.Group("/fuel-type")
	{
		fuelTypes.POST("", middleware.RequireSuperadmin(), handler.CreateFuelType)
		fuelTypes.GET("", middleware.RequireGuest(), handler.GetAllFuelTypes)
		fuelTypes.GET("/:id", middleware.RequireGuest(), handler.GetFuelType)
		fuelTypes.PATCH("/:id", middleware.RequireSuperadmin(), handler.UpdateFuelType)
		fuelTypes.DELETE("/:id", middleware.RequireSuperadmin(), handler.DeleteFuelType)
	}
}
