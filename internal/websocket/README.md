# WebSocket Module Documentation

## Overview
WebSocket module yang menggunakan **Gorilla WebSocket** untuk real-time communication. Module ini dirancang agar reusable dan dapat digunakan untuk berbagai keperluan WebSocket di aplikasi.

## Struktur Folder
```
internal/
  websocket/
    hub.go      - WebSocket hub untuk manage clients dan broadcast
    client.go   - Client connection management dan pump logic
    handler.go  - HTTP handler untuk WebSocket upgrade
```

## Features
- ✅ **Gorilla WebSocket** - Library WebSocket yang powerful dan reliable
- ✅ **Generic Filter System** - Support filtering message berdasarkan kriteria apapun
- ✅ **Automatic Reconnection** - Ping/pong untuk keep-alive
- ✅ **Slow Consumer Protection** - Auto-disconnect client yang lambat
- ✅ **Thread-Safe** - Concurrent access menggunakan mutex
- ✅ **Customizable** - Filter function dan message handler dapat dikustomisasi

## Komponen Utama

### 1. Hub (`hub.go`)
WebSocket hub yang mengelola semua client connections dan broadcasting.

```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
    filterFunc FilterFunc
}
```

**Key Methods:**
- `NewHub(filterFunc FilterFunc) *Hub` - Create new hub dengan filter function
- `Broadcast(filter interface{}, data []byte)` - Broadcast message ke clients
- `Register(client *Client)` - Register client baru
- `Unregister(client *Client)` - Remove client
- `ClientCount() int` - Get jumlah connected clients

### 2. Client (`client.go`)
Represents individual WebSocket connection.

```go
type Client struct {
    Conn   *websocket.Conn
    Send   chan []byte
    Filter interface{}
    hub    *Hub
}
```

**Key Methods:**
- `ReadPump()` - Read messages from client (with ping/pong)
- `WritePump()` - Write messages to client (with timeout)

**Constants:**
- `writeWait` = 10 seconds - Timeout untuk write message
- `pongWait` = 60 seconds - Timeout untuk receive pong
- `pingPeriod` = 54 seconds - Interval untuk send ping
- `maxMessageSize` = 512 bytes - Max message size dari client

### 3. Handler (`handler.go`)
HTTP handler untuk upgrade connection ke WebSocket.

```go
type Handler struct {
    hub            *Hub
    filterExtractor func(*gin.Context) interface{}
}
```

**Key Methods:**
- `NewHandler(hub *Hub, filterExtractor func(*gin.Context) interface{}) *Handler`
- `HandleConnection(c *gin.Context)` - Upgrade HTTP ke WebSocket

## Usage Example

### Basic Setup

```go
import "github.com/Alugara-Inovasi-Utama/emerti/internal/websocket"

// 1. Create filter function (optional)
filterFunc := func(clientFilter interface{}, messageFilter interface{}) bool {
    // Return true to send message to client
    // Example: filter by user ID
    if clientFilter == nil {
        return true // Send to all
    }
    
    clientUserID, ok := clientFilter.(int)
    if !ok {
        return false
    }
    
    messageUserID, ok := messageFilter.(int)
    if !ok {
        return true
    }
    
    return clientUserID == messageUserID
}

// 2. Create hub
hub := websocket.NewHub(filterFunc)

// 3. Create handler with filter extractor
handler := websocket.NewHandler(hub, func(c *gin.Context) interface{} {
    // Extract filter from request
    userID := c.Query("user_id")
    if userID == "" {
        return nil
    }
    // Convert and return filter
    id, _ := strconv.Atoi(userID)
    return id
})

// 4. Register route
router.GET("/ws", handler.HandleConnection)
```

### Broadcasting Messages

```go
// Broadcast to all clients
data := []byte(`{"message": "Hello everyone"}`)
hub.Broadcast(nil, data)

// Broadcast to specific filter
data := []byte(`{"message": "Hello user 1"}`)
hub.Broadcast(1, data) // Only send to clients with filter = 1
```

### Tracking Example (Current Implementation)

