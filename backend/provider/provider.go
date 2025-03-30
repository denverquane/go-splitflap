package provider

import (
	"encoding/json"
	"errors"
	"reflect"
)

type Providers struct {
	Providers []Provider `json:"providers"`
}

func (p *Providers) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Providers []ProviderJSON `json:"providers"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, v := range aux.Providers {
		if rout, ok := AllProviders[v.Type]; !ok {
			return errors.New("unrecognized routine type")
		} else {
			var newProv ProviderIface
			newProv = reflect.New(reflect.ValueOf(rout).Elem().Type()).Interface().(ProviderIface)
			if err := json.Unmarshal(v.Provider, newProv); err != nil {
				return err
			}
			p.Providers = append(p.Providers, Provider{
				Name:     v.Name,
				Type:     v.Type,
				Provider: newProv,
			})
		}
	}

	return nil
}

func (p *Providers) Start() error {
	for _, v := range p.Providers {
		err := v.Provider.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

type Provider struct {
	Name     string        `json:"name"`
	Type     ProviderType  `json:"type"`
	Provider ProviderIface `json:"provider"`
}

type ProviderJSON struct {
	Name     string          `json:"name"`
	Type     ProviderType    `json:"type"`
	Provider json.RawMessage `json:"provider"`
}

type ProviderIface interface {
	Start() error
	Stop()
}

type ProviderType string

var AllProviders = map[ProviderType]ProviderIface{
	WEATHER: &WeatherProvider{},
}
