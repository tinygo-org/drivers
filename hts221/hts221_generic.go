//go:build !nano_33_ble

package hts221

// Configure sets up the HTS221 device for communication.
func (d *Device) Configure() {
	// read calibration data
	d.calibration()
	// activate device and use block data update mode
	d.Power(true)
}
