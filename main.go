package main

import (
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"github.com/denverquane/go-splitflap/splitflap"
	"log/slog"
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
		return
		//slog.Info("creating new display and writing to file", "json file", DisplayFile)
		//hub = splitflap.NewDisplay(display.Size{
		//	Width:  6,
		//	Height: 1,
		//})
		//err = splitflap.WriteDisplayToFile(hub, DisplayFile)
		//if err != nil {
		//	slog.Error(err.Error())
		//	return
		//}
	}

	err = hub.CreateDashboard("Clock")
	if err != nil {
		slog.Error(err.Error())
	}

	err = hub.AddRoutineToDashboard(routine.CLOCK, "Clock", &routine.ClockConfig{
		RemoveLeadingZero: false,
		Military:          false,
		Precise:           false,
		AMPMText:          false,
		LocSize: display.LocationSize{Location: display.Location{X: 0, Y: 0}, Size: display.Size{
			Width:  6,
			Height: 1,
		}},
	})
	if err != nil {
		slog.Error(err.Error())
	}

	go hub.Run(killDisplayChan, messages)

	err = hub.ActivateDashboard("Clock")
	if err != nil {
		slog.Error(err.Error(), "dashboard", "Clock")
		return
	}

	// TODO HTTP server here
	for {
		time.Sleep(time.Second)
	}

	killClientChan <- struct{}{}
	killDisplayChan <- struct{}{}
}
