package server

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// SetupRoutineHandlers registers all routine-related routes
func SetupRoutineHandlers(r chi.Router) {
	r.Get("/", getAllRoutines())
}

// RoutineInfo holds the full information about a routine, including its parameters
type RoutineInfo struct {
	Parameters []routine.Parameter `json:"parameters"`
	MinSize    display.Min         `json:"min_size"`
	MaxSize    display.Max         `json:"max_size"`
	Config     interface{}         `json:"config"`
}

// getAllRoutines returns all available routine types with their parameters
func getAllRoutines() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := make(map[string]RoutineInfo)

		for routineType, routineInstance := range routine.AllRoutines {
			minS, maxS := routineInstance.SizeRange()
			routineInfo := RoutineInfo{
				Parameters: routineInstance.Parameters(),
				MinSize:    minS,
				MaxSize:    maxS,
				Config:     routineInstance,
			}
			response[string(routineType)] = routineInfo
		}

		bytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, bytes)
	}
}
