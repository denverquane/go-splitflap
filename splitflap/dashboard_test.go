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
			Precise:           true,
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
