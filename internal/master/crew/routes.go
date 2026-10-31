package crew

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

	masterRouter.POST("/crew", middleware.RequireCompanyManager(), handler.CreateCrew)
	masterRouter.GET("/crew", middleware.RequireGuest(), handler.GetAllCrews)
	masterRouter.GET("/crew/:id", middleware.RequireGuest(), handler.GetCrew)
	masterRouter.PATCH("/crew/:id", middleware.RequireCompanyManager(), handler.UpdateCrew)
	masterRouter.DELETE("/crew/:id", middleware.RequireCompanyManager(), handler.DeleteCrew)
}
