//go:build xiao_ble

package lsm6ds3tr

import (
	"device/nrf"
	"machine"
	"time"
)

// Configure sets up the device for communication.
func (d *Device) Configure(cfg Configuration) error {

	// Following lines are XIAO BLE Sense specific, they have nothing to do with sensor per se
	// Implementation adapted from https://github.com/Seeed-Studio/Seeed_Arduino_LSM6DS3/blob/master/LSM6DS3.cpp#L68-L77

	// Special mode for IMU power pin on this board.
	// Can not use pin.Configure() directly due to special mode and 32 bit size
	pinConfig := uint32(nrf.GPIO_PIN_CNF_DIR_Output<<nrf.GPIO_PIN_CNF_DIR_Pos) |
		uint32(nrf.GPIO_PIN_CNF_INPUT_Disconnect<<nrf.GPIO_PIN_CNF_INPUT_Pos) |
		uint32(nrf.GPIO_PIN_CNF_PULL_Disabled<<nrf.GPIO_PIN_CNF_PULL_Pos) |
		uint32(nrf.GPIO_PIN_CNF_DRIVE_H0H1<<nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		uint32(nrf.GPIO_PIN_CNF_SENSE_Disabled<<nrf.GPIO_PIN_CNF_SENSE_Pos)
	nrf.P1.PIN_CNF[8].Set(pinConfig) // LSM_PWR == P1_08

	// Enable IMU
	machine.LSM_PWR.High()

	// Wait a moment
	time.Sleep(10 * time.Millisecond)

	// Common initialisation code
	return d.doConfigure(cfg)
}
