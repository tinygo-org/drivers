// Package axp192 provides a driver for the axp192 I2C Enhanced single Cell
// Li-Battery and Power System Management IC.
//
// http://www.x-powers.com/en.php/Info/product_detail/article_id/29
// Datasheet: https://github.com/m5stack/M5-Schematic/blob/master/Core/AXP192%20Datasheet_v1.1_en_draft_2211.pdf
package axp192 // import "tinygo.org/x/drivers/axp192"

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Error uint8

const (
	ErrInvalidID Error = 0x1
)

func (e Error) Error() string {
	switch e {
	case ErrInvalidID:
		return "Invalid chip ID"
	default:
		return "Unknown error"
	}
}

type Device struct {
	bus     drivers.I2C
	buf     []byte
	Address uint8
}

// New returns AXP192 device for the provided I2C bus using default address.
func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:     i2c,
		buf:     make([]byte, 2),
		Address: Address,
	}
}

type Config struct {
}

// Configure the AXP192 device.
func (d *Device) Configure(config Config) error {
	return nil
}

// ReadPowerSupplyStatus reads power supply status.
func (d *Device) ReadPowerSupplyStatus() uint8 {
	return d.read8bit(RegPowerSupplyStatus)
}

// SetVbusIPSOutAccessManagement sets VBUS-IPSOUT access management.
func (d *Device) SetVbusIPSOutAccessManagement(a uint8) {
	d.write1Byte(RegVbusIPSOutAccessManagement, a)
}

// GetVbusIPSOutAccessManagement gets VBUS-IPSOUT access management.
func (d *Device) GetVbusIPSOutAccessManagement() uint8 {
	return d.read8bit(RegVbusIPSOutAccessManagement)
}

// SetGPIO1Control sets GPIO1 function.
func (d *Device) SetGPIO1Control(a uint8) {
	d.write1Byte(RegGPIO1Control, a)
}

// GetGPIO1Control gets GPIO1 function.
func (d *Device) GetGPIO1Control() uint8 {
	return d.read8bit(RegGPIO1Control)
}

// SetGPIO2Control sets GPIO2 function.
func (d *Device) SetGPIO2Control(a uint8) {
	d.write1Byte(RegGPIO2Control, a)
}

// GetGPIO2Control gets GPIO2 function.
func (d *Device) GetGPIO2Control() uint8 {
	return d.read8bit(RegGPIO2Control)
}

// SetGPIO20SignalStatus sets GPIO[2:0] signal status.
func (d *Device) SetGPIO20SignalStatus(a uint8) {
	d.write1Byte(RegGPIO20SignalStatus, a)
}

// GetGPIO20SignalStatus gets GPIO[2:0] signal status.
func (d *Device) GetGPIO20SignalStatus() uint8 {
	return d.read8bit(RegGPIO20SignalStatus)
}

// SetBackupBatteryChargingControl sets backup battery charge control.
func (d *Device) SetBackupBatteryChargingControl(a uint8) {
	d.write1Byte(RegBackupBatteryChargingControl, a)
}

// GetBackupBatteryChargingControl gets backup battery charge control.
func (d *Device) GetBackupBatteryChargingControl() uint8 {
	return d.read8bit(RegBackupBatteryChargingControl)
}

// SetDCDC1VoltageSet sets DC-DC1 output voltage.
func (d *Device) SetDCDC1VoltageSet(a uint8) {
	d.write1Byte(RegDCDC1VoltageSet, a)
}

// GetDCDC1VoltageSet gets DC-DC1 output voltage.
func (d *Device) GetDCDC1VoltageSet() uint8 {
	return d.read8bit(RegDCDC1VoltageSet)
}

// SetDCDC2VoltageSet sets DC-DC2 dynamic voltage parameter.
func (d *Device) SetDCDC2VoltageSet(a uint8) {
	d.write1Byte(RegDCDC2VoltageSet, a)
}

// GetDCDC2VoltageSet gets DC-DC2 dynamic voltage parameter.
func (d *Device) GetDCDC2VoltageSet() uint8 {
	return d.read8bit(RegDCDC2VoltageSet)
}

// SetDCDC3VoltageSet sets DC-DC3 output voltage.
func (d *Device) SetDCDC3VoltageSet(a uint8) {
	d.write1Byte(RegDCDC3VoltageSet, a)
}

// GetDCDC3VoltageSet gets DC-DC3 output voltage.
func (d *Device) GetDCDC3VoltageSet() uint8 {
	return d.read8bit(RegDCDC3VoltageSet)
}

