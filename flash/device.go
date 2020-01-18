package flash

import (
	"fmt"
	"time"
)

type JedecID [3]byte

func (id *JedecID) Manufacturer() uint8 {
	return id[0]
}

func (id *JedecID) MemoryType() uint8 {
	return id[1]
}

func (id *JedecID) Capacity() uint8 {
	return id[2]
}

func (id *JedecID) String() string {
	return fmt.Sprintf(
		"%2X %2X %2X", id.Manufacturer(), id.MemoryType(), id.Capacity())
}

type SerialNumber uint64

func (sn SerialNumber) String() string {
	return fmt.Sprintf("%8X", uint64(sn))
}

type Device struct {
	Transport *Transport
	ID        JedecID
	SerialNum SerialNumber
}

func (dev *Device) Begin() (err error) {

	if dev.ID, err = dev.ReadJEDEC(); err != nil {
		return err
	}

	// TODO: should check JEDEC ID against list of known devices

	// We don't know what state the flash is in so wait for any remaining writes and then reset.

	var s byte // status

	// The write in progress bit should be low.
	for s, err = dev.ReadStatus(); (s & 0x01) > 0; s, err = dev.ReadStatus() {
		if err != nil {
			return err
		}
	}

	// The suspended write/erase bit should be low.
	for s, err = dev.ReadStatus2(); (s & 0x80) > 0; s, err = dev.ReadStatus() {
		if err != nil {
			return err
		}
	}

	if err = dev.Transport.RunCommand(CmdEnableReset); err != nil {
		return err
	}
	if err = dev.Transport.RunCommand(CmdReset); err != nil {
		return err
	}

	// Wait 30us for the reset
	stop := time.Now().UnixNano() + int64(30*time.Microsecond)
	for stop > time.Now().UnixNano() {
	}

	// Speed up to max device frequency
	//_trans->setClockSpeed(_flash_dev->max_clock_speed_mhz*1000000UL);

	if err = dev.Transport.RunCommand(CmdWriteDisable); err != nil {
		return err
	}

	err = dev.WaitUntilReady()
	return err
}

func (dev *Device) ReadJEDEC() (JedecID, error) {
	jedecID := make([]byte, 3)
	if err := dev.Transport.ReadCommand(CmdReadJedecID, jedecID); err != nil {
		return JedecID{}, err
	}
	return JedecID{jedecID[0], jedecID[1], jedecID[2]}, nil
}

func (dev *Device) ReadSerialNumber() (SerialNumber, error) {
	sn := make([]byte, 12)
	if err := dev.Transport.ReadCommand(0x4B, sn); err != nil {
		return 0, err
	}
	return SerialNumber(uint64(sn[11]) | uint64(sn[10])<<0x8 |
		uint64(sn[9])<<0x10 | uint64(sn[8])<<0x18 | uint64(sn[7])<<0x20 |
		uint64(sn[6])<<0x28 | uint64(sn[5])<<0x30 | uint64(sn[4])<<0x38), nil
}

func (dev *Device) ReadBuffer(addr uint32, buf []byte) error {
	// TODO: check if Begin() was successful
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	return dev.Transport.ReadMemory(addr, buf)
}

func (dev *Device) ReadStatus() (status byte, err error) {
	return dev.Transport.ReadCommandByte(CmdReadStatus)
}

func (dev *Device) ReadStatus2() (status byte, err error) {
	return dev.Transport.ReadCommandByte(CmdReadStatus2)
}

func (dev *Device) WaitUntilReady() error {
	for s, err := dev.ReadStatus(); (s & 0x03) > 0; s, err = dev.ReadStatus() {
		if err != nil {
			return err
		}
	}
	return nil
}
