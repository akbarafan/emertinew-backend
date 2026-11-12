package ticket

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/telegram"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup, db *gorm.DB, tg telegram.Service) Service {
	repo := NewRepository(db)
	service := NewService(repo)
	if tg != nil {
		service.SetTelegramService(tg)
		tg.SetTicketHandler(service)
	}
	handler := NewHandler(service, tg)

	// Public Webhook route for Telegram Bot (no JWT authentication required)
	if public != nil {
		public.POST("/support/telegram/webhook", handler.TelegramWebhookHandler)
	}

	// Protected routes (JWT authentication required)
	t := protected.Group("/support")
	{
		// Static routes registered first to prevent wildcard parameter collisions in Gin router
		t.POST("/tickets/bulk-delete", handler.BulkDeleteTicketsHandler)
		t.DELETE("/tickets/bulk", handler.BulkDeleteTicketsHandler)
		t.GET("/metrics", handler.GetMetricsHandler)
		t.POST("/upload", handler.UploadAttachmentHandler)

		// Ticket collection & detail routes
		t.GET("/tickets", handler.GetAllTicketsHandler)
		t.POST("/tickets", handler.CreateTicketHandler)
		t.GET("/tickets/:id", handler.GetTicketByIDHandler)
		t.DELETE("/tickets/:id", handler.DeleteTicketHandler)
		t.POST("/tickets/:id/acknowledge", handler.AcknowledgeTicketHandler)
		t.POST("/tickets/:id/reply", handler.ReplyTicketHandler)
		t.POST("/tickets/:id/resolve", handler.ResolveTicketHandler)
		t.POST("/tickets/:id/close", handler.CloseTicketHandler)
	}

	return service
}
