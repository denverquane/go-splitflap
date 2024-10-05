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
	d.Routines = make(map[string]routine.Routine)

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, v := range aux.Routines {
		if rout, ok := routine.AllRoutines[v.Type]; !ok {
			return errors.New("unrecognized routine type")
		} else {
			if err := json.Unmarshal(v.Routine, rout); err != nil {
				return err
			}
			d.Routines[v.Name] = routine.Routine{
				Name:    v.Name,
				Type:    v.Type,
				Routine: rout,
			}
		}
	}

	return nil
}

func (d *Dashboard) AddRoutine(rout routine.Routine) error {
	if _, ok := routine.AllRoutines[rout.Type]; !ok {
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
