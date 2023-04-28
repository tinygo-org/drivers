// Package e220 implements a driver for the e220 LoRa module.
package e220

import "fmt"

// Config represents E220's configuration parameters
type Config struct {
	ModuleAddr         uint16
	UartSerialPortRate uint8
	AirDataRate        uint8
	SubPacket          uint8
	RssiAmbient        uint8
	TransmitPower      uint8
	Channel            uint8
	RssiByte           uint8
	TransmitMethod     uint8
	WorCycleSetting    uint8
	EncryptionKey      uint16
	Version            uint8
	errors             []error
}

func (c *Config) paramsToBytes(
	bytes *[]byte,
) error {
	// byte [8] is read only
	if len(*bytes) < E220RegisterLength-1 {
		return fmt.Errorf("length of bytes must be greater than or equal to %d: got=%d", E220RegisterLength-1, (*bytes))
	}
	(*bytes)[0] = byte((c.ModuleAddr & 0xFF00) >> 8)
	(*bytes)[1] = byte((c.ModuleAddr & 0x00FF) >> 0)
	(*bytes)[2] = byte(((c.UartSerialPortRate & 0x07) << 5) | (c.AirDataRate & 0x1F))
	reserved := byte(0b000)
	(*bytes)[3] = byte(((c.SubPacket & 0x03) << 6) | ((c.RssiAmbient & 0x01) << 5) | ((reserved & 0x07) << 2) | (c.TransmitPower & 0x03))
	(*bytes)[4] = byte(c.Channel)
	reserved = byte(0b000)
	(*bytes)[5] = byte(((c.RssiByte & 0x01) << 7) | ((c.TransmitMethod & 0x01) << 6) | ((reserved & 0x07) << 3) | (c.WorCycleSetting & 0x07))
	(*bytes)[6] = byte((c.EncryptionKey & 0xFF00) >> 8)
	(*bytes)[7] = byte((c.EncryptionKey & 0x00FF) >> 0)

	return nil
}

func (c *Config) bytesToParams(bytes []byte) error {
	if len(bytes) < E220RegisterLength {
		return fmt.Errorf("length of bytes must be greater than or equal to %d: got=%d", E220RegisterLength, (bytes))
	}
	c.ModuleAddr = uint16((uint16(bytes[0]) << 8) | (uint16(bytes[1]) << 0))
	c.UartSerialPortRate = uint8((bytes[2] & 0xE0) >> 5)
	c.AirDataRate = uint8((bytes[2] & 0x1F) >> 0)
	c.SubPacket = uint8((bytes[3] & 0xC0) >> 6)
	c.RssiAmbient = uint8((bytes[3] & 0x20) >> 5)
	c.TransmitPower = uint8((bytes[3] & 0x03) >> 0)
	c.Channel = bytes[4]
	c.RssiByte = uint8((bytes[5] & 0x80) >> 7)
	c.TransmitMethod = uint8((bytes[5] & 0x40) >> 6)
	c.WorCycleSetting = uint8((bytes[5] & 0x07) >> 0)
	c.EncryptionKey = uint16((uint16(bytes[6]) << 8) | (uint16(bytes[7]) << 0))
	c.Version = bytes[8]

	return nil
}

