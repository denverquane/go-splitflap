package server

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/splitflap"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// WebSocketManager manages all websocket client connections
type WebSocketManager struct {
	// Registered clients
	clients map[*websocket.Conn]bool

	// Mutex to protect clients map
	clientsMu sync.Mutex

	// Reference to the display for accessing state
	display *splitflap.Display
}

// DisplayState represents the current state of the display
type DisplayState struct {
	ActiveDashboard string    `json:"activeDashboard"`
	State           string    `json:"state"`
	CurrentTime     time.Time `json:"currentTime"`
}

// upgrader is used to upgrade HTTP connections to WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now, production should restrict this
	CheckOrigin: func(r *http.Request) bool { return true },
}

// NewWebSocketManager creates a new WebSocketManager
func NewWebSocketManager(display *splitflap.Display) *WebSocketManager {
	sub := make(chan struct{})
	display.SetStateSubscriber(sub)

	go func() {
		for range sub {
			BroadcastStateChange()
		}
	}()

	return &WebSocketManager{
		clients: make(map[*websocket.Conn]bool),
		display: display,
	}
}

// HandleWebSocket handles incoming WebSocket connections
func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection to WebSocket", "error", err)
		return
	}

	// Register client
	wsm.clientsMu.Lock()
	wsm.clients[conn] = true
	wsm.clientsMu.Unlock()

	// Send initial state
	wsm.sendStateToClient(conn)

	// Start go routine to handle incoming messages (pings, etc.)
	go wsm.handleConnection(conn)
}

// handleConnection handles incoming messages from a client
func (wsm *WebSocketManager) handleConnection(conn *websocket.Conn) {
	defer func() {
		// Unregister client when connection closes
		wsm.clientsMu.Lock()
		delete(wsm.clients, conn)
		wsm.clientsMu.Unlock()
		conn.Close()
	}()

	// Read loop - just to detect client disconnect for now
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			// Client disconnected or error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected close error", "error", err)
			}
			break
		}

		// Echo back any messages received (could be used for ping/pong)
		if messageType == websocket.TextMessage {
			// Just get current state if we receive any message
			wsm.sendStateToClient(conn)
		}
	}
}

// sendStateToClient sends the current display state to a specific client
func (wsm *WebSocketManager) sendStateToClient(conn *websocket.Conn) {
	state := wsm.getCurrentState()

	data, err := json.Marshal(state)
	if err != nil {
		slog.Error("Failed to marshal display state", "error", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		slog.Error("Failed to send display state", "error", err)
		return
	}
}

// BroadcastState sends the current display state to all connected clients
func (wsm *WebSocketManager) BroadcastState() {
	wsm.clientsMu.Lock()
	defer wsm.clientsMu.Unlock()

	// If no clients, no need to prepare state
	if len(wsm.clients) == 0 {
		return
	}

	state := wsm.getCurrentState()
	data, err := json.Marshal(state)
	if err != nil {
		slog.Error("Failed to marshal display state", "error", err)
		return
	}

	// Send to all clients
	for conn := range wsm.clients {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			// Remove client on error
			delete(wsm.clients, conn)
			conn.Close()
		}
	}
}

// getCurrentState gets the current state of the display
func (wsm *WebSocketManager) getCurrentState() DisplayState {
	return DisplayState{
		ActiveDashboard: wsm.display.ActiveDashboard(),
		State:           wsm.display.GetState(),
		CurrentTime:     time.Now(),
	}
}

// SetupWebSocketRoutes sets up WebSocket routes
func SetupWebSocketRoutes(r chi.Router, wsManager *WebSocketManager) {
	r.Get("/ws", wsManager.HandleWebSocket)
}
