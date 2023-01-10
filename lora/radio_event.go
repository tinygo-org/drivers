package lora

const (
	RadioEventRxDone = iota
	RadioEventTxDone
	RadioEventTimeout
	RadioEventWatchdog
	RadioEventCrcError
	RadioEventUnhandled
)

// RadioEvent are used for communicating in the radio Event Channel
type RadioEvent struct {
	EventType int
	IRQStatus uint16
	EventData []byte
}

// NewRadioEvent() returns a new RadioEvent that can be used in the RadioChannel
func NewRadioEvent(eType int, irqStatus uint16, eData []byte) RadioEvent {
	r := RadioEvent{EventType: eType, IRQStatus: irqStatus, EventData: eData}
	return r
}
