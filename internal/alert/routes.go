package alert

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	alerts := r.Group("/alerts")
	{
		alerts.POST("/events", handler.SyncEventsHandler)
		alerts.GET("/events", handler.ListEventsHandler)
		alerts.PATCH("/events/:id/ack", handler.AckEventHandler)
	}
}
