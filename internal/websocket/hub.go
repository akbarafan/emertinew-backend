package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	Conn   *websocket.Conn
	Send   chan []byte
	Filter interface{} // Generic filter (e.g., vessel ID, user ID)
	hub    *Hub
}

// Hub manages WebSocket connections and broadcasts
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	filterFunc FilterFunc
}

// Message represents data to broadcast
type Message struct {
	Filter interface{}
	Data   []byte
}

// FilterFunc determines if a message should be sent to a client
type FilterFunc func(clientFilter interface{}, messageFilter interface{}) bool

// NewHub creates a new WebSocket hub
func NewHub(filterFunc FilterFunc) *Hub {
	if filterFunc == nil {
		// Default: send to all clients
		filterFunc = func(clientFilter interface{}, messageFilter interface{}) bool {
			return true
		}
	}

	hub := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		filterFunc: filterFunc,
	}

	go hub.run()
	return hub
}

// run starts the hub's main loop
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client registered (total: %d)", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("[WS] Client unregistered (total: %d)", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if h.filterFunc(client.Filter, message.Filter) {
					select {
					case client.Send <- message.Data:
					default:
						// Client's send channel is full, close it
						h.mu.RUnlock()
						h.mu.Lock()
						delete(h.clients, client)
						close(client.Send)
						h.mu.Unlock()
						h.mu.RLock()
						log.Printf("WebSocket client removed due to slow consumer")
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all clients matching the filter
func (h *Hub) Broadcast(filter interface{}, data []byte) {
	select {
	case h.broadcast <- Message{Filter: filter, Data: data}:
	default:
		log.Println("[WS] Broadcast channel full, dropping message")
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	client.hub = h
	h.register <- client
}

// Unregister removes a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
