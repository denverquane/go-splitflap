package routine

import "github.com/denverquane/go-splitflap/display"

type Message struct {
	display.LocationSize
	Text string
}

type RoutineName string

type RoutineConfigIface interface {
	SetLocationSize(locAndSize display.LocationSize)
	GetLocationSize() display.LocationSize
}

type RoutineIface interface {
	Name() RoutineName
	SizeRange() (display.Min, display.Max)
	CheckConfig(config RoutineConfigIface) error
	SetConfig(config RoutineConfigIface) error
	Start(queue chan<- Message) error
	Stop()
}

func SupportsSize(routine RoutineIface, size display.Size) bool {
	mins, maxs := routine.SizeRange()
	return size.Width >= mins.Width && size.Width <= maxs.Width && size.Height >= mins.Height && size.Height <= maxs.Height
}
