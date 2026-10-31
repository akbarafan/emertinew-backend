package user

import (
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Initialize user module
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	user := router.Group("/user")
	{
		// Superadmin & CompanyAdmin can manage users (scoped in handler)
		user.POST("", middleware.RequireUserAdmin(), handler.Create)
		user.PATCH("/:id", middleware.RequireUserAdmin(), handler.Update)
		user.DELETE("/:id", middleware.RequireUserAdmin(), handler.Delete)

		user.GET("", middleware.RequireUserAdmin(), handler.GetAll)
		user.GET("/current", middleware.JWTMiddleware(), handler.GetCurrent)
		user.GET("/:id", middleware.RequireUserAdmin(), handler.GetByID)
	}
}
