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
	LocationID int    `json:"location_id"`
	Units      string `json:"units"`

	pollRateSecs int
	lastRefresh  time.Time
	nextRefresh  time.Time
	low, high    float64
	kill         chan struct{}
	lock         sync.RWMutex
}

func (wp *WeatherForecastProvider) SetPollRateSecs(rate int) {
	wp.lock.Lock()
	defer wp.lock.Unlock()

	wp.pollRateSecs = rate
	if wp.pollRateSecs < 60 {
		slog.Info("weather_forecast provider poll rate is < 60secs, setting to minimum of 60")
		wp.pollRateSecs = 60
	}
	wp.nextRefresh = wp.lastRefresh.Add(time.Duration(wp.pollRateSecs) * time.Second)
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
	// make the next refresh 0 so we refresh immediately
	wp.nextRefresh = time.Time{}
	go func() {

		for {
			select {
			case <-wp.kill:
				slog.Info("weather_forecast provider received kill signal, exiting")
				return
			default:
				now := time.Now()

				wp.lock.RLock()
				refresh := now.After(wp.nextRefresh)
				wp.lock.RUnlock()

				if refresh {
					var low, high float64

					err = forecast.DailyByID(wp.LocationID, 1)
					if val, ok := forecast.ForecastWeatherJson.(*openweathermap.Forecast5WeatherData); ok {
						if val.Cnt > 0 && len(val.List) > 0 {
							low = val.List[0].Main.TempMin
							high = val.List[0].Main.TempMax
						}
					}

					wp.lock.Lock()

					if err != nil {
						slog.Error(err.Error())
					} else {
						slog.Info("weather_forecast provider reported temps", "low", low, "high", high)
						wp.low = low
						wp.high = high
					}

					wp.lastRefresh = now
					wp.nextRefresh = now.Add(time.Second * time.Duration(wp.pollRateSecs))

					wp.lock.Unlock()
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
