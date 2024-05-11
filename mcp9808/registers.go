// Package mcp9808 implements a driver for the MCP9808 High Accuracy I2C Temperature Sensor
//
// Datasheet: https://cdn-shop.adafruit.com/datasheets/MCP9808.pdf
// Module: https://www.adafruit.com/product/1782
package mcp9808

// Constants/addresses used for I2C.
const (
	MCP9808_DEVICE_ID = 0x0400
	MCP9808_MANUF_ID  = 0x0054

	MCP9808_I2CADDR_DEFAULT = 0x18 //default I2C address
	MCP9808_REG_CONFIG      = 0x01 //MCP9808 config register

	MCP9808_REG_CONFIG_SHUTDOWN   = 0x0100 //shutdown config
	MCP9808_REG_CONFIG_CRITLOCKED = 0x0080 //critical trip lock
	MCP9808_REG_CONFIG_WINLOCKED  = 0x0040 //alarm window lock
	MCP9808_REG_CONFIG_INTCLR     = 0x0020 //interrupt clear
	MCP9808_REG_CONFIG_ALERTSTAT  = 0x0010 //alert output status
	MCP9808_REG_CONFIG_ALERTCTRL  = 0x0008 //alert output control
	MCP9808_REG_CONFIG_ALERTSEL   = 0x0004 //alert output select
	MCP9808_REG_CONFIG_ALERTPOL   = 0x0002 //alert output polarity
	MCP9808_REG_CONFIG_ALERTMODE  = 0x0001 //alert output mode

	MCP9808_REG_UPPER_TEMP   = 0x02 //upper alert boundary
	MCP9808_REG_LOWER_TEMP   = 0x03 //lower alert boundery
	MCP9808_REG_CRIT_TEMP    = 0x04 //critical temperature
	MCP9808_REG_AMBIENT_TEMP = 0x05 //ambient temperature
	MCP9808_REG_MANUF_ID     = 0x06 //manufacturer ID
	MCP9808_REG_DEVICE_ID    = 0x07 //device ID
	MCP9808_REG_RESOLUTION   = 0x08 //resolution
)

/*
=======   ============   ==============
value     resolution     reading Time
=======   ============   ==============

	0          0.5°C            30 ms
	1          0.25°C           65 ms
	2         0.125°C          130 ms
	3         0.0625°C         250 ms

=======   ============   ==============
*/
type resolution uint8

const (
	Low resolution = iota
	Medium
	High
	Maximum
)
