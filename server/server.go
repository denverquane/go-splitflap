package server

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/routine"
	"github.com/denverquane/go-splitflap/splitflap"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
)

func Run(port string, display *splitflap.Display) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/display", func(r chi.Router) {
		r.Get("/size", func(w http.ResponseWriter, r *http.Request) {
			bytes, err := json.Marshal(display.Size)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write(bytes)
		})
		r.Post("/repeat/{char}", func(w http.ResponseWriter, r *http.Request) {
			char := chi.URLParam(r, "char")
			str := strings.Repeat(string(char[0]), display.Size.Width*display.Size.Height)
			display.Set(str)
			w.Write([]byte("ok"))
		})

		//r.Post("/size", func(w http.ResponseWriter, r *http.Request) {
		//	var size display2.Size
		//	err := json.NewDecoder(r.Body).Decode(&size)
		//	if err != nil {
		//		slog.Error(err.Error())
		//		http.Error(w, err.Error(), 422)
		//		return
		//	}
		//	if size.Width < 1 || size.Height < 1 {
		//		http.Error(w, "invalid width/height supplied", 400)
		//		return
		//	}
		//	display.Size = size
		//	// TODO need to check all routines and make sure they still fit in this display...
		//	// TODO should the display be allowed to change sizes?
		//})
		r.Get("/clear", func(w http.ResponseWriter, r *http.Request) {
			display.Clear()
		})
		r.Get("/translations", func(w http.ResponseWriter, r *http.Request) {
			bytes, err := json.Marshal(display.Translations)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write(bytes)
		})
	})
	r.Route("/routines", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			bytes, err := json.Marshal(routine.AllRoutines)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write(bytes)
		})
	})
	r.Route("/dashboards", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			dashboards := slices.Collect(maps.Keys(display.Dashboards))
			bytes, err := json.Marshal(dashboards)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write(bytes)
		})
		r.Post("/deactivate", func(w http.ResponseWriter, r *http.Request) {
			display.DeactivateActiveDashboard()
			w.Write([]byte("ok"))
		})
		r.Get("/{dashboardName}", func(w http.ResponseWriter, r *http.Request) {
			dashboardName := chi.URLParam(r, "dashboardName")
			if dashboard, ok := display.Dashboards[dashboardName]; !ok {
				http.Error(w, "no dashboard found with that name", 400)
				return
			} else {
				bytes, err := json.Marshal(dashboard)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				w.Write(bytes)
			}
		})
		r.Post("/{dashboardName}", func(w http.ResponseWriter, r *http.Request) {
			dashboardName := chi.URLParam(r, "dashboardName")
			err := display.CreateDashboard(dashboardName)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			w.Write([]byte(dashboardName))
		})
		r.Post("/{dashboardName}/activate", func(w http.ResponseWriter, r *http.Request) {
			dashboardName := chi.URLParam(r, "dashboardName")
			if _, ok := display.Dashboards[dashboardName]; !ok {
				http.Error(w, "no dashboard found with that name", 400)
				return
			}
			err := display.ActivateDashboard(dashboardName)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write([]byte(dashboardName))
		})
		r.Post("/{dashboardName}/routines", func(w http.ResponseWriter, r *http.Request) {
			dashboardName := chi.URLParam(r, "dashboardName")
			if _, ok := display.Dashboards[dashboardName]; !ok {
				http.Error(w, "no dashboard found with that name", 400)
				return
			}
			var routineJson routine.RoutineJSON
			err := json.NewDecoder(r.Body).Decode(&routineJson)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, err.Error(), 422)
				return
			}
			if rout, ok := routine.AllRoutines[routineJson.Type]; !ok {
				http.Error(w, "no routine found by that type", 400)
				return
			} else {
				err = json.Unmarshal(routineJson.Routine, &rout)
				if err != nil {
					slog.Error(err.Error())
					http.Error(w, err.Error(), 400)
					return
				} else {
					err = display.AddRoutineToDashboard(dashboardName, routine.Routine{
						Name:    routineJson.Name,
						Type:    routineJson.Type,
						Routine: rout,
					})
					if err != nil {
						slog.Error(err.Error())
						http.Error(w, err.Error(), 400)
						return
					} else {
						w.Write([]byte(routineJson.Name))
					}
				}
			}
		})
	})

	r.Route("/rotations", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			rotations := slices.Collect(maps.Keys(display.DashboardRotation))
			bytes, err := json.Marshal(rotations)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write(bytes)
		})
		r.Post("/deactivate", func(w http.ResponseWriter, r *http.Request) {
			display.DeactivateDashboardRotation()
			display.DeactivateActiveDashboard()
			w.Write([]byte("ok"))
		})
		r.Post("/{rotationName}/activate", func(w http.ResponseWriter, r *http.Request) {
			rotationName := chi.URLParam(r, "rotationName")
			err := display.ActivateDashboardRotation(rotationName)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, err.Error(), 422)
				return
			}
			w.Write([]byte(rotationName))
		})
		r.Post("/{rotationName}", func(w http.ResponseWriter, r *http.Request) {
			rotationName := chi.URLParam(r, "rotationName")
			var rotation splitflap.DashboardRotation
			err := json.NewDecoder(r.Body).Decode(&rotation)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, err.Error(), 422)
				return
			}
			err = display.AddDashboardRotation(rotationName, rotation)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write([]byte(rotationName))
		})
	})
	return http.ListenAndServe(":"+port, r)
}
