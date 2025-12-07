package bno08x

import (
	"time"
)

// hal implements the hardware abstraction layer for bus communication.
type hal struct {
	device *Device
}

func newHAL(dev *Device) *hal {
	return &hal{
		device: dev,
	}
}

func (h *hal) open() error {
	// HAL is now open and ready for communication
	// Soft reset will be sent after handlers are registered
	return nil
}

func (h *hal) close() {}

func (h *hal) read(target []byte) (int, uint32, error) {
	return h.device.bus.read(target)
}

func (h *hal) write(frame []byte) (int, error) {
	if len(frame) > maxTransferOut {
		return 0, errFrameTooLarge
	}
	err := h.device.bus.write(frame)
	if err != nil {
		return 0, err
	}
	return len(frame), nil
}

func (h *hal) getTimeUs() uint32 {
	return uint32(time.Now().UnixNano() / 1000)
}
