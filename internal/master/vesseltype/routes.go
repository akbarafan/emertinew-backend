package vesseltype

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

	vesselTypes := r.Group("/vessel-type")
	{
		vesselTypes.POST("", middleware.RequireSuperadmin(), handler.CreateVesselType)
		vesselTypes.GET("", middleware.RequireGuest(), handler.GetAllVesselTypes)
		vesselTypes.GET("/:id", middleware.RequireGuest(), handler.GetVesselTypeByID)
		vesselTypes.PATCH("/:id", middleware.RequireSuperadmin(), handler.UpdateVesselType)
		vesselTypes.DELETE("/:id", middleware.RequireSuperadmin(), handler.DeleteVesselType)
	}
}
