package server

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/serdiev/usb_serial"
	"github.com/denverquane/go-splitflap/splitflap"
	"github.com/go-chi/chi/v5"
	"io"
	"log/slog"
	"net/http"
)

// UpdateDisplayRequest represents the request body for updating display text
type UpdateDisplayRequest struct {
	Text string `json:"text"`
}

// SetupDisplayHandlers registers all display-related routes
func SetupDisplayHandlers(r chi.Router, display *splitflap.Display) {
	r.Get("/size", getDisplaySize(display))
	r.Post("/clear", clearDisplay(display))
	r.Post("/update", updateDisplay(display))
	r.Get("/alphabet", getAlphabet())
	r.Get("/translations", getTranslations(display))
}

// getDisplaySize returns the current display dimensions
func getDisplaySize(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(display.Size)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, bytes)
	}
}

// clearDisplay deactivates all active dashboards/rotations and clears the display
func clearDisplay(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		display.DeactivateDashboardRotation()
		display.DeactivateActiveDashboard()
		display.Clear()

		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()

		respondJSON(w, []byte(`{"status":"ok"}`))
	}
}

// getTranslations returns the current display character translations
func getTranslations(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(display.Translations)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, bytes)
	}
}

func getAlphabet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(usb_serial.GlobalAlphabet)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, bytes)
	}
}

// updateDisplay directly sets the display text, regardless of active dashboard or rotation
func updateDisplay(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read and parse the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Failed to read request body", "error", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req UpdateDisplayRequest
		if err := json.Unmarshal(body, &req); err != nil {
			slog.Error("Failed to parse request body", "error", err)
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		// Validate the text
		if len(req.Text) == 0 {
			http.Error(w, "Text cannot be empty", http.StatusBadRequest)
			return
		}

		// Update the display with the new text
		display.Set(req.Text)

		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()

		respondJSON(w, []byte(`{"status":"ok"}`))
	}
}
