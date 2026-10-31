package manifest

import (
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	r.POST("/manifest", middleware.RequireAdmin(), handler.CreateManifest)
	r.GET("/manifest", middleware.RequireAdmin(), handler.GetAllManifests)
	r.GET("/manifest/:id", middleware.RequireAdmin(), handler.GetManifest)
	r.PATCH("/manifest/:id", middleware.RequireAdmin(), handler.UpdateManifest)
	r.DELETE("/manifest/:id", middleware.RequireAdmin(), handler.DeleteManifest)
}
