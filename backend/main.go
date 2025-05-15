package main

import (
	"flag"
	"github.com/denverquane/go-splitflap/display"
	gen "github.com/denverquane/go-splitflap/serdiev/generated"
	"github.com/denverquane/go-splitflap/serdiev/usb_serial"
	"github.com/denverquane/go-splitflap/server"
	"github.com/denverquane/go-splitflap/splitflap"
	"log/slog"
	"os"
)

const DisplayFile = "display.json"

func main() {
	// Command line flags
	useMock := flag.Bool("mock", true, "Use mock serial connection instead of real hardware")
	port := flag.String("port", "", "Serial port to connect to when not using mock")
	flag.Parse()

	state := make(chan string)
	handleState := func(stateMsg *gen.SplitflapState) {
		if len(usb_serial.GlobalAlphabet) == 0 {
			return
		}
		comb := ""
		for _, v := range stateMsg.Modules {
			comb += string(usb_serial.GlobalAlphabet[v.FlapIndex])
		}
		state <- comb
	}

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
			os.Exit(1)
		}
	}

	messages := make(chan splitflap.OutMessage)

	var splitflapClient *splitflap.Client

	// Initialize hardware connection if requested
	if *useMock || *port != "" {
		splitflapClient = new(splitflap.Client)
		*splitflapClient = splitflap.NewSplitflapClient()

		var err error
		if *useMock {
			modules := hub.Size.Height * hub.Size.Width
			slog.Info("Using mock serial connection", "modules", modules)
			err = connectMockSerial(splitflapClient, handleState, modules)
		} else {
			slog.Info("Connecting to hardware on port", "port", *port)
			err = splitflapClient.Connect(*port, handleState)
		}

		if err != nil {
			slog.Error("Failed to connect to splitflap", "error", err.Error())
			os.Exit(1)
		} else {
			go splitflapClient.Run(messages)
		}
	} else {
		slog.Info("No hardware connection requested, running in software-only mode")
	}

	for name, prov := range hub.Providers {
		// start providers using the poll rate set for their background processing
		pollRateSecs := prov.BackgroundPollRateSecs
		prov.Provider.SetPollRateSecs(pollRateSecs)
		err = prov.Provider.Start()
		if err != nil {
			slog.Error("failed to start provider with error", "error", err.Error(), "provider", name)
			return
		}
	}

	go hub.Run(messages, state)

	err = server.Run("3000", hub)
	if err != nil {
		slog.Error(err.Error())
	}
}

// Connect to a mock serial device
func connectMockSerial(client *splitflap.Client, notify func(state *gen.SplitflapState), modules int) error {
	mockConn := usb_serial.NewMockConnection(modules)
	sf := usb_serial.NewSplitflap(mockConn, notify, modules)
	sf.Start()

	client.SetSerial(sf)
	return nil
}
