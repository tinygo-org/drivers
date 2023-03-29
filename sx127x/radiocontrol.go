package sx127x

// SX127X radio transceiver has several pins that control NSS,
// and that are signalled when RX or TX operations are completed.
// This interface allows the creation of types that can control this
type RadioController interface {
	Init() error
	SetNss(state bool) error
	SetupInterrupts(handler func()) error
}
