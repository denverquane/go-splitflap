package main

import (
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/server"
	"github.com/denverquane/go-splitflap/splitflap"
	"log/slog"
	"os"
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
		if _, err = os.Stat(DisplayFile); os.IsNotExist(err) {
			slog.Info("file not found, creating new display and writing to file", "json file", DisplayFile)
			hub = splitflap.NewDisplay(display.Size{
				Width:  6,
				Height: 1,
			})
			err = splitflap.WriteDisplayToFile(hub, DisplayFile)
			if err != nil {
				slog.Error(err.Error())
				return
			}
		} else {
			slog.Error("file found, but contents cannot be parsed! Exiting!")
		}
	}

	go hub.Run(killDisplayChan, messages)

	err = server.Run("3000", hub)
	if err != nil {
		slog.Error(err.Error())
	}

	killClientChan <- struct{}{}
	killDisplayChan <- struct{}{}
}
