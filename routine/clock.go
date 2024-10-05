package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"log/slog"
	"strings"
	"time"
)

const CLOCK = "Clock"

type ClockRoutine struct {
	RemoveLeadingZero bool
	Military          bool
	Precise           bool
	AMPMText          bool
	LocSize           display.LocationSize
	kill              chan struct{}
}

func (c *ClockRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 11, Height: 1}
}

func (c *ClockRoutine) Start(queue chan<- Message) error {
	if !SupportsSize(c, c.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	go func() {
		slog.Info("Clock Routine Started")

		formatStr := c.getFormatString()

		for {
			select {
			case <-c.kill:
				return
			default:
				msg := Message{LocationSize: c.LocSize}
				t := time.Now()

				msg.Text = t.Format(formatStr)
				if c.RemoveLeadingZero && strings.HasPrefix(msg.Text, "0") {
					msg.Text = strings.Replace(msg.Text, "0", " ", 1)
				}
				msg.Text = display.LeftPad(msg.Text, c.LocSize.Size)
				queue <- msg
				time.Sleep(time.Millisecond * 500)
			}
		}
	}()
	return nil
}

func (c *ClockRoutine) Stop() {
	c.kill <- struct{}{}
}

func (c *ClockRoutine) getFormatString() string {
	if c.Military {
		if c.Precise && c.LocSize.Width > 7 {
			return "15:04:05"
		} else {
			return "15:04"
		}
	} else {
		format := ""
		if c.Precise && c.LocSize.Width > 7 {
			format = "03:04:05"
		} else {
			format = "03:04"
		}
		if c.AMPMText && c.LocSize.Width-len(format) > 2 {
			format += " PM"
		}
		return format
	}
}
