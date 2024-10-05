package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/routine"
)

type Dashboard struct {
	Routines map[string]routine.Routine
}

func (d *Dashboard) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Routines map[string]routine.RoutineJSON
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, v := range aux.Routines {
		switch v.Type {
		case routine.CLOCK:
			var clock routine.ClockRoutine
			if err := json.Unmarshal(v.Routine, &clock); err != nil {
				return err
			}
			d.Routines[v.Name] = routine.Routine{
				Name:    v.Name,
				Type:    routine.CLOCK,
				Routine: &clock,
			}
		}
	}

	return nil
}

func (d *Dashboard) AddRoutine(rout routine.Routine) error {
	if _, ok := AllRoutines[rout.Type]; !ok {
		return errors.New("unrecognized routine type")
	} else {
		d.Routines[rout.Name] = rout
		return nil
	}
}

func (d *Dashboard) Activate(messageQueue chan<- routine.Message) error {
	for _, rout := range d.Routines {
		err := rout.Routine.Start(messageQueue)
		if err != nil {
			return err
		}
	}
	return nil
}
