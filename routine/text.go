package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"log/slog"
	"time"
)

const TEXT = "Text"

type TextRoutine struct {
	Text        string
	RefreshSecs int
	LocSize     display.LocationSize
	kill        chan struct{}
}

func (t *TextRoutine) LocationSize() display.LocationSize {
	return t.LocSize
}

func (t *TextRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 1, Height: 1}, display.Max{Width: 100, Height: 100}
}

func (t *TextRoutine) Start(queue chan<- Message) error {
	if !SupportsSize(t, t.LocSize.Size) {
		return errors.New("routine does not support that size")
	}
	if len(t.Text) > t.LocSize.Width {
		return errors.New("text length exceeds defined routine width")
	}
	refreshTime := time.Now()
	go func() {
		slog.Info("Text Routine Started")

		for {
			select {
			case <-t.kill:
				return
			default:
				now := time.Now()
				if now.After(refreshTime) {
					queue <- Message{LocationSize: t.LocSize, Text: t.Text}
					refreshTime = now.Add(time.Duration(t.RefreshSecs) * time.Second)
				}

				time.Sleep(time.Second)
			}
		}
	}()
	return nil
}

func (t *TextRoutine) Stop() {
	t.kill <- struct{}{}
}
