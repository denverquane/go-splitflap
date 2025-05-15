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
	LocationID int    `json:"location_id"`
	Units      string `json:"units"`

	pollRateSecs int
	lastRefresh  time.Time
	nextRefresh  time.Time
	current      float64
	kill         chan struct{}
	lock         sync.RWMutex
}

func (wp *WeatherCurrentProvider) SetPollRateSecs(rate int) {
	wp.lock.Lock()
	defer wp.lock.Unlock()

	wp.pollRateSecs = rate
	if wp.pollRateSecs < 60 {
		slog.Info("weather_current provider poll rate is < 60secs, setting to minimum of 60")
		wp.pollRateSecs = 60
	}
	wp.nextRefresh = wp.lastRefresh.Add(time.Duration(wp.pollRateSecs) * time.Second)
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
	// make the next refresh 0 so we refresh immediately
	wp.nextRefresh = time.Time{}
	go func() {
		for {
			select {
			case <-wp.kill:
				slog.Info("weather_current provider received kill signal, exiting")
				return
			default:
				now := time.Now()

				wp.lock.RLock()
				refresh := now.After(wp.nextRefresh)
				wp.lock.RUnlock()

				if refresh {
					var cur float64
					err = current.CurrentByID(wp.LocationID)
					cur = current.Main.Temp

					wp.lock.Lock()

					if err != nil {
						slog.Error(err.Error())
					} else {
						slog.Info("weather_current provider reported temps", "current", cur, "units", wp.Units)
						wp.current = cur
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
