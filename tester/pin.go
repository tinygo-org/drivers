package tester

// Pin implements the Pin interface in memory for testing.
type Pin struct {
	c Failer
}

// NewPin returns an Pin mock Pin instance that uses c to flag errors
// if they happen.
func NewPin(c Failer) *Pin {
	return &Pin{
		c: c,
	}
}

func (p *Pin) Get() bool {
	return false
}

func (p *Pin) Set(v bool) {
	return
}

func (p *Pin) High() {

}

func (p *Pin) Low() {

}
