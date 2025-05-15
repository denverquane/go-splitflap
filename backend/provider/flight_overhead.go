package provider

import (
	"context"
	"github.com/navidys/gopensky"
	"log/slog"
	"sync"
	"time"
)

const FLIGHTS_OVERHEAD ProviderType = "FLIGHTS_OVERHEAD"

type FlightsOverheadProvider struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	LatRange  float64 `json:"lat_range"`
	LonRange  float64 `json:"lon_range"`

	pollRateSecs int
	lastRefresh  time.Time
	nextRefresh  time.Time
	current      string
	kill         chan struct{}
	lock         sync.RWMutex
}

func (fo *FlightsOverheadProvider) SetPollRateSecs(rate int) {
	fo.lock.Lock()
	defer fo.lock.Unlock()

	fo.pollRateSecs = rate
	if fo.pollRateSecs < 60 {
		slog.Info("flights_overhead provider poll rate is < 60secs, setting to minimum of 60")
		fo.pollRateSecs = 60
	}
	fo.nextRefresh = fo.lastRefresh.Add(time.Duration(fo.pollRateSecs) * time.Second)
}

func (fo *FlightsOverheadProvider) Start() error {
	conn, err := gopensky.NewConnection(context.Background(), "", "")
	if err != nil {
		return err
	}
	latMin := fo.Latitude - fo.LatRange
	lonMin := fo.Longitude - fo.LatRange
	latMax := fo.Latitude + fo.LatRange
	lonMax := fo.Longitude + fo.LatRange

	bbox := gopensky.NewBoundingBox(latMin, lonMin, latMax, lonMax)

	fo.kill = make(chan struct{})
	// make the next refresh 0 so we refresh immediately
	fo.nextRefresh = time.Time{}

	go func() {
		for {
			select {
			case <-fo.kill:
				slog.Info("flights_overhead provider received kill signal, exiting")
				return
			default:
				now := time.Now()

				fo.lock.RLock()
				refresh := now.After(fo.nextRefresh)
				fo.lock.RUnlock()

				if refresh {
					states, err := gopensky.GetStates(conn, 0, []string{}, bbox, true)
					if err != nil {
						slog.Error("Error getting states from gopensky", "error", err.Error())
					} else {
						slog.Info("Got states from gopensky", "states", states)
					}

					fo.lock.Lock()

					for _, v := range states.States {
						if v.Callsign != nil {
							slog.Info("callsign", "v", *v.Callsign)
						}

						fo.current = v.Icao24
					}

					fo.lastRefresh = now
					fo.nextRefresh = now.Add(time.Second * time.Duration(fo.pollRateSecs))

					fo.lock.Unlock()
				}
				time.Sleep(time.Second)
			}
		}
	}()
	return nil
}

func (fo *FlightsOverheadProvider) Stop() {
	fo.kill <- struct{}{}
}

func (fo *FlightsOverheadProvider) Values() PValues {
	fo.lock.RLock()
	defer fo.lock.RUnlock()

	return PValues{
		"current": fo.current,
	}
}
