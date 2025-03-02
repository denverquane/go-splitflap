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
	state := make(chan string)

	splitflapClient := splitflap.NewSplitflapClient()
	err := splitflapClient.Connect("COM5", state)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	messages := make(chan splitflap.OutMessage)

	go splitflapClient.Run(messages)

	hub, err := splitflap.LoadDisplayFromFile(DisplayFile)
	if err != nil {
		slog.Error("error loading display from json file", "json file", DisplayFile, "error", err.Error())
		if _, err = os.Stat(DisplayFile); os.IsNotExist(err) {
			slog.Info("file not found, creating new display and writing to file", "json file", DisplayFile)
			hub = splitflap.NewDisplay(display.Size{
				Width:  12,
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

	err = hub.Providers.Start()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	go hub.Run(messages, state)

	err = server.Run("3000", hub)
	if err != nil {
		slog.Error(err.Error())
	}
}
