package provider

import (
	"errors"
	"github.com/briandowns/openweathermap"
	"log/slog"
	"os"
	"sync"
	"time"
)

const WEATHER ProviderType = "WEATHER"

type WeatherType string

const (
	CURRENT WeatherType = "CURRENT"
	HIGH    WeatherType = "HIGH"
	LOW     WeatherType = "LOW"
)

type WeatherProvider struct {
	PollRateSecs int         `json:"poll_rate_secs"`
	LocationID   int         `json:"location_id"`
	Units        string      `json:"units"`
	Type         WeatherType `json:"type"`

	lastValue float64
	kill      chan struct{}
	lock      sync.RWMutex
}

func (wp *WeatherProvider) Start() error {
	apiKey := os.Getenv("OWM_API_KEY")
	if apiKey == "" {
		return errors.New("OWM_API_KEY is not set, can't start weather provider")
	}
	var err error
	var current *openweathermap.CurrentWeatherData
	var forecast *openweathermap.ForecastWeatherData
	if wp.Type == CURRENT {
		current, err = openweathermap.NewCurrent(wp.Units, "en", apiKey)
	} else {
		forecast, err = openweathermap.NewForecast("5", wp.Units, "en", apiKey)
	}
	if err != nil {
		return err
	}
	wp.kill = make(chan struct{})
	go func() {
		refreshTime := time.Now()
		for {
			select {
			case <-wp.kill:
				slog.Info("weather provider received kill signal, exiting")
				return
			default:
				now := time.Now()
				if now.After(refreshTime) {
					var value float64
					if wp.Type == CURRENT {
						err = current.CurrentByID(wp.LocationID)
						value = current.Main.Temp
					} else {
						err = forecast.DailyByID(wp.LocationID, 1)
						if val, ok := forecast.ForecastWeatherJson.(*openweathermap.Forecast5WeatherData); ok {
							if val.Cnt > 0 && len(val.List) > 0 {
								if wp.Type == HIGH {
									value = val.List[0].Main.TempMax
								} else if wp.Type == LOW {
									value = val.List[0].Main.TempMin
								}
							}
						}
					}
					if err != nil {
						slog.Error(err.Error())
					} else {
						slog.Info("Weather provider reported temp", "Type", wp.Type, "value", value)
						wp.lock.Lock()
						wp.lastValue = value
						wp.lock.Unlock()
					}
					refreshTime = now.Add(time.Second * time.Duration(wp.PollRateSecs))
				}
				time.Sleep(time.Second)
			}
		}
	}()
	return nil
}

func (wp *WeatherProvider) Stop() {
	wp.kill <- struct{}{}
}

func (wp *WeatherProvider) Values() PValues {
	wp.lock.RLock()
	defer wp.lock.RUnlock()

	return PValues{
		"value": wp.lastValue,
		"units": wp.Units,
	}
}
