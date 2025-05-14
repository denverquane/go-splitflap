package routine

import (
	"errors"
	"fmt"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"strings"
	"time"
)

const TIMER = "TIMER"

type TimerRoutine struct {
	End time.Time `json:"end"`

	size       display.Size
	start      time.Time
	lastUpdate time.Time
}

func (t *TimerRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 5, Height: 1}
}

func (t *TimerRoutine) Check() error {
	return nil
}

func (t *TimerRoutine) Init(size display.Size) error {
	if !supportsSize(t, size) {
		return errors.New("routine does not support that size")
	}

	t.size = size
	now := time.Now()
	// set the last update to 0 so that the first call to Update always renders text
	t.lastUpdate = time.Time{}
	t.start = now
	return nil
}

func (t *TimerRoutine) Update(now time.Time, _ provider.ProviderValues) *Message {
	if now.Sub(t.lastUpdate).Seconds() < 1 {
		return nil
	}

	msg := Message{}
	if now.After(t.End) {
		msg.Text = strings.Repeat("g", t.size.Width)
	} else {
		diff := now.Sub(t.start)
		mins := int(diff.Minutes()) % 60
		secs := int(diff.Seconds()) % 60
		msg.Text = display.LeftPad(fmt.Sprintf("%02d:%02d", mins, secs), t.size)
	}
	t.lastUpdate = now
	return &msg
}

func (t *TimerRoutine) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "End Time",
			Description: "The time when the timer should end",
			Field:       "end",
			Type:        "time",
		},
	}
}
