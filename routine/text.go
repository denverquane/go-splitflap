package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"log/slog"
	"time"
)

const TEXT = "TEXT"

type TextRoutine struct {
	Text    string               `json:"text"`
	LocSize display.LocationSize `json:"loc_size"`
	kill    chan struct{}
}

func (t *TextRoutine) LocationSize() display.LocationSize {
	return t.LocSize
}

func (t *TextRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 1, Height: 1}, display.Max{Width: 100, Height: 100}
}

func (t *TextRoutine) Check() error {
	if !SupportsSize(t, t.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	if len(t.Text) > t.LocSize.Width {
		return errors.New("text length exceeds defined routine width")
	}
	return nil
}

func (t *TextRoutine) Start(queue chan<- Message) {
	refreshTime := time.Now()
	t.kill = make(chan struct{})
	go func() {
		slog.Info("Text Routine Started")

		for {
			select {
			case <-t.kill:
				slog.Info("text routine received kill signal, exiting")
				return
			default:
				now := time.Now()
				if now.After(refreshTime) {
					queue <- Message{LocationSize: t.LocSize, Text: t.Text}
					refreshTime = now.Add(time.Duration(60) * time.Minute)
				} else {
					time.Sleep(time.Second)
				}
			}
		}
	}()
}

func (t *TextRoutine) SetState(_ string) {
	return
}

func (t *TextRoutine) Stop() {
	t.kill <- struct{}{}
}
