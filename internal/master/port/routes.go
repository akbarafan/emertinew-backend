package port

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(masterRouter *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	masterRouter.POST("/port", middleware.RequireSuperadmin(), handler.CreatePort)
	masterRouter.GET("/port", middleware.RequireGuest(), handler.GetAllPorts)
	masterRouter.GET("/port/:id", middleware.RequireGuest(), handler.GetPort)
	masterRouter.PATCH("/port/:id", middleware.RequireSuperadmin(), handler.UpdatePort)
	masterRouter.DELETE("/port/:id", middleware.RequireSuperadmin(), handler.DeletePort)
}
