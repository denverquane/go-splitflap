package display

type Size struct {
	Width  int
	Height int
}

type Location struct {
	X int
	Y int
}

type Min Size
type Max Size

type LocationSize struct {
	Location
	Size
}
