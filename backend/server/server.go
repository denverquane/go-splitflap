package server

import (
	"github.com/denverquane/go-splitflap/splitflap"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

// Global WebSocket manager to broadcast updates
var WebSocketMgr *WebSocketManager

// Run initializes and starts the HTTP server
func Run(port string, display *splitflap.Display) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	
	// Initialize WebSocket manager
	WebSocketMgr = NewWebSocketManager(display)
	
	// Set up API routes
	r.Route("/display", func(r chi.Router) {
		SetupDisplayHandlers(r, display)
	})
	
	r.Route("/routines", func(r chi.Router) {
		SetupRoutineHandlers(r)
	})
	
	r.Route("/dashboards", func(r chi.Router) {
		SetupDashboardHandlers(r, display)
	})
	
	r.Route("/rotations", func(r chi.Router) {
		SetupRotationHandlers(r, display)
	})
	
	// Set up WebSocket route
	SetupWebSocketRoutes(r, WebSocketMgr)
	
	// Start periodic state broadcast (every 5 seconds)
	go startPeriodicBroadcast(WebSocketMgr)
	
	slog.Info("Server started on port " + port)
	slog.Info("WebSocket endpoint available at ws://localhost:" + port + "/ws")
	
	// Start the server
	return http.ListenAndServe(":"+port, r)
}

// startPeriodicBroadcast periodically sends state updates to all clients
func startPeriodicBroadcast(wsManager *WebSocketManager) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			wsManager.BroadcastState()
		}
	}
}

// BroadcastStateChange can be called to immediately broadcast state to all clients
func BroadcastStateChange() {
	if WebSocketMgr != nil {
		WebSocketMgr.BroadcastState()
	}
}
