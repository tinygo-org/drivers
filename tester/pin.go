package tester

import "tinygo.org/x/drivers"

var _ drivers.Pin = (*Pin)(nil)
var _ drivers.Pin = (*NoopPin)(nil)

type pinExpectation struct {
	get   bool
	value bool
}

type Pin struct {
	c            Failer
	expectations []pinExpectation
}

// Expect a Get() call, and return the provided value.
func (p *Pin) ExpectGet(value bool) {
	p.expectations = append(p.expectations, pinExpectation{get: true, value: value})
}

// Expect a Set(bool) call, with the provided value.
// true is High, false is Low
func (p *Pin) ExpectSet(value bool) {
	p.expectations = append(p.expectations, pinExpectation{get: false, value: value})
}

func (p *Pin) Get() bool {
	if len(p.expectations) == 0 {
		p.c.Fatalf("unexpected pin read")
	}
	ex := p.expectations[0]
	if !ex.get {
		p.c.Fatalf("unexpected pin read")
	}
	p.expectations = p.expectations[1:]
	return ex.value
}

func (p *Pin) Set(value bool) {
	if len(p.expectations) == 0 {
		p.c.Fatalf("unexpected pin write")
	}
	ex := p.expectations[0]
	if ex.get {
		p.c.Fatalf("unexpected pin write")
	}
	if ex.value != value {
		p.c.Fatalf("unexpected pin write: got %v, expecting %v", value, ex.value)
	}
	p.expectations = p.expectations[1:]
}

func (p *Pin) High() {
	p.Set(true)
}

func (p *Pin) Low() {
	p.Set(false)
}

func NewPin(c Failer) *Pin {
	return &Pin{c, []pinExpectation{}}
}

// NoopPin is a pin that does nothing, and always returns true for Get()
type NoopPin struct{}

func NewNoopPin() *NoopPin {
	return &NoopPin{}
}

func (n *NoopPin) Get() bool     { return true }
func (n *NoopPin) High()         {}
func (n *NoopPin) Low()          {}
func (n *NoopPin) Set(high bool) {}
