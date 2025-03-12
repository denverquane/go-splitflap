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

	loc := display.Location{
		X: 0,
		Y: 0,
	}

	m := routine.Message{
		Text: "TEST",
	}

	current = mergeMessageToCurrentText(size, current, loc, m)

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

	loc := display.Location{
		X: 0,
		Y: 1,
	}

	m := routine.Message{
		Text: "TEST",
	}

	current = mergeMessageToCurrentText(size, current, loc, m)
	const expected = "      TEST  "

	if string(current) != expected {
		slog.Info("multiline text merge failure", "current", string(current), "expected", expected)
		t.Fatal("multiline text merging failed")
	}
}

func TestDisplay_ApplyTranslations(t *testing.T) {
	translations := map[rune]rune{
		176: 100, // degree symbol to lowercase d
	}
	str := []rune("°°°°")
	if len(str) != 4 {
		t.Fatal("test translation rune arr length should be 4")
	}
	newStr := applyTranslations(str, translations)
	if string(newStr) != "dddd" {
		t.Fatal("translations not applied")
	}
	if len(newStr) != 4 {
		t.Log(len(newStr))
		t.Fatal("translations changed the length of the string")
	}
}

func TestDisplay_arrangeToLayout(t *testing.T) {
	// 2nd 3 characters should be in reverse order
	layout := []int{
		0, 1, 2, 5, 4, 3,
	}
	text := "012345"

	newText := arrangeToLayout(text, layout)

	if string(newText) != "012543" {
		t.Fatal("arrange to layout failed")
	}
}

func TestDisplay_mergeMessageAndArrange(t *testing.T) {
	size := display.Size{
		Width:  12,
		Height: 2,
	}
	current := initMessage(size)

	m := routine.Message{
		Text: "ABCDEFGHIJKLMNOPQRSTUVWX", //the desired behavior is that we get the text looking like this on the display
		// (we don't care about the layout from the perspective of the routine sending messages)
	}

	current = mergeMessageToCurrentText(size, current, display.Location{X: 0, Y: 0}, m)

	// assuming we start the wiring from 0 and end with 23 (bottom left, going counterclockwise to top left),
	// the letters A-X would end up in the order specified:
	// (M-X are on the bottom row, but the first characters in our wiring. Then, A-L are reversed because our wiring
	// on the top row goes right-to-left)
	const expected = "MNOPQRSTUVWXLKJIHGFEDCBA"

	// Display Layout:
	//23 22 21 20 19 18 17 16 15 14 13 12
	//0  1  2  3  4  5  6  7  8  9  10 11

	layout := []int{
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
	}
	newStr := arrangeToLayout(string(current), layout)

	if newStr != expected {
		t.Error("expected", "\""+expected+"\"", "got", "\""+newStr+"\"")
		t.Fatal("merge and arrange to layout failed")
	}
}

func TestDisplay_StateLayout(t *testing.T) {
	layout := []int{
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
	}
	invLayout := invertLayout(layout)
	stateMsg := "MNOPQRSTUVWXLKJIHGFEDCBA"

	newState := arrangeToLayout(stateMsg, invLayout)

	const expected = "ABCDEFGHIJKLMNOPQRSTUVWX"
	if string(newState) != expected {
		t.Error("expected", "\""+expected+"\"", "got", "\""+string(newState)+"\"")
		t.Fatal("arrange state to layout failed")
	}
}
