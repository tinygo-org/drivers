package flash

import (
	"fmt"
	"time"
)

type SerialNumber uint64

func (sn SerialNumber) String() string {
	return fmt.Sprintf("%8X", uint64(sn))
}

type Device struct {
	transport transport
	attrs     Attrs
}

func (dev *Device) Begin() (err error) {

	dev.transport.begin()

	// TODO: should check JEDEC ID against list of known devices
	/*
		if dev.ID, err = dev.ReadJEDEC(); err != nil {
			return err
		}
		println("JEDEC:", dev.ID.String())
	*/

	// We don't know what state the flash is in so wait for any remaining writes and then reset.

	var s byte // status

	// The write in progress bit should be low.
	for s, err = dev.ReadStatus(); (s & 0x01) > 0; s, err = dev.ReadStatus() {
		if err != nil {
			return err
		}
	}

	// The suspended write/erase bit should be low.
	for s, err = dev.ReadStatus2(); (s & 0x80) > 0; s, err = dev.ReadStatus2() {
		if err != nil {
			return err
		}
	}

	if err = dev.transport.runCommand(CmdEnableReset); err != nil {
		return err
	}
	if err = dev.transport.runCommand(CmdReset); err != nil {
		return err
	}

	// Wait 30us for the reset
	stop := time.Now().UnixNano() + int64(30*time.Microsecond)
	for stop > time.Now().UnixNano() {
	}

	// Speed up to max device frequency
	//_trans->setClockSpeed(_flash_dev->max_clock_speed_mhz*1000000UL);

	if err = dev.transport.runCommand(CmdWriteDisable); err != nil {
		return err
	}

	err = dev.WaitUntilReady()
	return err
}

func (dev *Device) ReadJEDEC() (JedecID, error) {
	jedecID := make([]byte, 3)
	if err := dev.transport.readCommand(CmdReadJedecID, jedecID); err != nil {
		return JedecID{}, err
	}
	return JedecID{jedecID[0], jedecID[1], jedecID[2]}, nil
}

func (dev *Device) ReadSerialNumber() (SerialNumber, error) {
	sn := make([]byte, 12)
	if err := dev.transport.readCommand(0x4B, sn); err != nil {
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
	return dev.transport.readMemory(addr, buf)
}

func (dev *Device) ReadStatus() (status byte, err error) {
	buf := make([]byte, 1)
	err = dev.transport.readCommand(CmdReadStatus, buf)
	return buf[0], err
}

func (dev *Device) ReadStatus2() (status byte, err error) {
	buf := make([]byte, 1)
	err = dev.transport.readCommand(CmdReadStatus2, buf)
	return buf[0], err
}

func (dev *Device) WaitUntilReady() error {
	expire := time.Now().UnixNano() + int64(10*time.Second)
	for s, err := dev.ReadStatus(); (s & 0x03) > 0; s, err = dev.ReadStatus() {
		println("wait until ready status", s)
		if err != nil {
			return err
		}
		if time.Now().UnixNano() > expire {
			return fmt.Errorf("WaitUntilReady expired")
		}
	}
	return nil
}
