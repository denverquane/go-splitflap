package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"time"
)

const SEQUENCE = "SEQUENCE"

type element struct {
	Text     string `json:"text"`
	Duration int    `json:"duration_ms"`
}

type SequenceRoutine struct {
	Sequences []element `json:"sequences"`
	Cycle     bool      `json:"cycle"`

	idx        int
	size       display.Size
	lastUpdate time.Time
}

func (s *SequenceRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 1, Height: 1}, display.Max{Width: 100, Height: 100}
}

func (s *SequenceRoutine) Check() error {
	if len(s.Sequences) == 0 {
		return errors.New("no sequences provided")
	}
	for _, elem := range s.Sequences {
		if elem.Duration < 1 {
			return errors.New("sequence duration should not be less than 10 ms")
		}
	}
	return nil
}

func (s *SequenceRoutine) Init(size display.Size) error {
	if !supportsSize(s, size) {
		return errors.New("routine doesn't support that size")
	}
	for _, elem := range s.Sequences {
		if len(elem.Text) > size.Width*size.Height {
			return errors.New("text length in sequence exceeds defined routine size")
		}
	}

	s.size = size
	s.idx = -1
	s.lastUpdate = time.Time{}
	return nil
}

func (s *SequenceRoutine) Update(now time.Time, _ provider.ProviderValues) *Message {
	if s.idx >= len(s.Sequences) {
		return nil
	}
	// if we haven't sent the first string, send it immediately
	if s.idx == -1 {
		s.idx = 0
		s.lastUpdate = now
		return &Message{
			Text: display.LeftPad(s.Sequences[s.idx].Text, s.size),
		}
	}

	elem := s.Sequences[s.idx]

	// if we have sent the current text, then wait until the duration is done before advancing
	if int(now.Sub(s.lastUpdate).Milliseconds()) < elem.Duration {
		return nil
	}

	s.lastUpdate = now
	s.idx++
	if s.idx >= len(s.Sequences) {
		if !s.Cycle {
			return nil
		} else {
			s.idx = 0
		}
	}

	elem = s.Sequences[s.idx]
	return &Message{
		Text: display.LeftPad(elem.Text, s.size),
	}
}

func (s *SequenceRoutine) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "Sequences",
			Description: "Sequences of text and their respective durations in Milliseconds",
			Field:       "sequences",
			Type:        "{\"text\": string, \"duration\": int}",
		},
		{
			Name:        "Cycle",
			Description: "Should the sequence cycle around after completion",
			Field:       "cycle",
			Type:        "bool",
		},
	}
}
