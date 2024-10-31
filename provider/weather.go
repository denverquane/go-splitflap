package provider

import (
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"sync"
	"time"
)

type WeatherProvider struct {
	PollRateSecs int    `json:"poll_rate_secs"`
	LocationID   int    `json:"location_id"`
	Units        string `json:"units"`

	kill        chan struct{}
	subscribers map[string]chan any
	subLock     sync.RWMutex
}

func (wp *WeatherProvider) AddSubscriber(s chan any) string {
	id := uuid.NewString()

	wp.subLock.Lock()
	wp.subscribers[id] = s
	wp.subLock.Unlock()

	return id
}

func (wp *WeatherProvider) RemoveSubscriber(id string) {
	wp.subLock.Lock()
	delete(wp.subscribers, id)
	wp.subLock.Unlock()
}

func (wp *WeatherProvider) Start() error {
	apiKey := os.Getenv("OWM_API_KEY")
	if apiKey == "" {
		return errors.New("OWM_API_KEY is not set, can't start weather provider")
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
					// TODO do work

					wp.subLock.RLock()
					for _, v := range wp.subscribers {
						v <- 5
					}
					wp.subLock.RUnlock()
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
