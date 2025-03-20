package drivers

type Pin interface {
	Get() bool
	High()
	Low()
	Set(value bool)
}
