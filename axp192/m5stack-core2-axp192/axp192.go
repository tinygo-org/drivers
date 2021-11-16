package axp192

import (
	"time"

	"tinygo.org/x/drivers"
	axp192orig "tinygo.org/x/drivers/axp192"
)

// Device wraps an I2C connection to a AXP192 device.
type Device struct {
	*axp192orig.Device
	LED    Pin
	RST    Pin
	SPK_EN Pin
}

// New creates a new AXP192 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(i2c drivers.I2C) *Device {
	d := axp192orig.New(i2c)

	axp := &Device{
		Device: d,
	}
	axp.LED = Pin{pin: 1, axp: axp}
	axp.SPK_EN = Pin{pin: 2, axp: axp}
	axp.RST = Pin{pin: 4, axp: axp}

	axp.begin()

	return axp
}

type Config struct {
}

// Configure sets up the device for communication
func (d *Device) Configure(config Config) error {
	return d.Device.Configure(axp192orig.Config{})
}

func (d *Device) begin() {
	d.SetVbusIPSOutAccessManagement((d.GetVbusIPSOutAccessManagement() & 0x04) | 0x02)
	d.SetGPIO1Control(d.GetGPIO1Control() & 0xF8)
	d.SetGPIO2Control(d.GetGPIO2Control() & 0xF8)
	d.SetBackupBatteryChargingControl((d.GetBackupBatteryChargingControl() & 0x1C) | 0xA2)
	d.SetESPVoltage(3350)
	d.SetLcdVoltage(3300)
	d.SetLDOVoltage(2, 3300) //Periph power voltage preset (LCD_logic, SD card)
	d.SetLDOVoltage(3, 2000) //Vibrator power voltage preset

	d.SetLDOEnable(2, true)
	d.SetDCDC3(true) // LCD Backlight
	// GPIO4 : LCD Reset
	d.SetGPIO43FunctionControl((d.GetGPIO43FunctionControl() & 0x72) | 0x84)
	// Power On/Off Setting
	d.SetPEKParameterSet(0x4C)
	d.SetADCEnableSet(0xFF)

	d.RST.Low()
	time.Sleep(100 * time.Millisecond)
	d.RST.High()
	time.Sleep(100 * time.Millisecond)
}

// ToggleLED toggles LED connected to AXP192.
func (d *Device) ToggleLED() {
	v := d.GetGPIO20SignalStatus()
	if (v & 0x02) > 0 {
		d.SetGPIO20SignalStatus(v & 0xFD)
	} else {
		d.SetGPIO20SignalStatus(v | 0x02)
	}
}

// SetESPVoltage sets voltage of ESP32.
func (d *Device) SetESPVoltage(voltage uint16) {
	if voltage >= 3000 && voltage <= 3400 {
		d.SetDCVoltage(0, voltage)
	}
}

// SetLcdVoltage sets voltage of LCD.
func (d *Device) SetLcdVoltage(voltage uint16) {
	if voltage >= 2500 && voltage <= 3300 {
		d.SetDCVoltage(2, voltage)
	}
}

// SetDCDC3 enables or disables DCDC3.
func (d *Device) SetDCDC3(State bool) {
	v := d.GetDCDC13LDO23Switch()
	if State == true {
		v = (1 << 1) | v
	} else {
		v = ^(uint8(1) << 1) & v
	}
	d.SetDCDC13LDO23Switch(v)
}

// Pin is a single pin on AXP192.
type Pin struct {
	pin uint8
	axp *Device
}

// High sets this GPIO pin to high.
func (p Pin) High() {
	switch p.pin {
	case 1: // LED
		v := p.axp.GetGPIO20SignalStatus()
		p.axp.SetGPIO20SignalStatus(v | 0x02)
	case 2: // SPK_EN
	case 4: // RST
		v := p.axp.GetGPIO43SignalStatus()
		v |= uint8(0x02)
		p.axp.SetGPIO43SignalStatus(v)
	}
}

// Low sets this GPIO pin to low.
func (p Pin) Low() {
	switch p.pin {
	case 1: // LED
		v := p.axp.GetGPIO20SignalStatus()
		p.axp.SetGPIO20SignalStatus(v & 0xFD)
	case 2: // SPK_EN
	case 4: // RST
		v := p.axp.GetGPIO43SignalStatus()
		v &= ^uint8(0x02)
		p.axp.SetGPIO43SignalStatus(v)
	}
}

// Toggle switches an output pin from low to high or from high to low.
func (p Pin) Toggle() {
	switch p.pin {
	case 1: // LED
		v := p.axp.GetGPIO20SignalStatus()
		if (v & 0x02) == 0 {
			p.axp.SetGPIO20SignalStatus(v | 0x02)
		} else {
			p.axp.SetGPIO20SignalStatus(v & 0xFD)
		}
	case 2: // SPK_EN
	case 4: // RST
		v := p.axp.GetGPIO43SignalStatus()
		if (v & 0x02) == 0 {
			v |= uint8(0x02)
		} else {
			v &= ^uint8(0x02)
		}
		p.axp.SetGPIO43SignalStatus(v)
	}
}
