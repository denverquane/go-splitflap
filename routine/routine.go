package routine

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/display"
)

type Message struct {
	display.LocationSize
	Text string
}

type RoutineType string

type RoutineIface interface {
	SizeRange() (display.Min, display.Max)
	LocationSize() display.LocationSize
	Check() error
	Start(queue chan<- Message)
	SetState(state string)
	Stop()
}

type Routine struct {
	Name    string       `json:"name"`
	Type    RoutineType  `json:"type"`
	Routine RoutineIface `json:"routine"`
}

type RoutineJSON struct {
	Name    string          `json:"name"`
	Type    RoutineType     `json:"type"`
	Routine json.RawMessage `json:"routine"`
}

func SupportsSize(routine RoutineIface, size display.Size) bool {
	mins, maxs := routine.SizeRange()
	return size.Width >= mins.Width && size.Width <= maxs.Width && size.Height >= mins.Height && size.Height <= maxs.Height
}

var AllRoutines = map[RoutineType]RoutineIface{
	TEXT:     &TextRoutine{},
	CLOCK:    &ClockRoutine{},
	TIMER:    &TimerRoutine{},
	SLOWTEXT: &SlowText{},
	WEATHER:  &WeatherRoutine{},
}
