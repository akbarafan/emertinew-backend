package audit

import (
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) *Handler {
	// Initialize audit module
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	audit := router.Group("/audit")
	audit.Use(middleware.RequireAdmin()) // Only admin can access audit logs
	{
		audit.GET("/history", handler.GetHistory)
		audit.GET("/activity", handler.GetUserActivity)
	}

	return handler
}
