// Package mcp9808 implements a driver for the MCP9808 High Accuracy I2C Temperature Sensor
//
// Datasheet: https://cdn-shop.adafruit.com/datasheets/MCP9808.pdf
// Module: https://www.adafruit.com/product/1782
package mcp9808

// Constants/addresses used for I2C.
const (
	MCP9808_DEFAULT_ADDRESS        = 0x18
	MCP9808_MANUFACTURER_ID        = 0x54 // Value
	MCP9808_REG_CONFIGURATION      = 0x01
	MCP9808_REG_UPPER_TEMP         = 0x02
	MCP9808_REG_LOWER_TEMP         = 0x03
	MCP9808_REG_CRITICAL_TEMP      = 0x04
	MCP9808_REG__TEMP              = 0x05
	MCP9808_REG_MANUFACTURER_ID    = 0x06
	MCP9808_REG_DEVICE_ID          = 0x07
	MCP9808_REG_RESOLUTION         = 0x08
	MCP9808_RESOLUTION_HALF_C      = 0x0
	MCP9808_RESOLUTION_QUARTER_C   = 0x1
	MCP9808_RESOLUTION_EIGHTH_C    = 0x2
	MCP9808_RESOLUTION_SIXTEENTH_C = 0x3
)
