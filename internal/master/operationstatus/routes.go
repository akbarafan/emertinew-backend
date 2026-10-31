package operationstatus

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

	operationStatuses := r.Group("/operation-status")
	{
		operationStatuses.POST("", middleware.RequireSuperadmin(), handler.CreateOperationStatus)
		operationStatuses.GET("", middleware.RequireGuest(), handler.GetAllOperationStatuses)
		operationStatuses.GET("/:id", middleware.RequireGuest(), handler.GetOperationStatusByID)
		operationStatuses.PATCH("/:id", middleware.RequireSuperadmin(), handler.UpdateOperationStatus)
		operationStatuses.DELETE("/:id", middleware.RequireSuperadmin(), handler.DeleteOperationStatus)
	}
}
