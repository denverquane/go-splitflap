package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Display struct {
	Size         display.Size  `json:"size"`
	Translations map[rune]rune `json:"translations"`
	//Providers         provider.Providers            `json:"providers"`
	Dashboards        map[string]*Dashboard         `json:"dashboards"`
	DashboardRotation map[string]*DashboardRotation `json:"dashboard_rotation"`
	Layout            []int                         `json:"layout"`
	PollRate          int64                         `json:"poll_rate_ms"`

	activeDashboard string
	activeRotation  string

	state                     string
	stateSubscriber           chan<- struct{}
	dashboardRotationMessages chan string
	filepath                  string
	inMessages                chan routine.Message
}

func NewDisplay(size display.Size) *Display {
	layout := make([]int, size.Height*size.Width)
	for i := range layout {
		layout[i] = i
	}
	return &Display{
		Size:         size,
		Translations: make(map[rune]rune),
		//Providers:                 provider.Providers{},
		Dashboards:                make(map[string]*Dashboard),
		DashboardRotation:         make(map[string]*DashboardRotation),
		dashboardRotationMessages: make(chan string),
		Layout:                    layout,

		activeDashboard: "",
		activeRotation:  "",
		state:           "",
		stateSubscriber: nil,
		filepath:        "",
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

	if err = validateLayout(d.Size, d.Layout); err != nil {
		return nil, err
	}
	if d.PollRate < 100 {
		return nil, errors.New("poll_rate_ms must be >= 100")
	}
	d.filepath = path
	d.activeDashboard = ""
	d.activeRotation = ""
	d.inMessages = make(chan routine.Message)
	d.dashboardRotationMessages = make(chan string)
	return &d, nil
}

func WriteDisplayToFile(display *Display, path string) error {
	display.filepath = path
	return display.write()
}

func (d *Display) ActiveDashboard() string {
	return d.activeDashboard
}

func (d *Display) ActiveRotation() string {
	return d.activeRotation
}

func (d *Display) GetState() string {
	return d.state
}

func (d *Display) GetFilepath() string {
	return d.filepath
}

func (d *Display) SetStateSubscriber(s chan struct{}) {
	d.stateSubscriber = s
}

func (d *Display) Clear() {
	d.Set(strings.Repeat(" ", d.Size.Height*d.Size.Width))
}

func (d *Display) Set(str string) {
	go func() {
		d.inMessages <- routine.Message{
			Text: str,
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
	d.Dashboards[name] = &Dashboard{Routines: []*routine.Routine{}}
	return d.write()
}

func (d *Display) DeleteDashboard(name string) error {
	if name == d.activeDashboard {
		return errors.New("cannot delete currently active dashboard")
	}

	if _, ok := d.Dashboards[name]; !ok {
		return errors.New("dashboard with that name doesn't exist")
	}

	// Check if this dashboard is part of any rotation
	for rotName, rotation := range d.DashboardRotation {
		for _, dashRot := range rotation.Rotation {
			if dashRot.Name == name {
				return errors.New("dashboard is part of rotation: " + rotName)
			}
		}
	}

	delete(d.Dashboards, name)
	return d.write()
}

func (d *Display) AddRoutineToDashboard(dashboardName string, rout routine.Routine) error {
	loc := rout.Location
	size := rout.Size

	if dashboard, ok := d.Dashboards[dashboardName]; !ok {
		return errors.New("dashboard does not exist")
	} else if _, ok = routine.AllRoutines[rout.Type]; !ok {
		return errors.New("routine type does not exist")
	} else if loc.X < 0 || loc.Y < 0 || loc.X > d.Size.Width || loc.Y > d.Size.Height {
		return errors.New("cannot add routine out of display bounds")
	} else if loc.X+size.Width > d.Size.Width || loc.Y+size.Height > d.Size.Height {
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
	d.activeDashboard = ""
	return
}

func (d *Display) ActivateDashboard(name string) error {
	d.DeactivateActiveDashboard()
	if dashboard, ok := d.Dashboards[name]; !ok {
		return errors.New("dashboard does not exist")
	} else {
		err := dashboard.Init()
		if err != nil {
			return err
		}
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
		d.activeRotation = name
	}
	return nil
}

func (d *Display) DeactivateDashboardRotation() {
	if d.activeRotation == "" {
		return
	}
	if rot, ok := d.DashboardRotation[d.activeRotation]; !ok {
		slog.Error("no dashboard rotation with the value stored as the current active")
	} else {
		rot.Stop()
		d.Clear()
		slog.Info("Dashboard Rotation Stopped", "name", d.activeRotation)
		d.activeRotation = ""
	}
}

func (d *Display) Run(messages chan<- OutMessage, state <-chan string) {

	// TODO what if our display is already running when we change the layout?
	invLayout := invertLayout(d.Layout)

	ticker := time.NewTicker(time.Millisecond * time.Duration(d.PollRate))

	for {
		select {
		case name := <-d.dashboardRotationMessages:
			d.DeactivateActiveDashboard()
			d.Clear()
			err := d.ActivateDashboard(name)
			if err != nil {
				slog.Error(err.Error())
			}

		case msg := <-d.inMessages:
			messages <- OutMessage{payload: arrangeToLayout(msg.Text, d.Layout)}

		case s := <-state:
			s = arrangeToLayout(s, invLayout)
			slog.Info("Received state from display", "state", s)
			//if d.activeDashboard != "" {
			//	dash := d.Dashboards[d.activeDashboard]
			//	dash.SetState(d.Size, s)
			//}
			d.state = s
			if d.stateSubscriber != nil {
				d.stateSubscriber <- struct{}{}
			}
		case now := <-ticker.C:
			if d.activeDashboard == "" {
				break
			}

			msgs := d.Dashboards[d.activeDashboard].Update(now)
			if len(msgs) == 0 {
				break
			}
			currentMessage := initMessage(d.Size)
			for _, m := range msgs {
				// TODO should also make sure the routine sends back *enough* text to fill its specified size?
				if len(m.Text) > m.Width*m.Height {
					slog.Error("Routine Update() returned a message that is larger than the routine's specified size", "text", m.Text)
				} else {
					currentMessage = mergeMessageToCurrentText(d.Size, currentMessage, m.Location, m.Message)
				}
			}

			messages <- OutMessage{
				payload: arrangeToLayout(string(applyTranslations(currentMessage, d.Translations)), d.Layout),
			}
		}
	}
}

func initMessage(size display.Size) []rune {
	currentMessage := make([]rune, size.Width*size.Height)
	for i := range currentMessage {
		currentMessage[i] = ' '
	}
	return currentMessage
}

func applyTranslations(text []rune, translations map[rune]rune) []rune {
	for src, dst := range translations {
		for i, v := range text {
			if v == src {
				text[i] = dst
			}
		}
	}
	return text
}

func mergeMessageToCurrentText(displaySize display.Size, current []rune, loc display.Location, msg routine.Message) []rune {
	idx := loc.Y*displaySize.Width + loc.X
	copy(current[idx:], []rune(msg.Text))
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

// invert layout is for transforming state that comes back from the display into a form that's more intuitive to work
// with (inverting the layout we've specified for the display)
func invertLayout(layout []int) []int {
	l := len(layout)
	final := make([]int, l)
	for i, v := range layout {
		final[l-1-i] = l - 1 - v
	}
	return final
}

func arrangeToLayout(current string, layout []int) string {
	final := ""
	for _, v := range layout {
		final += string(current[v])
	}
	return final
}
