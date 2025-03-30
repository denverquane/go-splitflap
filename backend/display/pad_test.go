package display

import "testing"

func TestLeftPad(t *testing.T) {
	size := Size{Width: 6, Height: 1}
	str := "ABCD"

	newStr := LeftPad(str, size)
	if newStr != "  ABCD" {
		t.Errorf("TestLeftPad got %s, want %s", newStr, "  ABCD")
	}
}

func TestRightPad(t *testing.T) {
	size := Size{Width: 6, Height: 1}
	str := "ABCD"
	newStr := RightPad(str, size)
	if newStr != "ABCD  " {
		t.Errorf("TestRightPad got %s, want %s", newStr, "ABCD  ")
	}
}

func TestLeftPad_multiline(t *testing.T) {
	size := Size{Width: 6, Height: 2}
	str := "ABCD"
	newStr := LeftPad(str, size)
	if newStr != "        ABCD" {
		t.Errorf("TestLeftPad got \"%s\", want \"%s\"", newStr, "      ABCD")
	}
}
