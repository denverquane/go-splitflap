package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/denverquane/go-splitflap/serdiev/usb_serial"
	"github.com/denverquane/go-splitflap/splitflap"
	"github.com/go-chi/chi/v5"
)

// UpdateDisplayRequest represents the request body for updating display text
type UpdateDisplayRequest struct {
	Text         string `json:"text"`
	DurationSecs int64  `json:"duration_secs"`
}

// SetupDisplayHandlers registers all display-related routes
func SetupDisplayHandlers(r chi.Router, display *splitflap.Display) {
	r.Get("/state", getDisplayState(display))
	r.Get("/size", getDisplaySize(display))
	r.Post("/clear", clearDisplay(display))
	r.Post("/update", updateDisplay(display))
	r.Get("/alphabet", getAlphabet())
	r.Get("/translations", getTranslations(display))
	r.Post("/translations", updateTranslations(display))
}

func getDisplayState(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, []byte(display.GetState()))
	}
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
		// Convert rune map to string map for more consistent JSON serialization
		stringMap := make(map[string]string)
		for src, dst := range display.Translations {
			stringMap[string(src)] = string(dst)
		}

		bytes, err := json.Marshal(stringMap)
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

		dur := time.Duration(req.DurationSecs) * time.Second

		// Update the display with the new text
		display.Set(req.Text, dur)

		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()

		respondJSON(w, []byte(`{"status":"ok"}`))
	}
}

// UpdateTranslationsRequest represents the request body for updating character translations
type UpdateTranslationsRequest map[string]string

// updateTranslations handles updating the character translation map
func updateTranslations(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read and parse the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Failed to read request body", "error", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req UpdateTranslationsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			slog.Error("Failed to parse request body", "error", err)
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		// Convert string map to rune map
		translations := make(map[rune]rune)
		for src, dst := range req {
			// Convert strings to runes to properly handle Unicode characters
			srcRunes := []rune(src)
			dstRunes := []rune(dst)

			if len(srcRunes) != 1 || len(dstRunes) != 1 {
				http.Error(w, "Source and destination must be single Unicode characters", http.StatusBadRequest)
				return
			}
			translations[srcRunes[0]] = dstRunes[0]
		}

		// Update display translations
		display.Translations = translations

		// Save the updated display configuration
		err = splitflap.WriteDisplayToFile(display, display.GetFilepath())
		if err != nil {
			slog.Error("Failed to save display configuration", "error", err)
			http.Error(w, "Failed to save display configuration", http.StatusInternalServerError)
			return
		}

		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()

		respondJSON(w, []byte(`{"status":"ok"}`))
	}
}
