package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"github.com/denverquane/go-splitflap/routine"
	"reflect"
	"time"
)

type Dashboard struct {
	Routines []*routine.Routine `json:"routines"`
}

type DashboardMessage struct {
	display.Location
	display.Size
	routine.Message
}

// we want a special unmarshaller so we can take generic Routines (whose types/underlying implementations we aren't certain
// of yet when they're in JSON form) and convert them dynamically once we know the type field
func (d *Dashboard) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Routines []routine.RoutineJSON `json:"routines"`
	}{}
	d.Routines = []*routine.Routine{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, v := range aux.Routines {
		if rout, ok := routine.AllRoutines[v.Type]; !ok {
			return errors.New("unrecognized routine type")
		} else {
			var newRout routine.RoutineIface
			newRout = reflect.New(reflect.ValueOf(rout).Elem().Type()).Interface().(routine.RoutineIface)
			if err := json.Unmarshal(v.Routine, newRout); err != nil {
				return err
			}
			if err := newRout.Check(); err != nil {
				return err
			}

			d.Routines = append(d.Routines, &routine.Routine{
				RoutineBase: routine.RoutineBase{
					Type:     v.Type,
					Location: v.Location,
					Size:     v.Size,
				},
				Routine: newRout,
			})
		}
	}

	return nil
}

func (d *Dashboard) AddRoutine(rout routine.Routine) error {
	if _, ok := routine.AllRoutines[rout.Type]; !ok {
		return errors.New("unrecognized routine type")
	} else {
		err := rout.Routine.Check()
		if err != nil {
			return err
		}
		d.Routines = append(d.Routines, &rout)
		return nil
	}
}

func (d *Dashboard) Init() error {
	for _, rout := range d.Routines {
		err := rout.Routine.Init(rout.Size)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dashboard) Update(now time.Time, values provider.ProviderValues) []DashboardMessage {
	msgs := make([]DashboardMessage, 0)
	for _, rout := range d.Routines {
		msg := rout.Routine.Update(now, values)
		if msg != nil {
			dMsg := DashboardMessage{
				rout.Location,
				rout.Size,
				*msg,
			}
			msgs = append(msgs, dMsg)
		}
	}
	return msgs
}

// relevantStateSubset extracts a selection of the state to send to a routine.
// NOTE not currently used; might be used for something that computes the next state based on current state
func relevantStateSubset(displaySize display.Size, loc display.Location, size display.Size, state string) string {
	newState := ""
	if len(state) < size.Width*size.Height {
		return newState
	}
	for y := range size.Height {
		start := ((loc.Y + y) * displaySize.Width) + loc.X
		end := start + size.Width
		newState += state[start:end]
	}

	return newState
}