```go
// Initialize hub with vessel filter
Hub = websocket.NewHub(func(clientFilter interface{}, messageFilter interface{}) bool {
    if clientFilter == nil {
        return true // Send to all vessels
    }
    
    clientVesselID, ok := clientFilter.(*uint)
    if !ok || clientVesselID == nil {
        return true
    }
    
    messageVesselID, ok := messageFilter.(uint)
    if !ok {
        return true
    }
    
    return *clientVesselID == messageVesselID
})

// Create handler with vessel_id extractor
WSHandler = websocket.NewHandler(Hub, func(c *gin.Context) interface{} {
    vesselIDStr := c.Query("vessel_id")
    if vesselIDStr == "" {
        return nil // Subscribe to all vessels
    }
    
    if id, err := strconv.ParseUint(vesselIDStr, 10, 32); err == nil {
        vid := uint(id)
        return &vid
    }
    
    return nil
})

// Broadcast tracking data
jsonData, _ := json.Marshal(trackingDTO)
Hub.Broadcast(vesselID, jsonData)
```

## Client Connection Example

### JavaScript Client

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:6969/api/tracking/ws?vessel_id=1');

ws.onopen = () => {
    console.log('✅ Connected');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('📩 Received:', data);
};

ws.onerror = (error) => {
    console.error('❌ Error:', error);
};

ws.onclose = () => {
    console.log('🔌 Disconnected');
};
```

### Go Client

```go
import "github.com/gorilla/websocket"

u := url.URL{Scheme: "ws", Host: "localhost:6969", Path: "/api/tracking/ws"}
q := u.Query()
q.Set("vessel_id", "1")
u.RawQuery = q.Encode()

conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

for {
    _, message, err := conn.ReadMessage()
    if err != nil {
        break
    }
    log.Printf("Received: %s", message)
}
```

## Architecture

### Message Flow

```
1. Client connects via HTTP
   ↓
2. Upgrade to WebSocket (Handler.HandleConnection)
   ↓
3. Create Client instance
   ↓
4. Register to Hub
   ↓
5. Start ReadPump & WritePump goroutines
   ↓
6. Hub broadcasts messages
   ↓
7. Filter function determines which clients receive message
   ↓
8. Message sent via Client.Send channel
   ↓
9. WritePump sends to WebSocket connection
```

### Thread Safety

- Hub uses `sync.RWMutex` untuk protect `clients` map
- Client registration/unregistration via channels
- Broadcast via buffered channel (256 capacity)
- Each client has buffered send channel (256 capacity)

### Slow Consumer Protection

Jika client tidak dapat menerima message dengan cepat:
1. Send channel menjadi penuh
2. Hub akan skip sending dan close client
3. Client auto-unregister
4. Connection ditutup

## Configuration

### Timeouts
```go
const (
    writeWait = 10 * time.Second  // Write timeout
    pongWait = 60 * time.Second   // Pong timeout
    pingPeriod = 54 * time.Second // Ping interval
)
```

### CORS
```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // CHANGE IN PRODUCTION!
    },
}
```

**⚠️ Production:** Implement proper origin checking!

## Best Practices

1. **Always use filter functions** - Untuk menghindari broadcast ke semua client
2. **Set proper CORS** - Restrict origins in production
3. **Monitor client count** - Use `Hub.ClientCount()` untuk monitoring
4. **Handle disconnections** - Client auto-cleanup saat disconnect
5. **Use JSON for messages** - Standardize message format
6. **Log important events** - Hub sudah include basic logging
7. **Test with load** - Test dengan banyak concurrent clients

## Troubleshooting

### Client tidak terima message
- Check filter function logic
- Verify message filter value
- Check client send channel tidak penuh

### Connection closed unexpectedly
- Check client network stability
- Verify ping/pong timeout settings
- Check for slow consumer issues

### High memory usage
- Check untuk memory leaks di filter function
- Monitor jumlah active clients
- Verify clients properly disconnect

## Integration dengan Tracking

File yang menggunakan WebSocket module:
- `internal/tracking/routes.go` - Initialize hub dan register routes
- `internal/tracking/service.go` - Broadcast tracking updates

```go
// Di service.go
Hub.Broadcast(vesselID, jsonData)
```

## Testing

```bash
# Test dengan Go client
cd cmd/test-websocket
go run main.go -vessel=1

# Test dengan browser
# Buka test-websocket.html di browser

# Test dengan PowerShell
.\test-websocket.ps1 -VesselId 1
```

## Performance Notes

- Hub dapat handle ribuan concurrent clients
- Broadcast non-blocking (buffered channel)
- Client write non-blocking dengan timeout
- Automatic cleanup untuk disconnected clients
- Minimal memory overhead per client (~1KB)

## Future Enhancements

Potential improvements:
- [ ] Add authentication via token in query params
- [ ] Add message rate limiting
- [ ] Add metrics (Prometheus)
- [ ] Add message compression
- [ ] Add message persistence (Redis)
- [ ] Add clustering support (Redis Pub/Sub)
