package splitflap

import (
	"encoding/json"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"testing"
)

func TestDashboard_UnmarshalJSON(t *testing.T) {
	d := Dashboard{Routines: []routine.Routine{{
		Name: "clock",
		Type: routine.CLOCK,
		Routine: &routine.ClockRoutine{
			RemoveLeadingZero: true,
			Military:          true,
			AMPMText:          true,
			LocSize: display.LocationSize{
				Location: display.Location{
					X: 1,
					Y: 2,
				},
				Size: display.Size{
					Width:  3,
					Height: 4,
				},
			},
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

	for _, v := range d.Routines {
		if v.Name == "clock" {
			locSize := v.Routine.LocationSize()
			testLocSize := routine.AllRoutines[routine.CLOCK].LocationSize()
			if locSize.X == testLocSize.X || locSize.Y == testLocSize.Y || locSize.Width == testLocSize.Width || locSize.Height == testLocSize.Height {
				t.Error("expected global routine copy to be different than unmarshalled json, global var issue!")
			}
		}
	}
}

func TestDashboard_relevantStateSubset_simple(t *testing.T) {
	displaySize := display.Size{
		Width:  12,
		Height: 1,
	}
	subset := display.LocationSize{
		Location: display.Location{
			X: 2,
			Y: 0,
		},
		Size: display.Size{
			Width:  4,
			Height: 1,
		},
	}
	state := "ABCDEFGHIJKL"
	newState := relevantStateSubset(displaySize, subset, state)
	if newState != "CDEF" {
		t.Fatal("unexpected result for simple string subset")
	}
}

func TestDashboard_relevantStateSubset_multiline(t *testing.T) {
	displaySize := display.Size{
		Width:  6,
		Height: 2,
	}
	subset := display.LocationSize{
		Location: display.Location{
			X: 1,
			Y: 1,
		},
		Size: display.Size{
			Width:  4,
			Height: 1,
		},
	}
	state := "ABCDEF" + "GHIJKL"
	newState := relevantStateSubset(displaySize, subset, state)
	if newState != "HIJK" {
		t.Fatal("unexpected result for multiline string subset")
	}
}

func TestDashboard_relevantStateSubset_multiline_routine(t *testing.T) {
	displaySize := display.Size{
		Width:  6,
		Height: 2,
	}
	subset := display.LocationSize{
		Location: display.Location{
			X: 1,
			Y: 0,
		},
		Size: display.Size{
			Width:  4,
			Height: 2,
		},
	}
	state := "ABCDEF" + "GHIJKL"
	newState := relevantStateSubset(displaySize, subset, state)
	if newState != "BCDEHIJK" {
		t.Fatal("unexpected result for multiline routine string subset")
	}
}
