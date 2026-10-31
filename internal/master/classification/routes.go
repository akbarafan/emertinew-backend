package classification

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

	classifications := r.Group("/classification")
	{
		classifications.POST("", middleware.RequireSuperadmin(), handler.CreateClassification)
		classifications.GET("", middleware.RequireGuest(), handler.GetAllClassifications)
		classifications.GET("/:id", middleware.RequireGuest(), handler.GetClassificationByID)
		classifications.PATCH("/:id", middleware.RequireSuperadmin(), handler.UpdateClassification)
		classifications.DELETE("/:id", middleware.RequireSuperadmin(), handler.DeleteClassification)
	}
}
