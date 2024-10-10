package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"log/slog"
	"strings"
	"time"
)

const CLOCK = "CLOCK"

type ClockRoutine struct {
	RemoveLeadingZero bool                 `json:"remove_leading_zero"`
	Military          bool                 `json:"military"`
	AMPMText          bool                 `json:"AMPM_text"`
	Timezone          string               `json:"timezone"`
	LocSize           display.LocationSize `json:"loc_size"`
	tzLoc             *time.Location
	kill              chan struct{}
}

func (c *ClockRoutine) LocationSize() display.LocationSize {
	return c.LocSize
}

func (c *ClockRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 8, Height: 1}
}

func (c *ClockRoutine) Check() error {
	if !SupportsSize(c, c.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	if c.Military && c.AMPMText {
		return errors.New("military and ampm text cannot both be set on clock routine simultaneously")
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return err
	}
	return nil
}

func (c *ClockRoutine) Start(queue chan<- Message) {
	c.kill = make(chan struct{})
	if c.Timezone == "" {
		c.Timezone = "Local"
	}
	if loc, err := time.LoadLocation(c.Timezone); err != nil {
		panic(err)
	} else {
		c.tzLoc = loc
	}

	go func() {
		slog.Info("Clock Routine Started")

		formatStr := c.getFormatString()

		for {
			select {
			case <-c.kill:
				slog.Info("clock routine received kill signal, exiting")
				return
			default:
				msg := Message{LocationSize: c.LocSize}

				t := time.Now().In(c.tzLoc)

				msg.Text = t.Format(formatStr)
				if c.RemoveLeadingZero && strings.HasPrefix(msg.Text, "0") {
					msg.Text = strings.Replace(msg.Text, "0", " ", 1)
				}
				msg.Text = display.LeftPad(msg.Text, c.LocSize.Size)
				queue <- msg
				time.Sleep(time.Second)
			}
		}
	}()
}

func (c *ClockRoutine) Stop() {
	c.kill <- struct{}{}
}

func (c *ClockRoutine) getFormatString() string {
	if c.Military {
		return "15:04"
	} else {
		format := "03:04"
		if c.AMPMText && c.LocSize.Width-len(format) > 2 {
			format += " PM"
		}
		return format
	}
}
