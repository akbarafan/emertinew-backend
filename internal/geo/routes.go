package geo

import (
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) *GeofenceTicker {
	repo := NewRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service)

	geo := router.Group("/geo")
	{
		assignments := geo.Group("/assignments")
		{
			assignments.POST("", middleware.RequireAdmin(), handler.CreateAssignment)
			assignments.GET("", middleware.JWTMiddleware(), handler.ListAssignments)
			assignments.GET("/:id", middleware.JWTMiddleware(), handler.GetAssignment)
			assignments.PATCH("/:id", middleware.RequireAdmin(), handler.UpdateAssignment)
			assignments.DELETE("/:id", middleware.RequireAdmin(), handler.DeleteAssignment)
		}

		geo.GET("/cp-logs", middleware.JWTMiddleware(), handler.ListCpLogs)
		geo.GET("/violation-logs", middleware.JWTMiddleware(), handler.ListViolationLogs)
	}

	ticker := NewGeofenceTicker(db, repo)
	return ticker
}
