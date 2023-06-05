package sx126x

// SX126X radio transceiver has several pins that control
// RF_IN, RF_OUT, NSS, and BUSY.
// This interface allows the creation of struct
// that can drive the RF Switch (Used in Lora RX and Lora Tx)
type RadioController interface {
	Init() error
	SetRfSwitchMode(mode int) error
	SetNss(state bool) error
	WaitWhileBusy() error
	SetupInterrupts(handler func()) error
}
