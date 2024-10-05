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
	Start(queue chan<- Message) error
	Stop()
}

type Routine struct {
	Name    string
	Type    RoutineType
	Routine RoutineIface
}

type RoutineJSON struct {
	Name    string
	Type    RoutineType
	Routine json.RawMessage
}

func SupportsSize(routine RoutineIface, size display.Size) bool {
	mins, maxs := routine.SizeRange()
	return size.Width >= mins.Width && size.Width <= maxs.Width && size.Height >= mins.Height && size.Height <= maxs.Height
}