// Validate validates configuration parameters
// Detected errors can be retrieved with Errors().
func (c *Config) Validate() {
	params := []struct {
		name  string
		value uint8
		limit uint8
	}{
		// ModuleAddr does not have limitation
		{
			"UartSerialPortRate", c.UartSerialPortRate, UartSerialPortRate115200Bps,
		},
		{
			"AirDataRate", c.AirDataRate, AirDataRate2148Bps,
		},
		{
			"SubPacket32Bytes", c.SubPacket, SubPacket32Bytes,
		},
		{
			"RSSIAmbient", c.RssiAmbient, RSSIAmbientEnable,
		},
		{
			"TransmitPower", c.TransmitPower, TransmitPower0Dbm,
		},
		// Channel does not have limitation
		{
			"RSSIByte", c.RssiByte, RSSIByteEnable,
		},
		{
			"TransmitMethod", c.TransmitMethod, TransmitMethodFixed,
		},
		{
			"WorCycleSetting", c.WorCycleSetting, WorCycleSetting4000ms,
		},
		// EncryptionKey does not have limitation
		// Version does not have limitation
	}

	c.errors = make([]error, 0, 8)
	for _, p := range params {
		if err := assertLtOrEq(p.name, p.value, p.limit); err != nil {
			c.errors = append(c.errors, err)
		}
	}

	var bandWidth uint16
	switch c.AirDataRate {
	case AirDataRate15625Bps:
		bandWidth = 125
	case AirDataRate9375Bps:
		bandWidth = 125
	case AirDataRate5469Bps:
		bandWidth = 125
	case AirDataRate3125Bps:
		bandWidth = 125
	case AirDataRate1758Bps:
		bandWidth = 125
	case AirDataRate31250Bps:
		bandWidth = 250
	case AirDataRate18750Bps:
		bandWidth = 250
	case AirDataRate10938Bps:
		bandWidth = 250
	case AirDataRate6250Bps:
		bandWidth = 250
	case AirDataRate3516Bps:
		bandWidth = 250
	case AirDataRate1953Bps:
		bandWidth = 250
	case AirDataRate62500Bps:
		bandWidth = 500
	case AirDataRate37500Bps:
		bandWidth = 500
	case AirDataRate21875Bps:
		bandWidth = 500
	case AirDataRate12500Bps:
		bandWidth = 500
	case AirDataRate7031Bps:
		bandWidth = 500
	case AirDataRate3906Bps:
		bandWidth = 500
	case AirDataRate2148Bps:
		bandWidth = 500
	default:
		// It does not come here because the same check is actually performed in the previous process.
		c.errors = append(c.errors, fmt.Errorf("invalid AirDataRate value: %d", c.AirDataRate))
	}

	switch bandWidth {
	case 125:
		if c.Channel > 37 {
			c.errors = append(
				c.errors,
				fmt.Errorf("if band width is %dKHz, c.Channel must be less than or equal to %d, got=%d", 125, 37, c.Channel),
			)
		}
	case 250:
		if c.Channel > 36 {
			c.errors = append(
				c.errors,
				fmt.Errorf("if band width is %dKHz, c.Channel must be less than or equal to %d, got=%d", 250, 36, c.Channel),
			)
		}
	case 500:
		if c.Channel > 30 {
			c.errors = append(
				c.errors,
				fmt.Errorf("if band width is %dKHz, c.Channel must be less than or equal to %d, got=%d", 500, 30, c.Channel),
			)
		}
	default:
		c.errors = append(c.errors, fmt.Errorf("invalid band width value: %d", bandWidth))
	}
}

func assertLtOrEq(name string, val, limit uint8) error {
	if val > limit {
		return fmt.Errorf("%s must be less than or equal to %d, got=%d", name, limit, val)
	}
	return nil
}

// Errors returns errors of configurations parameters
// Must call Validate() before calling this.
func (c *Config) Errors() []error {
	return c.errors
}

func (c *Config) String() string {
	moduleAddr := fmt.Sprintf("ModuleAddr0x%04X", c.ModuleAddr)
	channel := fmt.Sprintf("Channel%02d", c.Channel)
	uartSerialPortRate := c.uartSerialPortRateString()
	airDataRate := c.airDataRateString()
	subPacket := c.subPacketString()
	rssiAmbient := c.rssiAmbientString()
	transmitPower := c.transmitPowerString()
	rssiByte := c.rssiByteString()
	transmitMethod := c.transmitMethodString()
	worCycleSetting := c.worCycleSettingString()
	encryptionKey := fmt.Sprintf("EncryptionKey0x%04X", c.ModuleAddr)
	version := fmt.Sprintf("Version0x%02X", c.Version)

	return fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
		moduleAddr, channel, uartSerialPortRate, airDataRate,
		subPacket, rssiAmbient, transmitPower, rssiByte,
		transmitMethod, worCycleSetting, encryptionKey, version,
	)
}

func (c *Config) uartSerialPortRateString() string {
	var output string
	switch c.UartSerialPortRate {
	case UartSerialPortRate1200Bps:
		output = "UartSerialPortRate1200Bps"
	case UartSerialPortRate2400Bps:
		output = "UartSerialPortRate2400Bps"
	case UartSerialPortRate4800Bps:
		output = "UartSerialPortRate4800Bps"
	case UartSerialPortRate9600Bps:
		output = "UartSerialPortRate9600Bps"
	case UartSerialPortRate19200Bps:
		output = "UartSerialPortRate19200Bps"
	case UartSerialPortRate38400Bps:
		output = "UartSerialPortRate38400Bps"
	case UartSerialPortRate57600Bps:
		output = "UartSerialPortRate57600Bps"
	case UartSerialPortRate115200Bps:
		output = "UartSerialPortRate115200Bps"
	default:
		output = "UartSerialPortRate: invalid parameter"
	}
	return output
}

