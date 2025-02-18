package tester

type Pin struct {
	c Failer
}

func NewPin(c Failer) *Pin {
	return &Pin{c}
}

type NoopPin struct{}

func NewNoopPin() *NoopPin {
	return &NoopPin{}
}

func (n *NoopPin) Get() bool     { return true }
func (n *NoopPin) High()         {}
func (n *NoopPin) Low()          {}
func (n *NoopPin) Set(high bool) {}
