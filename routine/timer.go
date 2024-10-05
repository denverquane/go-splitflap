package routine

import (
	"errors"
	"fmt"
	"github.com/denverquane/go-splitflap/display"
	"strings"
	"time"
)

const TIMER = "Timer"

type TimerConfig struct {
	End     time.Time
	LocSize display.LocationSize
}

func (t *TimerConfig) GetLocationSize() display.LocationSize {
	return t.LocSize
}

func (t *TimerConfig) SetLocationSize(locSize display.LocationSize) {
	t.LocSize = locSize
}

type TimerRoutine struct {
	config TimerConfig
	kill   chan struct{}
}

func (c *TimerRoutine) Name() RoutineName {
	return TIMER
}

func (c *TimerRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 5, Height: 1}
}

func (c *TimerRoutine) CheckConfig(config RoutineConfigIface) error {
	cfg := config.(*TimerConfig)
	if !SupportsSize(c, cfg.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	return nil
}

func (c *TimerRoutine) SetConfig(config RoutineConfigIface) error {
	if err := c.CheckConfig(config); err != nil {
		return err
	}
	cfg := config.(*TimerConfig)
	c.config = *cfg
	c.kill = make(chan struct{})
	return nil
}

func (c *TimerRoutine) Start(queue chan<- Message) error {
	go timerFunc(queue, c.kill, c.config)
	return nil
}

func (c *TimerRoutine) Stop() {
	c.kill <- struct{}{}
}

func timerFunc(queue chan<- Message, kill <-chan struct{}, cfg TimerConfig) {
	start := time.Now()
	for {
		select {
		case <-kill:
			return
		default:
			t := time.Now()
			msg := Message{LocationSize: cfg.LocSize}
			if t.After(cfg.End) {
				msg.Text = strings.Repeat("g", cfg.LocSize.Width)
			} else {
				diff := t.Sub(start)
				mins := int(diff.Minutes()) % 60
				secs := int(diff.Seconds()) % 60
				msg.Text = display.LeftPad(fmt.Sprintf("%02d:%02d", mins, secs), cfg.LocSize.Size)
			}
			queue <- msg
		}
		time.Sleep(time.Millisecond * 500)
	}
}
