package ticket

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	t := protected.Group("/support")
	{
		t.GET("/tickets", handler.GetAllTicketsHandler)
		t.GET("/tickets/:id", handler.GetTicketByIDHandler)
		t.POST("/tickets", handler.CreateTicketHandler)
		t.POST("/tickets/:id/acknowledge", handler.AcknowledgeTicketHandler)
		t.POST("/tickets/:id/reply", handler.ReplyTicketHandler)
		t.POST("/tickets/:id/resolve", handler.ResolveTicketHandler)
		t.POST("/tickets/:id/close", handler.CloseTicketHandler)
		t.GET("/metrics", handler.GetMetricsHandler)
	}
}
