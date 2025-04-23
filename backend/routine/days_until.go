package routine

import (
	"errors"
	"fmt"
	"github.com/denverquane/go-splitflap/display"
	"time"
)

const DAYSUNTIL = "DAYSUNTIL"

type DaysUntilRoutine struct {
	End string `json:"end_date"`

	endDate    time.Time
	size       display.Size
	lastUpdate time.Time
}

func (d *DaysUntilRoutine) SizeRange() (display.Min, display.Max) {
	return display.Min{Width: 1, Height: 1}, display.Max{Width: 100, Height: 1}
}

func (d *DaysUntilRoutine) Check() error {
	return nil
}

func (d *DaysUntilRoutine) Init(size display.Size) error {
	if !supportsSize(d, size) {
		return errors.New("routine does not support that size")
	}
	end, err := time.Parse("01/02/2006", d.End)
	if err != nil {
		return err
	}

	d.size = size
	d.lastUpdate = time.Time{}
	d.endDate = end

	return nil
}

func (d *DaysUntilRoutine) Update(now time.Time) *Message {
	if now.Sub(d.lastUpdate) < time.Minute {
		return nil
	}
	d.lastUpdate = now

	days := d.endDate.Sub(now).Hours() / 24.0

	m := Message{
		Text: fmt.Sprintf("%d", int(days)),
	}
	return &m
}

func (d *DaysUntilRoutine) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "End Date",
			Description: "The end date the routine is counting down to, in MM/DD/YYYY format",
			Field:       "end_date",
			Type:        "string",
		},
	}
}
