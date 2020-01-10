package touch

type Pointer interface {
	GetTouchPoint() Point
}

type Point struct {
	X int
	Y int
	Z int
}
