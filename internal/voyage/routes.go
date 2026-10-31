package voyage

import (
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	v := protected.Group("/voyage")
	{
		// Read endpoints available to all authenticated roles (including viewer)
		v.GET("", handler.GetAllVoyagesHandler)
		v.GET("/active", handler.GetAllActiveVoyagesHandler)
		v.GET("/active/:vessel_id", handler.GetActiveVoyageHandler)
		v.GET("/vessel/:vessel_id", handler.GetVoyagesByVesselHandler)

		// Mutation endpoints require voyage manager role (superadmin, company_admin, operator)
		v.POST("", middleware.RequireVoyageManager(), handler.CreateVoyageHandler)
		v.POST("/:id/complete", middleware.RequireVoyageManager(), handler.CompleteVoyageHandler)
		v.PATCH("/:id", middleware.RequireVoyageManager(), handler.UpdateVoyageHandler)
		v.DELETE("/:id", middleware.RequireVoyageManager(), handler.DeleteVoyageHandler)
	}
}
