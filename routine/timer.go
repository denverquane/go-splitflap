package routine

import (
	"errors"
	"fmt"
	"github.com/denverquane/go-splitflap/display"
	"strings"
	"time"
)

const TIMER = "Timer"

type TimerRoutine struct {
	End     time.Time
	LocSize display.LocationSize
	kill    chan struct{}
}

func (t *TimerRoutine) LocationSize() display.LocationSize {
	return t.LocSize
}

func (t *TimerRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 5, Height: 1}
}

func (t *TimerRoutine) Start(queue chan<- Message) error {
	if !SupportsSize(t, t.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	go func() {
		start := time.Now()
		for {
			select {
			case <-t.kill:
				return
			default:
				now := time.Now()
				msg := Message{LocationSize: t.LocSize}
				if now.After(t.End) {
					msg.Text = strings.Repeat("g", t.LocSize.Width)
				} else {
					diff := now.Sub(start)
					mins := int(diff.Minutes()) % 60
					secs := int(diff.Seconds()) % 60
					msg.Text = display.LeftPad(fmt.Sprintf("%02d:%02d", mins, secs), t.LocSize.Size)
				}
				queue <- msg
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
	return nil
}

func (t *TimerRoutine) Stop() {
	t.kill <- struct{}{}
}
