package server

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/routine"
	"github.com/denverquane/go-splitflap/splitflap"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

// SetupDashboardHandlers registers all dashboard-related routes
func SetupDashboardHandlers(r chi.Router, display *splitflap.Display) {
	r.Get("/", getAllDashboards(display))
	r.Get("/active", getActiveDashboard(display))
	r.Post("/{dashboardName}", createOrUpdateDashboard(display))
	r.Delete("/{dashboardName}", deleteDashboard(display))
	r.Post("/{dashboardName}/activate", activateDashboard(display))
}

// getAllDashboards returns all dashboards
func getAllDashboards(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(display.Dashboards)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, bytes)
	}
}

func getActiveDashboard(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, []byte(display.ActiveDashboard()))
	}
}

// deleteDashboard removes a dashboard
func deleteDashboard(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dashboardName := chi.URLParam(r, "dashboardName")
		err := display.DeleteDashboard(dashboardName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Write([]byte(dashboardName))
	}
}

// activateDashboard makes a dashboard active on the display
func activateDashboard(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dashboardName := chi.URLParam(r, "dashboardName")
		if _, ok := display.Dashboards[dashboardName]; !ok {
			http.Error(w, "no dashboard found with that name", http.StatusBadRequest)
			return
		}
		err := display.ActivateDashboard(dashboardName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Broadcast the state change to all WebSocket clients
		BroadcastStateChange()

		w.Write([]byte(dashboardName))
	}
}

func createOrUpdateDashboard(display *splitflap.Display) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dashboardName := chi.URLParam(r, "dashboardName")
		if dashboardName == "" {
			http.Error(w, "empty dashboard name is not allowed", http.StatusBadRequest)
			return
		}

		var routineJsons []routine.RoutineJSON
		err := json.NewDecoder(r.Body).Decode(&routineJsons)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		created := false
		if _, ok := display.Dashboards[dashboardName]; !ok {
			err = display.CreateDashboard(dashboardName)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			created = true
		}

		for _, routineJson := range routineJsons {
			if rout, ok := routine.AllRoutines[routineJson.Type]; !ok {
				http.Error(w, "no routine found by that type", http.StatusBadRequest)
				return
			} else {
				err = json.Unmarshal(routineJson.Routine, &rout)
				if err != nil {
					if created {
						display.DeleteDashboard(dashboardName)
					}

					slog.Error(err.Error())
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				err = display.AddRoutineToDashboard(dashboardName, routine.Routine{
					RoutineBase: routine.RoutineBase{
						Type:     routineJson.Type,
						Location: routineJson.Location,
						Size:     routineJson.Size,
					},
					Routine: rout,
				})

				if err != nil {
					if created {
						display.DeleteDashboard(dashboardName)
					}
					slog.Error(err.Error())
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
		}
		w.Write([]byte(dashboardName))
	}
}
