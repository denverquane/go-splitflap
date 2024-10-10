package splitflap

import (
	"encoding/json"
	"errors"
	"github.com/denverquane/go-splitflap/routine"
	"log/slog"
	"reflect"
	"sync"
)

type Dashboard struct {
	Routines []routine.Routine `json:"routines"`
	active   bool
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
	} else if d.active {
		return errors.New("cant add a routine to an already active dashboard")
	} else {
		for i, r := range d.Routines {
			if r.Name == rout.Name {
				slog.Info("routine found with that name already; overwriting")
				err := rout.Routine.Check()
				if err != nil {
					return err
				}
				r.Routine = rout.Routine
				d.Routines[i] = r
				return nil
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

func (d *Dashboard) Activate(messageQueue chan<- routine.Message) {
	wg := sync.WaitGroup{}
	for _, rout := range d.Routines {
		go func(r routine.RoutineIface) {
			wg.Add(1)
			r.Start(messageQueue)
			wg.Done()
		}(rout.Routine)
	}
	wg.Wait()
	d.active = true
}

func (d *Dashboard) Deactivate() {
	wg := sync.WaitGroup{}
	for _, rout := range d.Routines {
		go func(r routine.RoutineIface) {
			wg.Add(1)
			r.Stop()
			wg.Done()
		}(rout.Routine)
	}
	wg.Wait()
	d.active = false
}
