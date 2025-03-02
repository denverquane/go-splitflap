package splitflap

import (
	"github.com/denverquane/go-splitflap/display"
	"github.com/denverquane/go-splitflap/routine"
	"log/slog"
	"testing"
)

func TestDisplay_mergeMessageToCurrentText_simple(t *testing.T) {
	size := display.Size{
		Width:  6,
		Height: 1,
	}
	current := initMessage(size)

	m := routine.Message{
		LocationSize: display.LocationSize{
			Location: display.Location{
				X: 0,
				Y: 0,
			},
			//Size: display.Size{
			//	Width:  4,
			//	Height: 1,
			//},
		},
		Text: "TEST",
	}

	current = mergeMessageToCurrentText(size, current, m)

	if string(current) != "TEST  " {
		t.Fatal("simple text merging failed")
	}
}

func TestDisplay_mergeMessageToCurrentText_multiline(t *testing.T) {
	size := display.Size{
		Width:  6,
		Height: 2,
	}
	current := initMessage(size)

	m := routine.Message{
		LocationSize: display.LocationSize{
			Location: display.Location{
				X: 0,
				Y: 1,
			},
			//Size: display.Size{
			//	Width:  4,
			//	Height: 1,
			//},
		},
		Text: "TEST",
	}

	current = mergeMessageToCurrentText(size, current, m)
	const expected = "      TEST  "

	if string(current) != expected {
		slog.Info("multiline text merge failure", "current", string(current), "expected", expected)
		t.Fatal("multiline text merging failed")
	}
}

func TestDisplay_arrangeToLayout(t *testing.T) {
	// 2nd 3 characters should be in reverse order
	layout := []int{
		0, 1, 2, 5, 4, 3,
	}
	text := []byte("012345")

	newText := arrangeToLayout(text, layout)

	if string(newText) != "012543" {
		t.Fatal("arrange to layout failed")
	}
}

func TestDisplay_mergeMessageAndArrange(t *testing.T) {
	size := display.Size{
		Width:  6,
		Height: 2,
	}
	current := initMessage(size)

	m := routine.Message{
		LocationSize: display.LocationSize{
			Location: display.Location{
				X: 0,
				Y: 1,
			},
		},
		Text: "TEST",
	}

	current = mergeMessageToCurrentText(size, current, m)

	// first line is reversed, 2nd line is normal.
	// assumes a layout whose wiring starts in the lower left corner, and wraps counter-clockwise to the 1st line
	layout := []int{
		11, 10, 9, 8, 7, 6, 0, 1, 2, 3, 4, 5,
	}
	current = arrangeToLayout(current, layout)

	// we expect the first line to be backwards, and the 2nd line preserved
	const expected = "  TSET      "
	if string(current) != expected {
		t.Error("expected", "\""+expected+"\"", "got", "\""+string(current)+"\"")
		t.Fatal("merge and arrange to layout failed")
	}
}