func (c *Config) airDataRateString() string {
	var output string
	switch c.AirDataRate {
	case AirDataRate15625Bps:
		output = "AirDataRate15625Bps"
	case AirDataRate9375Bps:
		output = "AirDataRate9375Bps"
	case AirDataRate5469Bps:
		output = "AirDataRate5469Bps"
	case AirDataRate3125Bps:
		output = "AirDataRate3125Bps"
	case AirDataRate1758Bps:
		output = "AirDataRate1758Bps"
	case AirDataRate31250Bps:
		output = "AirDataRate31250Bps"
	case AirDataRate18750Bps:
		output = "AirDataRate18750Bps"
	case AirDataRate10938Bps:
		output = "AirDataRate10938Bps"
	case AirDataRate6250Bps:
		output = "AirDataRate6250Bps"
	case AirDataRate3516Bps:
		output = "AirDataRate3516Bps"
	case AirDataRate1953Bps:
		output = "AirDataRate1953Bps"
	case AirDataRate62500Bps:
		output = "AirDataRate62500Bps"
	case AirDataRate37500Bps:
		output = "AirDataRate37500Bps"
	case AirDataRate21875Bps:
		output = "AirDataRate21875Bps"
	case AirDataRate12500Bps:
		output = "AirDataRate12500Bps"
	case AirDataRate7031Bps:
		output = "AirDataRate7031Bps"
	case AirDataRate3906Bps:
		output = "AirDataRate3906Bps"
	case AirDataRate2148Bps:
		output = "AirDataRate2148Bps"
	default:
		output = "AirDataRate: invalid parameter"
	}
	return output
}

func (c *Config) subPacketString() string {
	var output string
	switch c.SubPacket {
	case SubPacket200Bytes:
		output = "SubPacket200Bytes"
	case SubPacket128Bytes:
		output = "SubPacket128Bytes"
	case SubPacket64Bytes:
		output = "SubPacket64Bytes"
	case SubPacket32Bytes:
		output = "SubPacket32Bytes"
	default:
		output = "SubPacket: invalid parameter"
	}
	return output
}

func (c *Config) rssiAmbientString() string {
	var output string
	switch c.RssiAmbient {
	case RSSIAmbientDisable:
		output = "RSSIAmbientDisable"
	case RSSIAmbientEnable:
		output = "RSSIAmbientEnable"
	default:
		output = "RSSIAmbient: invalid parameter"
	}
	return output
}

func (c *Config) transmitPowerString() string {
	var output string
	switch c.TransmitPower {
	case TransmitPowerUnavailable:
		output = "TransmitPowerUnavailable"
	case TransmitPower13Dbm:
		output = "TransmitPower13Dbm"
	case TransmitPower7Dbm:
		output = "TransmitPower7Dbm"
	case TransmitPower0Dbm:
		output = "TransmitPower0Dbm"
	default:
		output = "TransmitPower: invalid parameter"
	}
	return output
}

func (c *Config) rssiByteString() string {
	var output string
	switch c.RssiByte {
	case RSSIByteDisable:
		output = "RSSIByteDisable"
	case RSSIByteEnable:
		output = "RSSIByteEnable"
	default:
		output = "RSSIByte: invalid parameter"
	}
	return output
}

func (c *Config) transmitMethodString() string {
	var output string
	switch c.TransmitMethod {
	case TransmitMethodTransparent:
		output = "TransmitMethodTransparent"
	case TransmitMethodFixed:
		output = "TransmitMethodFixed"
	default:
		output = "TransmitMethod: invalid parameter"
	}
	return output
}

func (c *Config) worCycleSettingString() string {
	var output string
	switch c.WorCycleSetting {
	case WorCycleSetting500ms:
		output = "WorCycleSetting500ms"
	case WorCycleSetting1000ms:
		output = "WorCycleSetting1000ms"
	case WorCycleSetting1500ms:
		output = "WorCycleSetting1500ms"
	case WorCycleSetting2000ms:
		output = "WorCycleSetting2000ms"
	case WorCycleSetting2500ms:
		output = "WorCycleSetting2500ms"
	case WorCycleSetting3000ms:
		output = "WorCycleSetting3000ms"
	case WorCycleSetting3500ms:
		output = "WorCycleSetting3500ms"
	case WorCycleSetting4000ms:
		output = "WorCycleSetting4000ms"
	default:
		output = "WorCycleSetting: invalid parameter"
	}
	return output
}
