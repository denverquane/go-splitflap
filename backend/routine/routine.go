package routine

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/display"
	"time"
)

type Message struct {
	Text string
}

type RoutineType string

// RoutineIface is the "contract" that any custom functionality should conform to
type RoutineIface interface {
	SizeRange() (display.Min, display.Max) // what range of sizes the routine is capable of supporting
	Check() error                          // verify that the config is valid, all required fields are set, API keys work, etc
	Init(size display.Size) error          // perform any initialization behavior that is required for further operation in subsequent Update calls
	Update(now time.Time) *Message         // return nil if no update is required, but non-nil messages indicate a change to what the routine displays
	Parameters() []Parameter               // return all fields that are expected to be provided (via JSON) for proper configuration
}

// Parameter represents configurable fields in a given Routine. If you write your own Routines, be meticulous
// about specifying all required parameters and their expected types! This is how the web-ui knows what config it should
// prompt the user to provide for a given routine
type Parameter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Field       string `json:"field"`
	Type        string `json:"type"`
}

type RoutineBase struct {
	Type     RoutineType      `json:"type"`
	Location display.Location `json:"location"`
	Size     display.Size     `json:"size"`
}

type Routine struct {
	RoutineBase
	Routine RoutineIface `json:"routine"`
}

// RoutineJSON is solely for handling unknown routines that are specified via JSON, but have yet to be unmarshalled
// into a concrete, well-defined RoutineIface implementation (via reflection)
type RoutineJSON struct {
	RoutineBase
	Routine json.RawMessage `json:"routine"`
}

func supportsSize(routine RoutineIface, size display.Size) bool {
	mins, maxs := routine.SizeRange()
	return size.Width >= mins.Width && size.Width <= maxs.Width && size.Height >= mins.Height && size.Height <= maxs.Height
}

// AllRoutines is a global record of all the routines the display can support. If you add your own routine, you should
// "register" it by adding it below!
var AllRoutines = map[RoutineType]RoutineIface{
	TEXT:     &TextRoutine{},
	CLOCK:    &ClockRoutine{},
	TIMER:    &TimerRoutine{},
	WEATHER:  &WeatherRoutine{},
	SEQUENCE: &SequenceRoutine{},
}
