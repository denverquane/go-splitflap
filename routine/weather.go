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
	PollRateSecs int                  `json:"poll_rate_secs"`
	WeatherType  WeatherType          `json:"weather_type"`
	Units        string               `json:"units"`
	ShowUnits    bool                 `json:"show_units"`
	ShowDegree   bool                 `json:"show_degree"`
	LocationID   int                  `json:"location_id"`
	LocSize      display.LocationSize `json:"loc_size"`
	kill         chan struct{}
}

func (w *WeatherRoutine) LocationSize() display.LocationSize {
	return w.LocSize
}

func (w *WeatherRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 3, Height: 1}, display.Max{Width: 10, Height: 1}
}

func (w *WeatherRoutine) Check() error {
	if !SupportsSize(w, w.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
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

func (w *WeatherRoutine) Start(queue chan<- Message) {
	w.kill = make(chan struct{})
	go func() {
		refreshTime := time.Now()
		for {
			select {
			case <-w.kill:
				slog.Info("weather routine received kill signal, exiting")
				return
			default:
				now := time.Now()
				if now.After(refreshTime) {
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
								temp = owm.Main.Temp
							case HIGH:
								temp = owm.Main.TempMin
							case LOW:
								temp = owm.Main.TempMin
							}
						}
					}

					queue <- Message{
						LocationSize: w.LocSize,
						Text:         display.LeftPad(w.formatTemp(temp), w.LocSize.Size),
					}

					refreshTime = now.Add(time.Second * time.Duration(w.PollRateSecs))
				} else {
					time.Sleep(time.Second)
				}
			}
		}
	}()
}

func (w *WeatherRoutine) Stop() {
	w.kill <- struct{}{}
}

func (w *WeatherRoutine) formatTemp(val float64) string {
	var str string
	if w.LocSize.Width < 5 {
		str = fmt.Sprintf("%.0f", val)
	} else {
		str = fmt.Sprintf("%.1f", val)
	}
	if w.ShowDegree {
		if len(str) < w.LocSize.Width {
			str += "°"
		} else {
			slog.Info("not adding degree symbol to weather because output is full", "string", str, "config width", w.LocSize.Width)
		}
	}
	if w.ShowUnits {
		if len(str) < w.LocSize.Width {
			str += w.Units
		} else {
			slog.Info("not adding units to weather because output is full", "string", str, "config width", w.LocSize.Width)
		}
	}

	return str
}
