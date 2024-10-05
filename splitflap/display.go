package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"io"
	"log/slog"
	"os"
)

var Routines = map[routine.RoutineName]routine.RoutineIface{
	routine.CLOCK: &routine.ClockRoutine{},
	routine.TIMER: &routine.TimerRoutine{},
}

type Display struct {
	Size         display.Size
	Translations map[byte]byte
	Dashboards   map[string]Dashboard

	filepath string

	activeDashboard *Dashboard

	inMessages chan routine.Message
}

func NewDisplay(size display.Size) *Display {
	return &Display{
		Size:            size,
		Translations:    make(map[byte]byte),
		Dashboards:      make(map[string]Dashboard),
		filepath:        "",
		activeDashboard: nil,
		inMessages:      make(chan routine.Message),
	}
}

func LoadDisplayFromFile(path string) (*Display, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var d Display
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		return nil, err
	}
	d.filepath = path
	d.activeDashboard = nil
	d.inMessages = make(chan routine.Message)
	return &d, nil
}

func WriteDisplayToFile(display *Display, path string) error {
	display.filepath = path
	return display.write()
}

func (d *Display) write() error {
	if d.filepath == "" {
		return errors.New("filepath not set in Display struct")
	}
	f, err := os.Create(d.filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	_, err = f.Write(bytes)
	return err
}

func (d *Display) CreateDashboard(name string) error {
	if _, ok := d.Dashboards[name]; ok {
		return errors.New("dashboard already exists with that name")
	}
	d.Dashboards[name] = Dashboard{RoutineConfig: make(map[routine.RoutineName]routine.RoutineConfigIface)}
	d.write()
	return nil
}

func (d *Display) AddRoutineToDashboard(routineName routine.RoutineName, dashboardName string, config routine.RoutineConfigIface) error {
	if d.activeDashboard != nil {
		if _, ok := d.activeDashboard.RoutineConfig[routineName]; ok {
			return errors.New("can't add routine while this routine is active on the current running dashboard")
		}
	}
	locAndSize := config.GetLocationSize()
	if dashboard, ok := d.Dashboards[dashboardName]; !ok {
		return errors.New("dashboard does not exist")
	} else if routine, ok := Routines[routineName]; !ok {
		return errors.New("routine does not exist")
	} else if _, ok = dashboard.RoutineConfig[routineName]; ok {
		return errors.New("routine is already present in this dashboard")
	} else if locAndSize.X < 0 || locAndSize.Y < 0 || locAndSize.X > d.Size.Width || locAndSize.Y > d.Size.Height {
		return errors.New("cannot add routine out of display bounds")
	} else if locAndSize.X+locAndSize.Width > d.Size.Width || locAndSize.Y+locAndSize.Height > d.Size.Height {
		return errors.New("adding routine with specified size would exceed display bounds")
	} else {
		// TODO check and make sure this doesn't overlap existing routines
		config.SetLocationSize(locAndSize)
		err := routine.SetConfig(config)
		if err != nil {
			return err
		}
		dashboard.RoutineConfig[routine.Name()] = config
		d.write()
		return nil
	}
}

func (d *Display) DeactivateDashboard() error {
	if d.activeDashboard == nil {
		return nil
	}

	routineNames := d.activeDashboard.RoutineNames()
	for _, routineName := range routineNames {
		Routines[routineName].Stop()
	}
	d.activeDashboard = nil
	return nil
}

func (d *Display) ActivateDashboard(name string) error {
	if d.activeDashboard != nil {
		err := d.DeactivateDashboard()
		if err != nil {
			return err
		}
	}
	if dashboard, ok := d.Dashboards[name]; !ok {
		return errors.New("dashboard does not exist")
	} else {
		// first iterate the routines and make sure the config works
		for routineName, routineConfig := range dashboard.RoutineConfig {
			rout := Routines[routineName]
			err := rout.SetConfig(routineConfig)
			if err != nil {
				return err
			} else {
				slog.Info("Set Routine Config", "dashboard", name, "routine", routineName)
			}
		}

		// once we've checked all the routines, then start each of them
		for routineName := range dashboard.RoutineConfig {
			rout := Routines[routineName]
			err := rout.Start(d.inMessages)
			if err != nil {
				return err
			} else {
				slog.Info("Activated Dashboard Routine", "dashboard", name, "routine", routineName)
			}
		}
	}
	return nil
}

func (d *Display) Run(kill <-chan struct{}, messages chan<- string) {
	currentMessage := make([]byte, d.Size.Height*d.Size.Width)
	for i := range currentMessage {
		currentMessage[i] = ' ' // ASCII value for space is 32
	}

	for {
		select {
		case <-kill:
			return
		case m := <-d.inMessages:
			copy(currentMessage[m.X:], m.Text)
			messages <- string(currentMessage)
		}
	}
}
