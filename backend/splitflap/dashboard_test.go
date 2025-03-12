package splitflap

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"testing"
)

func TestDashboard_UnmarshalJSON(t *testing.T) {
	d := Dashboard{Routines: []*routine.Routine{{
		RoutineBase: routine.RoutineBase{
			Name: "clock",
			Type: routine.CLOCK,
			Location: display.Location{
				X: 1,
				Y: 2,
			},
			Size: display.Size{
				Width:  3,
				Height: 4,
			},
		},

		Routine: &routine.ClockRoutine{
			RemoveLeadingZero: true,
			Military:          true,
			AMPMText:          true,
		},
	},
	}}
	jsonbytes, err := json.Marshal(d)
	if err != nil {
		t.Error(err)
	}
	err = d.UnmarshalJSON(jsonbytes)
	if err != nil {
		t.Error(err)
	}

	size := d.Routines[0].Size
	loc := d.Routines[0].Location
	if size.Height != 4 || size.Width != 3 {
		t.Error("Wrong size")
	}
	if loc.X != 1 || loc.Y != 2 {
		t.Error("Wrong location")
	}
}

func TestDashboard_relevantStateSubset_simple(t *testing.T) {
	displaySize := display.Size{
		Width:  12,
		Height: 1,
	}
	loc := display.Location{
		X: 2,
		Y: 0,
	}

	size := display.Size{
		Width:  4,
		Height: 1,
	}
	state := "ABCDEFGHIJKL"
	newState := relevantStateSubset(displaySize, loc, size, state)
	if newState != "CDEF" {
		t.Fatal("unexpected result for simple string subset")
	}
}

func TestDashboard_relevantStateSubset_multiline(t *testing.T) {
	displaySize := display.Size{
		Width:  6,
		Height: 2,
	}
	loc := display.Location{
		X: 1,
		Y: 1,
	}
	size := display.Size{
		Width:  4,
		Height: 1,
	}

	state := "ABCDEF" + "GHIJKL"
	newState := relevantStateSubset(displaySize, loc, size, state)
	if newState != "HIJK" {
		t.Fatal("unexpected result for multiline string subset")
	}
}

func TestDashboard_relevantStateSubset_multiline_routine(t *testing.T) {
	displaySize := display.Size{
		Width:  6,
		Height: 2,
	}
	loc := display.Location{
		X: 1,
		Y: 0,
	}
	size := display.Size{
		Width:  4,
		Height: 2,
	}

	state := "ABCDEF" + "GHIJKL"
	newState := relevantStateSubset(displaySize, loc, size, state)
	if newState != "BCDEHIJK" {
		t.Fatal("unexpected result for multiline routine string subset")
	}
}
