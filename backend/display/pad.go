package display

import (
	"strings"
)

func LeftPad(str string, size Size) string {
	return pad(str, size, true)
}

func RightPad(str string, size Size) string {
	return pad(str, size, false)
}

func pad(str string, size Size, left bool) string {
	diff := size.Width*size.Height - len(str)
	if diff < 1 {
		return str
	}
	p := strings.Repeat(" ", diff)
	if left {
		return p + str
	} else {
		return str + p
	}
}
