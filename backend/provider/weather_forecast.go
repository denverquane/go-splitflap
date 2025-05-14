package provider

import (
	"errors"
	"github.com/briandowns/openweathermap"
	"log/slog"
	"os"
	"sync"
	"time"
)

const WEATHER_FORECAST ProviderType = "WEATHER_FORECAST"

type WeatherForecastProvider struct {
	PollRateSecs int    `json:"poll_rate_secs"`
	LocationID   int    `json:"location_id"`
	Units        string `json:"units"`

	low, high float64
	kill      chan struct{}
	lock      sync.RWMutex
}

func (wp *WeatherForecastProvider) Start() error {
	apiKey := os.Getenv("OWM_API_KEY")
	if apiKey == "" {
		return errors.New("OWM_API_KEY is not set, can't start weather provider")
	}

	forecast, err := openweathermap.NewForecast("5", wp.Units, "en", apiKey)
	if err != nil {
		return err
	}
	wp.kill = make(chan struct{})
	go func() {
		refreshTime := time.Now()
		for {
			select {
			case <-wp.kill:
				slog.Info("weather forecast provider received kill signal, exiting")
				return
			default:
				now := time.Now()
				if now.After(refreshTime) {
					var low, high float64

					err = forecast.DailyByID(wp.LocationID, 1)
					if val, ok := forecast.ForecastWeatherJson.(*openweathermap.Forecast5WeatherData); ok {
						if val.Cnt > 0 && len(val.List) > 0 {
							low = val.List[0].Main.TempMin
							high = val.List[0].Main.TempMax
						}
					}

					if err != nil {
						slog.Error(err.Error())
					} else {
						slog.Info("Weather forecast provider reported temps", "low", low, "high", high)

						wp.lock.Lock()
						wp.low = low
						wp.high = high
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

func (wp *WeatherForecastProvider) Stop() {
	wp.kill <- struct{}{}
}

func (wp *WeatherForecastProvider) Values() PValues {
	wp.lock.RLock()
	defer wp.lock.RUnlock()

	return PValues{
		"low":   wp.low,
		"high":  wp.high,
		"units": wp.Units,
	}
}
