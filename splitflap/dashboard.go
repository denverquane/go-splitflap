package splitflap

import (
	"encoding/json"
	"fmt"
	"github.com/denverquane/go-splitflap/routine"
	"maps"
	"slices"
)

type Dashboard struct {
	RoutineConfig map[routine.RoutineName]routine.RoutineConfigIface
}

func (d *Dashboard) UnmarshalJSON(data []byte) error {
	type Alias Dashboard

	aux := &struct {
		RoutineConfig map[routine.RoutineName]json.RawMessage
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if d.RoutineConfig == nil {
		d.RoutineConfig = make(map[routine.RoutineName]routine.RoutineConfigIface)
	}

	for routineCfgName, rawRoutineCfg := range aux.RoutineConfig {
		switch routineCfgName {
		case routine.CLOCK:
			var Clock routine.ClockConfig
			if err := json.Unmarshal(rawRoutineCfg, &Clock); err != nil {
				return err
			}
			d.RoutineConfig[routineCfgName] = &Clock
		case routine.TIMER:
			var Timer routine.TimerConfig
			if err := json.Unmarshal(rawRoutineCfg, &Timer); err != nil {
				return err
			}
			d.RoutineConfig[routineCfgName] = &Timer
		default:
			return fmt.Errorf("unknown config name, can't unmarshal json: %s", routineCfgName)
		}
	}
	return nil
}

func (d *Dashboard) RoutineNames() []routine.RoutineName {
	return slices.Collect(maps.Keys(d.RoutineConfig))
}
