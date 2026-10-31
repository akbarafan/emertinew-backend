package materialhull

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

	materialHulls := r.Group("/material-hull")
	{
		materialHulls.POST("", middleware.RequireSuperadmin(), handler.CreateMaterialHull)
		materialHulls.GET("", middleware.RequireGuest(), handler.GetAllMaterialHulls)
		materialHulls.GET("/:id", middleware.RequireGuest(), handler.GetMaterialHullByID)
		materialHulls.PATCH("/:id", middleware.RequireSuperadmin(), handler.UpdateMaterialHull)
		materialHulls.DELETE("/:id", middleware.RequireSuperadmin(), handler.DeleteMaterialHull)
	}
}
