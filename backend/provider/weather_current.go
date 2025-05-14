package provider

import (
	"errors"
	"github.com/briandowns/openweathermap"
	"log/slog"
	"os"
	"sync"
	"time"
)

const WEATHER_CURRENT ProviderType = "WEATHER_CURRENT"

type WeatherCurrentProvider struct {
	PollRateSecs int    `json:"poll_rate_secs"`
	LocationID   int    `json:"location_id"`
	Units        string `json:"units"`

	current float64
	kill    chan struct{}
	lock    sync.RWMutex
}

func (wp *WeatherCurrentProvider) Start() error {
	apiKey := os.Getenv("OWM_API_KEY")
	if apiKey == "" {
		return errors.New("OWM_API_KEY is not set, can't start weather provider")
	}

	current, err := openweathermap.NewCurrent(wp.Units, "en", apiKey)

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
					var cur float64
					err = current.CurrentByID(wp.LocationID)
					cur = current.Main.Temp
					if err != nil {
						slog.Error(err.Error())
					} else {
						slog.Info("Current weather provider reported temps", "current", cur)

						wp.lock.Lock()
						wp.current = cur
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

func (wp *WeatherCurrentProvider) Stop() {
	wp.kill <- struct{}{}
}

func (wp *WeatherCurrentProvider) Values() PValues {
	wp.lock.RLock()
	defer wp.lock.RUnlock()

	return PValues{
		"current": wp.current,
		"units":   wp.Units,
	}
}
