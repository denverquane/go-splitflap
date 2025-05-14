package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"time"
)

const TEXT = "TEXT"

type TextRoutine struct {
	Text string `json:"text"`

	lastUpdate time.Time
}

func (t *TextRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 1, Height: 1}, display.Max{Width: 100, Height: 100}
}

func (t *TextRoutine) Check() error {

	return nil
}

func (t *TextRoutine) Init(size display.Size) error {
	if !supportsSize(t, size) {
		return errors.New("routine does not support that size")
	}
	if len(t.Text) > size.Width*size.Height {
		return errors.New("text length exceeds defined routine size")
	}

	// set the last update to 0 so that the first call to Update always renders text
	t.lastUpdate = time.Time{}
	return nil
}

func (t *TextRoutine) Update(now time.Time, _ provider.ProviderValues) *Message {
	if now.Sub(t.lastUpdate).Seconds() < 1 {
		return nil
	}

	t.lastUpdate = now
	msg := Message{Text: t.Text}
	return &msg
}

func (t *TextRoutine) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "Text",
			Description: "The text content to display",
			Field:       "text",
			Type:        "string",
		},
	}
}