// SetLDO23VoltageSet sets LDO2/3 output voltage.
func (d *Device) SetLDO23VoltageSet(a uint8) {
	d.write1Byte(RegLDO23VoltageSet, a)
}

// GetLDO23VoltageSet gets LDO2/3 output voltage.
func (d *Device) GetLDO23VoltageSet() uint8 {
	return d.read8bit(RegLDO23VoltageSet)
}

// SetDCDC13LDO23Switch sets DC-DC1/3 & LOD2/3 output control.
func (d *Device) SetDCDC13LDO23Switch(a uint8) {
	d.write1Byte(RegDCDC13LDO23Switch, a)
}

// GetDCDC13LDO23Switch gets DC-DC1/3 & LOD2/3 output control.
func (d *Device) GetDCDC13LDO23Switch() uint8 {
	return d.read8bit(RegDCDC13LDO23Switch)
}

// SetGPIO43FunctionControl sets GPIO[4:3] pin function.
func (d *Device) SetGPIO43FunctionControl(a uint8) {
	d.write1Byte(RegGPIO43FunctionControl, a)
}

// GetGPIO43FunctionControl gets GPIO[4:3] pin function.
func (d *Device) GetGPIO43FunctionControl() uint8 {
	return d.read8bit(RegGPIO43FunctionControl)
}

// SetPEKParameterSet sets PEK press key parameter.
func (d *Device) SetPEKParameterSet(a uint8) {
	d.write1Byte(RegPEKParameterSet, a)
}

// GetPEKParameterSet gets PEK press key parameter.
func (d *Device) GetPEKParameterSet() uint8 {
	return d.read8bit(RegPEKParameterSet)
}

// SetADCEnableSet sets ADC enable 1.
func (d *Device) SetADCEnableSet(a uint8) {
	d.write1Byte(RegADCEnableSet, a)
}

// GetADCEnableSet gets ADC enable 1.
func (d *Device) GetADCEnableSet() uint8 {
	return d.read8bit(RegADCEnableSet)
}

// SetGPIO43SignalStatus sets GPIO[4:3] signal status.
func (d *Device) SetGPIO43SignalStatus(a uint8) {
	d.write1Byte(RegGPIO43SignalStatus, a)
}

// GetGPIO43SignalStatus gets GPIO[4:3] signal status.
func (d *Device) GetGPIO43SignalStatus() uint8 {
	return d.read8bit(RegGPIO43SignalStatus)
}

// SetDCVoltage sets DC voltage.
func (d *Device) SetDCVoltage(number uint8, voltage uint16) {
	if voltage < 700 {
		voltage = 0
	} else {
		voltage = (voltage - 700) / 25
	}

	switch number {
	case 0:
		v := d.GetDCDC1VoltageSet()
		d.SetDCDC1VoltageSet((v & 0x80) | (uint8(voltage) & 0x7F))
	case 1:
		v := d.GetDCDC2VoltageSet()
		d.SetDCDC2VoltageSet((v & 0x80) | (uint8(voltage) & 0x7F))
	case 2:
		v := d.GetDCDC3VoltageSet()
		d.SetDCDC3VoltageSet((v & 0x80) | (uint8(voltage) & 0x7F))
	}
}

// SetLDOVoltage sets LDO voltage.
func (d *Device) SetLDOVoltage(number uint8, voltage uint16) {
	if voltage > 3300 {
		voltage = 15
	} else {
		voltage = (voltage / 100) - 18
	}

	switch number {
	case 2:
		v := d.GetLDO23VoltageSet()
		d.SetLDO23VoltageSet((v & 0x0F) | (uint8(voltage) << 4))
		break
	case 3:
		v := d.GetLDO23VoltageSet()
		d.SetLDO23VoltageSet((v & 0xF0) | uint8(voltage))
		break
	}
}

// SetLDOEnable enable LDO.
func (d *Device) SetLDOEnable(number uint8, state bool) {
	mark := uint8(0x01)
	mark <<= number
	switch number {
	case 2:
		v := d.GetDCDC13LDO23Switch()
		d.SetDCDC13LDO23Switch(v | mark)
	case 3:
		v := d.GetDCDC13LDO23Switch()
		d.SetDCDC13LDO23Switch(v & (^mark))
	}
}

func (d *Device) write1Byte(reg, data uint8) {
	legacy.WriteRegister(d.bus, d.Address, reg, []byte{data})
}

func (d *Device) read8bit(reg uint8) uint8 {
	legacy.ReadRegister(d.bus, d.Address, reg, d.buf[:1])
	return d.buf[0]
}
