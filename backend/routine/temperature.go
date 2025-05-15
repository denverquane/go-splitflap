package routine

import (
	"errors"
	"fmt"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"log/slog"
	"math"
	"time"
)

const TEMPERATURE = "TEMPERATURE"

type TemperatureRoutine struct {
	ProviderName  string `json:"provider_name"`
	ProviderValue string `json:"provider_value"`
	ShowUnits     bool   `json:"show_units"`
	ShowDegree    bool   `json:"show_degree"`
	RoundDecimal  bool   `json:"round_decimal"`

	size       display.Size
	lastUpdate time.Time
}

func (w *TemperatureRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 3, Height: 1}, display.Max{Width: 12, Height: 1}
}

func (w *TemperatureRoutine) Check() error {
	// TODO check if the provider is configured, and that the value is provided

	return nil
}

func (w *TemperatureRoutine) Init(size display.Size) error {
	if !supportsSize(w, size) {
		return errors.New("routine does not support that size")
	}

	w.size = size
	// set the last update to 0 so that the first call to Update always renders text
	w.lastUpdate = time.Time{}
	return nil
}

func (w *TemperatureRoutine) Update(now time.Time, values provider.ProviderValues) *Message {
	if int(now.Sub(w.lastUpdate).Seconds()) < 1 {
		return nil
	}

	if weatherVals, ok := values[w.ProviderName]; ok {
		units := weatherVals["units"].(string)
		temp := weatherVals[w.ProviderValue].(float64)

		msg := Message{
			Text: display.LeftPad(w.formatTemp(temp, units), w.size),
		}

		w.lastUpdate = now
		return &msg
	}
	return nil
}

func (w *TemperatureRoutine) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "Provider Name",
			Description: "The name of the provider to subscribe to",
			Field:       "provider_name",
			Type:        "string",
		},
		{
			Name:        "Provider Value",
			Description: "The name of the value that the provider populates, that this routine should then use",
			Field:       "provider_value",
			Type:        "string",
		},
		{
			Name:        "Show Units",
			Description: "Whether to show the temperature unit symbol",
			Field:       "show_units",
			Type:        "bool",
		},
		{
			Name:        "Show Degree Symbol",
			Description: "Whether to show the degree symbol",
			Field:       "show_degree",
			Type:        "bool",
		},
		{
			Name:        "Round Decimal",
			Description: "Should decimals be rounded to the closest whole number",
			Field:       "round_decimal",
			Type:        "bool",
		},
	}
}

func (t *TemperatureRoutine) GetProviderName() string {
	return t.ProviderName
}

func (w *TemperatureRoutine) formatTemp(val float64, units string) string {
	var str string
	if w.RoundDecimal {
		str = fmt.Sprintf("%d", int64(math.Round(val)))
	} else {
		if w.size.Width < 5 {
			str = fmt.Sprintf("%.0f", val)
		} else {
			str = fmt.Sprintf("%.1f", val)
		}
	}

	if w.ShowDegree {
		if len(str) < w.size.Width {
			str += "°"
		} else {
			slog.Info("not adding degree symbol to weather because output is full", "string", str, "config width", w.size.Width)
		}
	}
	if w.ShowUnits {
		if len(str) < w.size.Width {
			str += units
		} else {
			slog.Info("not adding units to weather because output is full", "string", str, "config width", w.size.Width)
		}
	}

	return str
}
