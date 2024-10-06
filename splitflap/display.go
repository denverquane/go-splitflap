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

type Display struct {
	Size         display.Size         `json:"size"`
	Translations map[byte]byte        `json:"translations"`
	Dashboards   map[string]Dashboard `json:"dashboards"`

	filepath string

	activeDashboard string

	inMessages chan routine.Message
}

func NewDisplay(size display.Size) *Display {
	return &Display{
		Size:            size,
		Translations:    make(map[byte]byte),
		Dashboards:      make(map[string]Dashboard),
		filepath:        "",
		activeDashboard: "",
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
	d.activeDashboard = ""
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
	d.Dashboards[name] = Dashboard{Routines: []routine.Routine{}}
	return d.write()
}

func (d *Display) AddRoutineToDashboard(dashboardName string, rout routine.Routine) error {
	locAndSize := rout.Routine.LocationSize()
	if dashboard, ok := d.Dashboards[dashboardName]; !ok {
		return errors.New("dashboard does not exist")
	} else if _, ok = routine.AllRoutines[rout.Type]; !ok {
		return errors.New("routine type does not exist")
	} else if locAndSize.X < 0 || locAndSize.Y < 0 || locAndSize.X > d.Size.Width || locAndSize.Y > d.Size.Height {
		return errors.New("cannot add routine out of display bounds")
	} else if locAndSize.X+locAndSize.Width > d.Size.Width || locAndSize.Y+locAndSize.Height > d.Size.Height {
		return errors.New("adding routine with specified size would exceed display bounds")
	} else {
		err := dashboard.AddRoutine(rout)
		if err != nil {
			return err
		}
		d.Dashboards[dashboardName] = dashboard
		return d.write()
	}
}

func (d *Display) DeactivateDashboard() error {
	if d.activeDashboard == "" {
		return nil
	}

	for _, rout := range d.Dashboards[d.activeDashboard].Routines {
		go rout.Routine.Stop()
	}
	d.activeDashboard = ""
	return nil
}

func (d *Display) ActivateDashboard(name string) error {
	if d.activeDashboard != "" {
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
		d.activeDashboard = name
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
