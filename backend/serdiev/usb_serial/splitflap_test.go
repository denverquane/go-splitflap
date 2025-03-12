package usb_serial

import (
	"testing"
)

func TestAlphabetDistance(t *testing.T) {
	GlobalAlphabet = []rune("ABCDEF")

	dist := AlphabetDistance('A', 'B')
	if dist != 1 {
		t.Fatal("simple alphabet distance fail")
	}

	dist = AlphabetDistance('F', 'A')
	if dist != 1 {
		t.Fatal("wraparound alphabet distance fail")
	}

}
