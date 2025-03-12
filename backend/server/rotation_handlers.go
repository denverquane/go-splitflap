package server

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/splitflap"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

// SetupRotationHandlers registers all rotation-related routes
func SetupRotationHandlers(r chi.Router, display *splitflap.Display) {
	r.Get("/", getAllRotations(display))
	r.Post("/deactivate", deactivateRotation(display))
	r.Post("/{rotationName}/activate", activateRotation(display))
	r.Post("/{rotationName}", createOrUpdateRotation(display))
	r.Delete("/{rotationName}", deleteRotation(display))
}

// getAllRotations returns all dashboard rotations
func getAllRotations(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(display.DashboardRotation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, bytes)
	}
}

// deactivateRotation stops any active dashboard rotation
func deactivateRotation(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		display.DeactivateDashboardRotation()
		
		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()
		
		respondJSON(w, []byte(`{"status":"ok"}`))
	}
}

// activateRotation starts a dashboard rotation
func activateRotation(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rotationName := chi.URLParam(r, "rotationName")
		err := display.ActivateDashboardRotation(rotationName)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		
		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()
		
		w.Write([]byte(rotationName))
	}
}

// createOrUpdateRotation creates or updates a dashboard rotation
func createOrUpdateRotation(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rotationName := chi.URLParam(r, "rotationName")
		var rotation splitflap.DashboardRotation
		err := json.NewDecoder(r.Body).Decode(&rotation)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		err = display.AddDashboardRotation(rotationName, rotation)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()
		
		w.Write([]byte(rotationName))
	}
}

// deleteRotation removes a dashboard rotation
func deleteRotation(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rotationName := chi.URLParam(r, "rotationName")
		if _, ok := display.DashboardRotation[rotationName]; !ok {
			http.Error(w, "no rotation found with that name", http.StatusNotFound)
			return
		}
		delete(display.DashboardRotation, rotationName)
		
		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()
		
		w.Write([]byte(rotationName))
	}
}