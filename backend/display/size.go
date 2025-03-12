package display

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Location struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Min Size
type Max Size

type SizeRange struct {
	Min
	Max
}
