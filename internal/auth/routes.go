package auth

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc) {
	// Initialize auth module
	repo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	// Public routes
	auth := router.Group("/auth")
	{
		auth.POST("/login", handler.Login)

		// Protected routes
		auth.POST("/logout", authMiddleware, handler.Logout)
	}
}
