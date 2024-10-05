package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"io"
	"os"
	"strings"
)

var AllRoutines = map[routine.RoutineType]routine.RoutineIface{
	routine.CLOCK:   &routine.ClockRoutine{},
	routine.TIMER:   &routine.TimerRoutine{},
	routine.WEATHER: &routine.WeatherRoutine{},
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

func (d *Display) AddTranslation(src, dst byte) {
	d.Translations[src] = dst
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
	d.Dashboards[name] = Dashboard{Routines: make(map[string]routine.Routine)}
	d.write()
	return nil
}

func (d *Display) AddRoutineToDashboard(dashboardName string, routine routine.Routine) error {
	locAndSize := routine.GetLocationSize()
	if dashboard, ok := d.Dashboards[dashboardName]; !ok {
		return errors.New("dashboard does not exist")
	} else if _, ok := AllRoutines[routine.GetType()]; !ok {
		return errors.New("routine does not exist")
	} else if locAndSize.X < 0 || locAndSize.Y < 0 || locAndSize.X > d.Size.Width || locAndSize.Y > d.Size.Height {
		return errors.New("cannot add routine out of display bounds")
	} else if locAndSize.X+locAndSize.Width > d.Size.Width || locAndSize.Y+locAndSize.Height > d.Size.Height {
		return errors.New("adding routine with specified size would exceed display bounds")
	} else {
		config.SetLocationSize(locAndSize)
		err := dashboard.AddRoutine(routineType, routineName, config)
		if err != nil {
			return err
		}
		d.Dashboards[dashboardName] = dashboard
		d.write()
		return nil
	}
}

func (d *Display) DeactivateDashboard() error {
	if d.activeDashboard == nil {
		return nil
	}
	for _, rout := range d.activeDashboard.Routines {
		rout.Routine.Stop()
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
		err := dashboard.Activate(d.inMessages)
		if err != nil {
			return err
		}
		d.activeDashboard = &dashboard
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
			if d.Translations != nil && len(d.Translations) > 0 {
				m.Text = applyTranslations(m.Text, d.Translations)
			}
			copy(currentMessage[m.X:], m.Text)
			messages <- string(currentMessage)
		}
	}
}

func applyTranslations(text string, translations map[byte]byte) string {
	for src, dst := range translations {
		text = strings.ReplaceAll(text, string(src), string(dst))
	}
	return text
}
