package routine

import (
	"github.com/denverquane/go-splitflap/display"
	"strings"
	"time"
)

const SLOWTEXT = "SLOWTEXT"

type SlowText struct {
	Text           string               `json:"text"`
	LocSize        display.LocationSize `json:"loc_size"`
	InitialDelayMs int                  `json:"initial_delay_ms"`
	LetterDelayMs  int                  `json:"letter_delay_ms"`
	kill           chan struct{}
}

func (s *SlowText) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 1, Height: 1}, display.Max{Width: 100, Height: 100}
}

func (s *SlowText) LocationSize() display.LocationSize {
	return s.LocSize
}

func (s *SlowText) Check() error {
	return nil
}

func (s *SlowText) Start(queue chan<- Message) {
	s.kill = make(chan struct{})
	totalLen := s.LocSize.Width * s.LocSize.Height
	charLen := len(s.Text)
	idx := 1
	go func() {
		queue <- Message{
			LocationSize: s.LocSize,
			Text:         strings.Repeat(" ", totalLen),
		}
		time.Sleep(time.Duration(s.InitialDelayMs) * time.Millisecond)
		for {
			select {
			case <-s.kill:
				return
			default:
				if idx <= charLen {
					msg := strings.Repeat(string(display.HOLD_CHAR), idx-1) + s.Text[idx-1:idx] + strings.Repeat(string(display.HOLD_CHAR), totalLen-idx)
					queue <- Message{
						LocationSize: s.LocSize,
						Text:         msg,
					}
					idx++
					time.Sleep(time.Duration(s.LetterDelayMs) * time.Millisecond)
				}
			}
		}
	}()
}

func (s *SlowText) Stop() {
	s.kill <- struct{}{}
}
