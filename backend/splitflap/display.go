package splitflap

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"github.com/denverquane/go-splitflap/routine"
)

type Display struct {
	Size         display.Size                  `json:"size"`
	Translations map[rune]rune                 `json:"translations"`
	Providers    map[string]*provider.Provider `json:"providers"`
	Dashboards   map[string]*Dashboard         `json:"dashboards"`
	Layout       []int                         `json:"layout"`
	PollRate     int64                         `json:"poll_rate_ms"`

	activeDashboard string

	state           string
	stateSubscriber chan<- struct{}
	filepath        string
	inMessages      chan routine.Message
	lockoutUntil    time.Time
}

func NewDisplay(size display.Size) *Display {
	layout := make([]int, size.Height*size.Width)
	for i := range layout {
		layout[i] = i
	}
	return &Display{
		Size:         size,
		Translations: make(map[rune]rune),
		Providers:    make(map[string]*provider.Provider),
		Dashboards:   make(map[string]*Dashboard),
		Layout:       layout,

		activeDashboard: "",
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
	d.inMessages = make(chan routine.Message)
	return &d, nil
}

func WriteDisplayToFile(display *Display, path string) error {
	display.filepath = path
	return display.write()
}

func (d *Display) ActiveDashboard() string {
	return d.activeDashboard
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
	d.Set(strings.Repeat(" ", d.Size.Height*d.Size.Width), 0)
}

func (d *Display) Set(str string, duration time.Duration) {
	d.inMessages <- routine.Message{
		Text:     str,
		Duration: duration,
	}
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

	delete(d.Dashboards, name)
	return d.write()
}

// TODO prevent overlapping routines?
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
	d.deactivateProvidersForDashboard(d.activeDashboard)
	d.activeDashboard = ""
	return
}

func (d *Display) ActivateDashboard(name string) error {
	d.DeactivateActiveDashboard()
	d.activateProvidersForDashboard(name)
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

func (d *Display) Run(messages chan<- OutMessage, state <-chan string) {

	// TODO what if our display is already running when we change the layout?
	invLayout := invertLayout(d.Layout)

	providerTicker := time.NewTicker(time.Millisecond * time.Duration(d.PollRate))

	ticker := time.NewTicker(time.Millisecond * time.Duration(d.PollRate))

	values := make(provider.ProviderValues)

	// we use a single update loop to prevent races or needing locks, and communication with the splitflap is "serial" anyways
	for {
		select {

		// inMessages is the channel for actual final messages to be sent directly to the splitflap.
		// So these can come from routines further down, but also manually overridden via API endpoints, for example
		case msg := <-d.inMessages:
			// if a message provides a minimum duration, then make sure no other routines interrupt it
			if msg.Duration > 0 {
				d.lockoutUntil = time.Now().Add(msg.Duration)
			}
			messages <- OutMessage{payload: arrangeToLayout(msg.Text, d.Layout)}

			// process state received from the Splitflap
		case s := <-state:
			s = arrangeToLayout(s, invLayout)
			slog.Info("Received state from display", "state", s)

			d.state = s
			if d.stateSubscriber != nil {
				d.stateSubscriber <- struct{}{}
			}

		case <-providerTicker.C:
			for name, p := range d.Providers {
				values[name] = p.Provider.Values()
			}

			// run the update loop every tick
		case now := <-ticker.C:

			// TODO is this correct? Should a routine ever be updated if it doesn't belong to a dashboard?
			if d.activeDashboard == "" {
				break
			}
			// don't update if we have a minimum duration specified for the current state
			if now.Before(d.lockoutUntil) {
				break
			}

			msgs := d.Dashboards[d.activeDashboard].Update(now, values)
			if len(msgs) == 0 {
				break
			}
			currentMessage := initMessage(d.Size)
			for _, m := range msgs {
				// TODO should also make sure the routine sends back *enough* text to fill its specified size?
				if len([]rune(m.Text)) > m.Width*m.Height {
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

func (d *Display) activateProvidersForDashboard(dashboard string) {
	d.swapProviderPollrate(dashboard, true)
}

func (d *Display) deactivateProvidersForDashboard(dashboard string) {
	d.swapProviderPollrate(dashboard, false)
}

func (d *Display) swapProviderPollrate(dashboard string, active bool) {
	if dash, ok := d.Dashboards[dashboard]; ok {
		for _, rout := range dash.Routines {
			providerName := rout.Routine.GetProviderName()
			if providerName != "" {
				if prov, ok := d.Providers[providerName]; ok {
					var pollrate int
					if active {
						pollrate = prov.ActivePollRateSecs
					} else {
						pollrate = prov.BackgroundPollRateSecs
					}
					prov.Provider.SetPollRateSecs(pollrate)
					slog.Info("provider pollrate changed because of a change to a dependant routine", "provider", providerName, "active", active, "pollrate", pollrate)
				}
			}
		}
	}
}
