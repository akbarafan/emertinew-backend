package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins - configure appropriately for production
		return true
	},
}

// Handler handles WebSocket upgrade requests
type Handler struct {
	hub             *Hub
	filterExtractor func(*gin.Context) interface{}
	onConnect       func(*Client)
}

// NewHandler creates a new WebSocket handler
// onConnect is optional callback that runs after client is registered
func NewHandler(hub *Hub, filterExtractor func(*gin.Context) interface{}, onConnect ...func(*Client)) *Handler {
	h := &Handler{
		hub:             hub,
		filterExtractor: filterExtractor,
	}
	if len(onConnect) > 0 {
		h.onConnect = onConnect[0]
	}
	return h
}

// HandleConnection handles WebSocket connection upgrade
func (h *Handler) HandleConnection(c *gin.Context) {
	// Extract filter from request (e.g., vessel_id from query params)
	var filter interface{}
	if h.filterExtractor != nil {
		filter = h.filterExtractor(c)
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	// Create new client
	client := &Client{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Filter: filter,
	}

	// Register client with hub
	h.hub.Register(client)

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()

	log.Printf("[WS] Client connected (filter: %v)", filter)

	// Call onConnect callback if set (send historical data, etc.)
	if h.onConnect != nil {
		go h.onConnect(client)
	}
}
