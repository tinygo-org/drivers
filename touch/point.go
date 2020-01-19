package touch

// Pointer is a device that is capable of reading a single touch point
type Pointer interface {
	ReadTouchPoint() Point
}

// Point represents the result of reading a single touch point from a screen.
// X and Y are the horizontal and vertical coordinates of the touch, while Z
// represents the touch pressure.  In general, client code will want to inspect
// the value of Z to see if it is above some threshold to determine if a touch
// is detected at all.
type Point struct {
	X int
	Y int
	Z int
}
