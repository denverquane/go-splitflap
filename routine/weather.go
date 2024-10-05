package routine

import (
	"errors"
	"fmt"
	"github.com/briandowns/openweathermap"
	"github.com/denverquane/go-splitflap/display"
	"log"
	"time"
)

const WEATHER = "Weather"

type WeatherRoutine struct {
	ApiKey       string
	PollRateSecs int
	Units        string
	ShowUnits    bool
	ShowDegree   bool
	LocationID   int
	LocSize      display.LocationSize
	kill         chan struct{}
}

func (w *WeatherRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 10, Height: 2}
}

func (w *WeatherRoutine) Start(queue chan<- Message) error {
	if !SupportsSize(w, w.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	if w.Units != "F" && w.Units != "C" && w.Units != "K" {
		return errors.New("weather units is not one of C, F, or K")
	}
	go func() {
		waitTime := time.Now()
		for {
			select {
			case <-w.kill:
				return
			default:
				now := time.Now()
				if now.After(waitTime) {
					owm, err := openweathermap.NewCurrent(w.Units, "en", w.ApiKey)
					if err != nil {
						log.Println(err)
					} else {
						err = owm.CurrentByID(w.LocationID)
						if err != nil {
							log.Println(err)
						} else {
							queue <- Message{
								LocationSize: w.LocSize,
								Text:         display.LeftPad(w.formatTemp(owm.Main.Temp), w.LocSize.Size),
							}
						}
					}

					waitTime = now.Add(time.Second * time.Duration(w.PollRateSecs))
				} else {
					time.Sleep(time.Second)
				}
			}
		}
	}()
	return nil
}

func (w *WeatherRoutine) Stop() {
	w.kill <- struct{}{}
}

func (w *WeatherRoutine) formatTemp(val float64) string {
	str := fmt.Sprintf("%2.1f", val)
	if len(str) < w.LocSize.Width {
		if w.ShowDegree {
			str += "°"
		}
	}
	if len(str) < w.LocSize.Width {
		if w.ShowUnits {
			str += w.Units
		}
	}

	return str
}
