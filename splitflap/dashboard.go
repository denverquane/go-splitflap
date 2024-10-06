package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/routine"
	"reflect"
)

type Dashboard struct {
	Routines []routine.Routine `json:"routines"`
}

func (d *Dashboard) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Routines []routine.RoutineJSON `json:"routines"`
	}{}
	d.Routines = []routine.Routine{}

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
			d.Routines = append(d.Routines, routine.Routine{
				Name:    v.Name,
				Type:    v.Type,
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
		for _, r := range d.Routines {
			if r.Name == rout.Name {
				return errors.New("routine with that name already exists in this dashboard")
			}
		}
		err := rout.Routine.Check()
		if err != nil {
			return err
		}
		d.Routines = append(d.Routines, rout)
		return nil
	}
}

func (d *Dashboard) Activate(messageQueue chan<- routine.Message) error {
	for r, rout := range d.Routines {
		err := rout.Routine.Start(messageQueue)
		if err != nil {
			return err
		}
		d.Routines[r] = rout
	}
	return nil
}
