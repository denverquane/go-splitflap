package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"github.com/denverquane/go-splitflap/routine"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Display struct {
	Size              display.Size                  `json:"size"`
	Translations      map[string]string             `json:"translations"`
	Providers         provider.Providers            `json:"providers"`
	Dashboards        map[string]Dashboard          `json:"dashboards"`
	DashboardRotation map[string]*DashboardRotation `json:"dashboard_rotation"`
	Layout            []int                         `json:"layout"`

	dashboardRotationMessages chan string

	filepath string

	activeDashboard         string
	activeDashboardRotation string

	inMessages chan routine.Message
}

func NewDisplay(size display.Size) *Display {
	layout := make([]int, size.Height*size.Width)
	for i := range layout {
		layout[i] = i
	}
	return &Display{
		Size:                      size,
		Translations:              make(map[string]string),
		Providers:                 provider.Providers{},
		Dashboards:                make(map[string]Dashboard),
		DashboardRotation:         make(map[string]*DashboardRotation),
		dashboardRotationMessages: make(chan string),
		Layout:                    layout,
		filepath:                  "",
		activeDashboard:           "",
		activeDashboardRotation:   "",
		inMessages:                make(chan routine.Message),
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

	if err := validateLayout(d.Size, d.Layout); err != nil {
		return nil, err
	}
	d.filepath = path
	d.activeDashboard = ""
	d.activeDashboardRotation = ""
	d.inMessages = make(chan routine.Message)
	d.dashboardRotationMessages = make(chan string)
	return &d, nil
}

func WriteDisplayToFile(display *Display, path string) error {
	display.filepath = path
	return display.write()
}

func (d *Display) Clear() {
	go func() {
		d.inMessages <- routine.Message{
			LocationSize: display.LocationSize{Location: display.Location{X: 0, Y: 0}, Size: display.Size{Width: d.Size.Height * d.Size.Width, Height: 1}}, //TODO
			Text:         strings.Repeat(" ", d.Size.Height*d.Size.Width),
		}
	}()
}

func (d *Display) Set(str string) {
	go func() {
		d.inMessages <- routine.Message{
			LocationSize: display.LocationSize{Location: display.Location{X: 0, Y: 0}, Size: display.Size{Width: d.Size.Height * d.Size.Width, Height: 1}}, //TODO
			Text:         str,
		}
	}()
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

func (d *Display) DeactivateActiveDashboard() {
	if d.activeDashboard == "" {
		return
	}
	dd := d.Dashboards[d.activeDashboard]
	dd.Deactivate()
	d.activeDashboard = ""
	return
}

func (d *Display) ActivateDashboard(name string) error {
	d.DeactivateActiveDashboard()
	if dashboard, ok := d.Dashboards[name]; !ok {
		return errors.New("dashboard does not exist")
	} else {
		dashboard.Activate(d.inMessages)
		d.activeDashboard = name
	}
	return nil
}

func (d *Display) AddDashboardRotation(rotationName string, rot DashboardRotation) error {
	if _, ok := d.DashboardRotation[rotationName]; ok {
		return errors.New("dashboard rotation already exists with that name")
	} else if len(rot.Rotation) < 2 {
		return errors.New("2 or more dashboards are required to form a rotation")
	} else {
		for _, r := range rot.Rotation {
			if _, okok := d.Dashboards[r.Name]; !okok {
				return errors.New("dashboard in rotation does not exist")
			}
			if r.DurationSecs < 1 {
				return errors.New("can't have a dashboard in a rotation with less than 1 sec duration")
			}
		}
		d.DashboardRotation[rotationName] = &rot
	}
	return d.write()
}

func (d *Display) ActivateDashboardRotation(name string) error {
	if rot, ok := d.DashboardRotation[name]; !ok {
		return errors.New("no dashboard rotation with that name")
	} else {
		d.DeactivateActiveDashboard()

		rot.Start(d.dashboardRotationMessages)

		slog.Info("Dashboard Rotation Started", "name", name)
		d.activeDashboardRotation = name
	}
	return nil
}

func (d *Display) DeactivateDashboardRotation() {
	if d.activeDashboardRotation == "" {
		return
	}
	if rot, ok := d.DashboardRotation[d.activeDashboardRotation]; !ok {
		slog.Error("no dashboard rotation with the value stored as the current active")
	} else {
		rot.Stop()
		d.Clear()
		slog.Info("Dashboard Rotation Stopped", "name", d.activeDashboardRotation)
		d.activeDashboardRotation = ""
	}
}

func (d *Display) Run(messages chan<- OutMessage, state <-chan string) {
	currentMessage := initMessage(d.Size)

	for {
		select {
		case name := <-d.dashboardRotationMessages:
			d.DeactivateActiveDashboard()
			d.Clear()
			err := d.ActivateDashboard(name)
			if err != nil {
				slog.Error(err.Error())
			}

		case s := <-state:
			if d.activeDashboard != "" {
				dash := d.Dashboards[d.activeDashboard]
				dash.SetState(d.Size, s)
			}

		case m := <-d.inMessages:
			if d.Translations != nil && len(d.Translations) > 0 {
				m.Text = applyTranslations(m.Text, d.Translations)
			}
			currentMessage = mergeMessageToCurrentText(d.Size, currentMessage, m)

			currentMessage = arrangeToLayout(currentMessage, d.Layout)

			messages <- OutMessage{
				payload: string(currentMessage),
			}
		}
	}
}

func initMessage(size display.Size) []byte {
	currentMessage := make([]byte, size.Height*size.Width)
	for i := range currentMessage {
		currentMessage[i] = ' ' // ASCII value for space is 32
	}
	return currentMessage
}

func applyTranslations(text string, translations map[string]string) string {
	for src, dst := range translations {
		text = strings.ReplaceAll(text, src, dst)
	}
	return text
}

func mergeMessageToCurrentText(displaySize display.Size, current []byte, msg routine.Message) []byte {
	copy(current[msg.Y*displaySize.Width+msg.X:], msg.Text)
	return current
}

func validateLayout(size display.Size, layout []int) error {
	if len(layout) != size.Height*size.Width {
		return errors.New("invalid layout size, does not match width*height")
	}
	for _, v := range layout {
		if v < 0 || v >= size.Height*size.Width {
			return errors.New("invalid layout value provided, is greater or less than display dimensions")
		}
	}
	return nil
}

func arrangeToLayout(current []byte, layout []int) []byte {
	final := make([]byte, len(current))

	for i := range layout {
		final[i] = current[layout[i]]
	}
	return final
}
