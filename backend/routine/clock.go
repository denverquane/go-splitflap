package routine

import (
	"errors"
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/provider"
	"strings"
	"time"
)

const CLOCK = "CLOCK"

type ClockRoutine struct {
	RemoveLeadingZero bool   `json:"remove_leading_zero"`
	Military          bool   `json:"military"`
	AMPMText          bool   `json:"AMPM_text"`
	Timezone          string `json:"timezone"`

	size       display.Size
	formatStr  string
	tzLoc      *time.Location
	lastUpdate time.Time
}

func (c *ClockRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 5, Height: 1}, display.Max{Width: 100, Height: 1}
}

func (c *ClockRoutine) Check() error {
	if c.Military && c.AMPMText {
		return errors.New("military and ampm text cannot both be set on clock routine simultaneously")
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return err
	}

	return nil
}

func (c *ClockRoutine) Init(size display.Size) error {
	if !supportsSize(c, size) {
		return errors.New("routine does not support that size")
	}

	c.size = size
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return err
	}
	c.tzLoc = loc
	c.formatStr = c.getFormatString()
	// set the last update to 0 so that the first call to Update always renders text
	c.lastUpdate = time.Time{}
	return nil
}

func (c *ClockRoutine) Update(now time.Time, _ provider.ProviderValues) *Message {
	if now.Sub(c.lastUpdate).Seconds() < 1 {
		return nil
	}

	msg := Message{}

	t := now.In(c.tzLoc)

	msg.Text = t.Format(c.formatStr)
	if c.RemoveLeadingZero && strings.HasPrefix(msg.Text, "0") {
		msg.Text = strings.Replace(msg.Text, "0", " ", 1)
	}
	msg.Text = display.LeftPad(msg.Text, c.size)
	c.lastUpdate = now
	return &msg
}

func (c *ClockRoutine) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "Remove Leading Zero",
			Description: "If the time is 04:30, should it be displayed as 4:30 instead",
			Field:       "remove_leading_zero",
			Type:        "bool",
		},
		{
			Name:        "24-hour Format",
			Description: "Use 24-hour time format (military time)",
			Field:       "military",
			Type:        "bool",
		},
		{
			Name:        "Show AM/PM",
			Description: "Show AM/PM text after the time",
			Field:       "AMPM_text",
			Type:        "bool",
		},
		{
			Name:        "Timezone",
			Description: "IANA timezone name (e.g., 'America/New_York', 'Europe/London')",
			Field:       "timezone",
			Type:        "string",
		},
	}
}

func (c *ClockRoutine) GetProviderName() string {
	return ""
}

func (c *ClockRoutine) getFormatString() string {
	if c.Military {
		return "15:04"
	} else {
		format := "03:04"
		if c.AMPMText && c.size.Width-len(format) > 2 {
			format += " PM"
		}
		return format
	}
}
