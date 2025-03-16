package drivers

// Pin is a digital pin interface. It is notably implemented by the
// machine.Pin type.
type Pin interface {
	Get() bool
	High()
	Low()
	Set(high bool)
}
