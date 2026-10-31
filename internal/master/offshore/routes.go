package offshore

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

	masterRouter.POST("/offshore", middleware.RequireCompanyManager(), handler.CreateOffshore)
	masterRouter.GET("/offshore", middleware.RequireGuest(), handler.GetAllOffshores)
	masterRouter.GET("/offshore/:id", middleware.RequireGuest(), handler.GetOffshore)
	masterRouter.PATCH("/offshore/:id", middleware.RequireCompanyManager(), handler.UpdateOffshore)
	masterRouter.DELETE("/offshore/:id", middleware.RequireCompanyManager(), handler.DeleteOffshore)
}
