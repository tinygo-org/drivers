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

func (p *Pin) ExpectGet(high bool) {
	p.expectations = append(p.expectations, pinExpectation{get: true, value: high})
}

func (p *Pin) ExpectSet(high bool) {
	p.expectations = append(p.expectations, pinExpectation{get: false, value: high})
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

func (p *Pin) Set(high bool) {
	if len(p.expectations) == 0 {
		p.c.Fatalf("unexpected pin write")
	}
	ex := p.expectations[0]
	if ex.get {
		p.c.Fatalf("unexpected pin write")
	}
	if ex.value != high {
		p.c.Fatalf("unexpected pin write: got %v, expecting %v", high, ex.value)
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

type NoopPin struct{}

func NewNoopPin() *NoopPin {
	return &NoopPin{}
}

func (n *NoopPin) Get() bool     { return true }
func (n *NoopPin) High()         {}
func (n *NoopPin) Low()          {}
func (n *NoopPin) Set(high bool) {}
