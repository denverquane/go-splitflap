package routine

import (
	"errors"
	"fmt"
	"github.com/briandowns/openweathermap"
	"github.com/denverquane/go-splitflap/display"
	"log"
	"log/slog"
	"os"
	"time"
)

const WEATHER = "WEATHER"

type WeatherType int

const (
	CURRENT WeatherType = iota
	HIGH
	LOW
)

type WeatherRoutine struct {
	PollRateSecs int         `json:"poll_rate_secs"`
	WeatherType  WeatherType `json:"weather_type"`
	Units        string      `json:"units"`
	ShowUnits    bool        `json:"show_units"`
	ShowDegree   bool        `json:"show_degree"`
	LocationID   int         `json:"location_id"`

	size       display.Size
	lastUpdate time.Time
}

func (w *WeatherRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 3, Height: 1}, display.Max{Width: 12, Height: 1}
}

func (w *WeatherRoutine) Check() error {
	if w.Units != "F" && w.Units != "C" && w.Units != "K" {
		return errors.New("weather units is not one of C, F, or K")
	}
	if w.PollRateSecs < 1 {
		return errors.New("poll_rate_secs cannot be < 1")
	}
	if w.LocationID == 0 {
		return errors.New("location_id was not provided")
	}
	if w.WeatherType != CURRENT && w.WeatherType != HIGH && w.WeatherType != LOW {
		return errors.New("weather_type was not recognized")
	}
	// todo check API key
	return nil
}

func (w *WeatherRoutine) Init(size display.Size) error {
	if !supportsSize(w, size) {
		return errors.New("routine does not support that size")
	}

	w.size = size
	// set the last update to 0 so that the first call to Update always renders text
	w.lastUpdate = time.Time{}
	return nil
}

func (w *WeatherRoutine) Update(now time.Time) *Message {
	if int(now.Sub(w.lastUpdate).Seconds()) < w.PollRateSecs {
		return nil
	}

	var temp float64
	// TODO ideally we'd have some sort of singleton "provider" of data, and then individual
	// routines that can reach out for that cached data and format/display it...
	owm, err := openweathermap.NewCurrent(w.Units, "en", os.Getenv("OWM_API_KEY"))
	if err != nil {
		slog.Error(err.Error())
	} else {
		err = owm.CurrentByID(w.LocationID)
		if err != nil {
			log.Println(err)
		} else {
			switch w.WeatherType {
			case CURRENT:
				// TODO move to weather provider?
			case HIGH:
			case LOW:
				temp = owm.Main.Temp
			}
		}
	}

	msg := Message{
		Text: display.LeftPad(w.formatTemp(temp), w.size),
	}

	w.lastUpdate = now
	return &msg
}

func (w *WeatherRoutine) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "Location ID",
			Description: "OpenWeatherMap location ID",
			Field:       "location_id",
			Type:        "int",
		},
		{
			Name:        "Weather Type",
			Description: "Type of weather data to display (0: Current, 1: High, 2: Low)",
			Field:       "weather_type",
			Type:        "int",
		},
		{
			Name:        "Units",
			Description: "Temperature units (F, C, or K)",
			Field:       "units",
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
			Name:        "Poll Rate",
			Description: "How often to update the weather data (in seconds)",
			Field:       "poll_rate_secs",
			Type:        "int",
		},
	}
}

func (w *WeatherRoutine) formatTemp(val float64) string {
	var str string
	if w.size.Width < 5 {
		str = fmt.Sprintf("%.0f", val)
	} else {
		str = fmt.Sprintf("%.1f", val)
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
			str += w.Units
		} else {
			slog.Info("not adding units to weather because output is full", "string", str, "config width", w.size.Width)
		}
	}

	return str
}
