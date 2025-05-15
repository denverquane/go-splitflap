package provider

import (
	"encoding/json"
	"errors"
	"reflect"
)

type ProviderJSON struct {
	Type                   ProviderType    `json:"type"`
	ActivePollRateSecs     int             `json:"active_poll_rate_secs"`     // when a routine is currently active that uses values from this provider, how often should it poll/update
	BackgroundPollRateSecs int             `json:"background_poll_rate_secs"` // when no routines are currently using values from this provider
	Provider               json.RawMessage `json:"config"`
}

type Provider struct {
	Type                   ProviderType `json:"type"`
	ActivePollRateSecs     int          `json:"active_poll_rate_secs"`
	BackgroundPollRateSecs int          `json:"background_poll_rate_secs"`
	Provider               iface        `json:"config"`
}

// iface is the interface that any new providers should conform to. It should be able to stop and start, and provide
// any data via the Values call
type iface interface {
	Start() error
	SetPollRateSecs(int)
	Stop()
	Values() PValues
}

type ProviderType string

// simple key/value pairing of any values a given provider supplies
type PValues map[string]any

// ProviderValues is a mapping of Provider names (there can be multiple instances of a given provider type) to values
// that the provider supplies
type ProviderValues map[string]PValues

func (p *Provider) UnmarshalJSON(data []byte) error {
	aux := ProviderJSON{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if rout, ok := AllProviders[aux.Type]; !ok {
		return errors.New("unrecognized routine type")
	} else {
		var newProv iface
		newProv = reflect.New(reflect.ValueOf(rout).Elem().Type()).Interface().(iface)
		if err := json.Unmarshal(aux.Provider, newProv); err != nil {
			return err
		}
		p.ActivePollRateSecs = aux.ActivePollRateSecs
		p.BackgroundPollRateSecs = aux.BackgroundPollRateSecs
		p.Type = aux.Type
		p.Provider = newProv
	}

	return nil
}

var AllProviders = map[ProviderType]iface{
	WEATHER_CURRENT:  &WeatherCurrentProvider{},
	WEATHER_FORECAST: &WeatherForecastProvider{},
	FLIGHTS_OVERHEAD: &FlightsOverheadProvider{},
}
