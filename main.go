package main

import (
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"github.com/denverquane/go-splitflap/splitflap"
	"log/slog"
	"os"
	"time"
)

const DisplayFile = "display.json"

func main() {
	cfg, err := splitflap.ClientConfigFromEnv()
	if err != nil {
		slog.Error("Error creating Splitflap Client Config from env", "error", err.Error())
		return
	}
	splitflapClient := splitflap.NewSplitflapClient(*cfg)

	killClientChan := make(chan struct{})
	killDisplayChan := make(chan struct{})

	messages := make(chan string)

	go splitflapClient.Run(killClientChan, messages)

	hub, err := splitflap.LoadDisplayFromFile(DisplayFile)
	if err != nil {
		slog.Error("error loading display from json file", "json file", DisplayFile, "error", err.Error())
		//return
		slog.Info("creating new display and writing to file", "json file", DisplayFile)
		hub = splitflap.NewDisplay(display.Size{
			Width:  6,
			Height: 1,
		})
		err = splitflap.WriteDisplayToFile(hub, DisplayFile)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}

	hub.AddTranslation('°', 'd')

	err = hub.CreateDashboard("Weather")
	if err != nil {
		slog.Error(err.Error())
	}

	rout := routine.Routine{
		Name: "Current Temp",
		Type: routine.WEATHER,
		Routine: &routine.WeatherRoutine{
			ApiKey:       os.Getenv("OWM_API_KEY"),
			PollRateSecs: 60,
			Units:        "F",
			ShowUnits:    false,
			ShowDegree:   false,
			LocationID:   5505411,
			LocSize: display.LocationSize{
				Location: display.Location{
					X: 1, Y: 0,
				},
				Size: display.Size{
					Width:  5,
					Height: 1,
				},
			},
		},
	}

	err = hub.AddRoutineToDashboard("Weather", rout)

	if err != nil {
		slog.Error(err.Error())
	}

	go hub.Run(killDisplayChan, messages)

	err = hub.ActivateDashboard("Weather")
	if err != nil {
		slog.Error(err.Error(), "dashboard", "Weather")
		return
	}

	// TODO HTTP server here
	for {
		time.Sleep(time.Second)
	}

	killClientChan <- struct{}{}
	killDisplayChan <- struct{}{}
}
