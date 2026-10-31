package tracking

import (
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/websocket"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	// Hub is the WebSocket hub for tracking broadcasts
	Hub *websocket.Hub
	// WSHandler handles WebSocket connections
	WSHandler *websocket.Handler
	// SchedulerService is used for broadcasting from scheduler
	SchedulerService Service
)

// InitializeWebSocket initializes the WebSocket hub
func InitializeWebSocket() {
	// Initialize WebSocket Hub with filter function
	Hub = websocket.NewHub(func(clientFilter interface{}, messageFilter interface{}) bool {
		// If client has no filter (nil), send all messages
		if clientFilter == nil {
			return true
		}

		// Parse client filter
		filter, ok := clientFilter.(*WebSocketFilter)
		if !ok || filter == nil {
			return true
		}

		// If no vessel_id filter, send all messages
		if filter.VesselID == nil {
			return true
		}

		// Parse message filter (TrackingBroadcast)
		broadcast, ok := messageFilter.(*TrackingBroadcast)
		if !ok || broadcast == nil {
			return false
		}

		// Check vessel_id filter
		return *filter.VesselID == broadcast.VesselID
	})

	// Create WebSocket handler with filter extractor (only vessel_id)
	WSHandler = websocket.NewHandler(Hub,
		// Filter extractor
		func(c *gin.Context) interface{} {
			// Extract vessel_id only
			if vesselIDStr := c.Query("vessel_id"); vesselIDStr != "" {
				if id, err := strconv.ParseUint(vesselIDStr, 10, 32); err == nil {
					vid := uint(id)
					return &WebSocketFilter{VesselID: &vid}
				}
			}
			return nil // No filter, subscribe to all
		},
	)
}

// RegisterPublicRoutes registers public routes (WebSocket - no auth)
func RegisterPublicRoutes(r *gin.RouterGroup) {
	// WebSocket endpoint (public - no authentication required)
	r.GET("/tracking/ws", WSHandler.HandleConnection)
}

// RegisterRoutes registers protected tracking routes
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) Service {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Set SchedulerService for broadcasting
	SchedulerService = service

	tracking := r.Group("/tracking")
	{
		// API: Tracking per-minute (excludes current minute)
		tracking.GET("", handler.GetAllTrackingHandler)
		tracking.GET("/:vessel_id", handler.GetTrackingByVesselHandler)
	}

	return service
}
