package routine

import (
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/serdiev/usb_serial"
	"log"
	"log/slog"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

const SLOWTEXT = "SLOWTEXT"

type SlowText struct {
	Text          string       `json:"text"`
	Size_         display.Size `json:"size"`
	LetterDelayMs int          `json:"letter_delay_ms"`
	kill          chan struct{}
	state         string
	stateLock     sync.RWMutex
}

func (s *SlowText) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 1, Height: 1}, display.Max{Width: 100, Height: 100}
}

func (s *SlowText) Size() display.Size {
	return s.Size_
}

func (s *SlowText) Check() error {
	return nil
}

const msPerFlap = 65

type LetterStart struct {
	char  rune
	start int
	pos   int
}

type LetterStarts []LetterStart

func (s *SlowText) Start(queue chan<- Message) {
	s.kill = make(chan struct{})
	s.stateLock = sync.RWMutex{}

	totalLen := s.Size_.Width * s.Size_.Height

	charLen := len(s.Text)
	idx := 0
	empty := strings.Repeat(" ", totalLen)

	letterStarts := createLetterStarts(s.Text, s.LetterDelayMs)
	for _, v := range letterStarts {
		slog.Info("letter", "char", string(v.char), "start", v.start, "pos", v.pos)
	}

	go func() {
		queue <- Message{
			Text: empty,
		}

		s.waitForEmptySequence(empty)
		slog.Info("Empty, ready to begin")

		offset := 0
		for {
			select {
			case <-s.kill:
				return
			default:
				if idx < charLen {
					s.stateLock.RLock()
					// continually send characters that are 1 less than the current state, so it never catches up
					newState := spinLetters(letterStarts, idx+1, []rune(s.state))
					s.stateLock.RUnlock()

					log.Println(newState)
					queue <- Message{
						Text: newState,
					}
					delay := 0
					if idx < charLen-1 {
						delay = letterStarts[idx+1].start - offset
						offset += delay
					}
					slog.Info("sleeping", "duration", delay)
					time.Sleep(time.Millisecond * time.Duration(delay))
					idx++
				} else {
					// TODO spin for a bit before settling the characters finally
					time.Sleep(time.Millisecond * 1000)
					queue <- Message{
						Text: s.Text,
					}
					// do we want to exit entirely?
					return
				}
			}
		}
	}()
}

func (s *SlowText) waitForEmptySequence(empty string) {
	for {
		s.stateLock.RLock()
		checkState := s.state
		s.stateLock.RUnlock()

		if checkState == empty {
			return
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func createLetterStarts(text string, delayMs int) LetterStarts {
	startTimes := make(LetterStarts, len(text))
	minimum := 0
	for i := range startTimes {
		aIdx := usb_serial.AlphabetIndex(rune(text[i]))
		if aIdx == 0 {
			aIdx = len(usb_serial.GlobalAlphabet) - 1
		}
		v := i*delayMs - aIdx*msPerFlap
		if v < minimum {
			minimum = v
		}
		startTimes[i] = LetterStart{
			char:  rune(text[i]),
			start: v,
			pos:   i,
		}
	}
	for i, v := range startTimes {
		v.start += int(math.Abs(float64(minimum)))
		startTimes[i] = v
	}
	sort.Slice(startTimes, func(i, j int) bool {
		return startTimes[i].start < startTimes[j].start
	})
	return startTimes
}

func spinLetters(letters LetterStarts, idx int, state []rune) string {
	for i := 0; i < idx; i++ {
		letterStart := letters[i]
		charIdx := usb_serial.AlphabetIndex(state[letterStart.pos])
		if charIdx == 0 {
			charIdx = len(usb_serial.GlobalAlphabet)
		}
		state[letterStart.pos] = usb_serial.GlobalAlphabet[charIdx-1]
	}
	return string(state)
}

func (s *SlowText) SetState(state string) {
	s.stateLock.Lock()
	s.state = state
	slog.Info("Slow text received state: " + s.state)
	s.stateLock.Unlock()
}

func (s *SlowText) Stop() {
	s.kill <- struct{}{}
}

func (s *SlowText) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "Text",
			Description: "The text content to display",
			Field:       "text",
			Type:        "string",
		},
		{
			Name:        "Letter Delay",
			Description: "The delay in milliseconds between letters appearing",
			Field:       "letter_delay_ms",
			Type:        "int",
		},
	}
}
