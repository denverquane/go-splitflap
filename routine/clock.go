package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"log/slog"
	"strings"
	"time"
)

const CLOCK = "Clock"

type ClockConfig struct {
	RemoveLeadingZero bool
	Military          bool
	Precise           bool
	AMPMText          bool
	LocSize           display.LocationSize
}

func (c *ClockConfig) GetLocationSize() display.LocationSize {
	return c.LocSize
}

func (c *ClockConfig) SetLocationSize(locSize display.LocationSize) {
	c.LocSize = locSize
}

type ClockRoutine struct {
	config ClockConfig
	kill   chan struct{}
}

func (c *ClockRoutine) Name() RoutineName {
	return CLOCK
}

func (c *ClockRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 11, Height: 1}
}

func (c *ClockRoutine) CheckConfig(config RoutineConfigIface) error {
	cfg := config.(*ClockConfig)
	if !SupportsSize(c, cfg.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	c.config = *cfg
	c.kill = make(chan struct{})
	return nil
}

func (c *ClockRoutine) SetConfig(config RoutineConfigIface) error {
	cfg := config.(*ClockConfig)
	if !SupportsSize(c, cfg.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	c.config = *cfg
	c.kill = make(chan struct{})
	return nil
}

func (c *ClockRoutine) Start(queue chan<- Message) error {
	go clockFunc(queue, c.kill, c.config)
	return nil
}

func (c *ClockRoutine) Stop() {
	c.kill <- struct{}{}
}

func clockFunc(queue chan<- Message, kill <-chan struct{}, cfg ClockConfig) {
	slog.Info("Clock Routine Started")

	formatStr := getFormatString(cfg)

	for {
		select {
		case <-kill:
			return
		default:
			msg := Message{LocationSize: cfg.LocSize}
			t := time.Now()

			msg.Text = t.Format(formatStr)
			if cfg.RemoveLeadingZero && strings.HasPrefix(msg.Text, "0") {
				msg.Text = strings.Replace(msg.Text, "0", " ", 1)
			}
			msg.Text = display.LeftPad(msg.Text, cfg.LocSize.Size)
			queue <- msg
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func getFormatString(cfg ClockConfig) string {
	if cfg.Military {
		if cfg.Precise && cfg.GetLocationSize().Width > 7 {
			return "15:04:05"
		} else {
			return "15:04"
		}
	} else {
		format := ""
		if cfg.Precise && cfg.GetLocationSize().Width > 7 {
			format = "03:04:05"
		} else {
			format = "03:04"
		}
		if cfg.AMPMText && cfg.GetLocationSize().Width-len(format) > 2 {
			format += " PM"
		}
		return format
	}
}
