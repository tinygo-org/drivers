package e220

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestParamsToBytes(t *testing.T) {
	tests := []struct {
		config   Config
		expected []byte
	}{
		// min
		{
			Config{
				0x0000,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x0000,
				0x00,
				nil,
			},
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		// arbitary
		{
			Config{
				0x55AA,
				0x03,
				0x0A,
				0x02,
				0x00,
				0x01,
				0x14,
				0x00,
				0x01,
				0x03,
				0xAA55,
				0x00,
				nil,
			},
			[]byte{0x55, 0xAA, 0x6A, 0x81, 0x14, 0x43, 0xAA, 0x55},
		},
		// max
		{
			Config{
				0xFFFF,
				0xFF,
				0xFF,
				0xFF,
				0xFF,
				0xFF,
				0xFF,
				0xFF,
				0xFF,
				0xFF,
				0xFFFF,
				0x00,
				nil,
			},
			[]byte{0xFF, 0xFF, 0xFF, 0xE3, 0xFF, 0xC7, 0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		got := make([]byte, 8)
		tt.config.paramsToBytes(
			&got,
		)
		if !bytes.Equal(got, tt.expected) {
			t.Errorf("bytes are not equall: want=%02X got=%02X", tt.expected, got)
		}
	}
}

func TestBytesToParams(t *testing.T) {
	tests := []struct {
		expected    Config
		configBytes []byte
	}{
		// min
		{
			Config{
				0x0000,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x0000,
				0x00,
				nil,
			},
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		// arbitary
		{
			Config{
				0x55AA,
				0x03,
				0x0A,
				0x02,
				0x00,
				0x01,
				0x14,
				0x00,
				0x01,
				0x03,
				0xAA55,
				0xAA,
				nil,
			},
			[]byte{0x55, 0xAA, 0x6A, 0x81, 0x14, 0x43, 0xAA, 0x55, 0xAA},
		},
		// max
		{
			Config{
				0xFFFF,
				0x07,
				0x1F,
				0x03,
				0x01,
				0x03,
				0xFF,
				0x01,
				0x01,
				0x07,
				0xFFFF,
				0xFF,
				nil,
			},
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		got := Config{}
		got.bytesToParams(
			tt.configBytes,
		)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("objects are not equall: want=%02X got=%02X", tt.expected, got)
		}
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
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
	tests := []struct {
		name   string
		fields fields
		want   []error
	}{
		{
			"all fields are valid(minimum value)",
			fields{
				0x0000,
				UartSerialPortRate1200Bps,
				AirDataRate15625Bps,
				SubPacket200Bytes,
				RSSIAmbientDisable,
				TransmitPowerUnavailable,
				0,
				RSSIByteDisable,
				TransmitMethodFixed,
				WorCycleSetting500ms,
				0x0000,
				0x00,
				nil,
			},
			[]error{},
		},
		{
			"all fields are valid(maximum value)",
			fields{
				0xFFFF,
				UartSerialPortRate115200Bps,
				AirDataRate2148Bps,
				SubPacket32Bytes,
				RSSIAmbientEnable,
				TransmitPower0Dbm,
				30, // limitaion by band-width configration
				RSSIByteEnable,
				TransmitMethodFixed,
				WorCycleSetting4000ms,
				0xFFFF,
				0xFF,
				nil,
			},
			[]error{},
		},
		{
			"some fields are invalid",
			fields{
				0xFFFF,
				UartSerialPortRate115200Bps + 1,
				AirDataRate2148Bps,
				SubPacket32Bytes,
				RSSIAmbientEnable,
				TransmitPower0Dbm + 1,
				30, // limitaion by band-width configuration
				RSSIByteEnable,
				TransmitMethodFixed,
				WorCycleSetting4000ms + 1,
				0xFFFF,
				0xFF,
				nil,
			},
			[]error{
				fmt.Errorf("%s must be less than or equal to %d, got=%d", "UartSerialPortRate", UartSerialPortRate115200Bps, UartSerialPortRate115200Bps+1),
				fmt.Errorf("%s must be less than or equal to %d, got=%d", "TransmitPower", TransmitPower0Dbm, TransmitPower0Dbm+1),
				fmt.Errorf("%s must be less than or equal to %d, got=%d", "WorCycleSetting", WorCycleSetting4000ms, WorCycleSetting4000ms+1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
				errors:             tt.fields.errors,
			}
			c.Validate()
			got := c.Errors()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("errors are not equall: want=%q got=%q", tt.want, got)
			}
		})
	}
}

func TestConfig_uartSerialPortRateString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{UartSerialPortRate: UartSerialPortRate1200Bps},
			"UartSerialPortRate1200Bps",
		},
		{
			"valid max value",
			fields{UartSerialPortRate: UartSerialPortRate115200Bps},
			"UartSerialPortRate115200Bps",
		},
		{
			"invalid value",
			fields{UartSerialPortRate: UartSerialPortRate115200Bps + 1},
			"UartSerialPortRate: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.uartSerialPortRateString(); got != tt.want {
				t.Errorf("Config.uartSerialPortRateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_airDataRateString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{AirDataRate: AirDataRate15625Bps},
			"AirDataRate15625Bps",
		},
		{
			"valid max value",
			fields{AirDataRate: AirDataRate2148Bps},
			"AirDataRate2148Bps",
		},
		{
			"invalid value",
			fields{AirDataRate: AirDataRate2148Bps + 1},
			"AirDataRate: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.airDataRateString(); got != tt.want {
				t.Errorf("Config.airDataRateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_subPacketString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{SubPacket: SubPacket200Bytes},
			"SubPacket200Bytes",
		},
		{
			"valid max value",
			fields{SubPacket: SubPacket32Bytes},
			"SubPacket32Bytes",
		},
		{
			"invalid value",
			fields{SubPacket: SubPacket32Bytes + 1},
			"SubPacket: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.subPacketString(); got != tt.want {
				t.Errorf("Config.subPacketString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_rssiAmbientString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{RssiAmbient: RSSIAmbientDisable},
			"RSSIAmbientDisable",
		},
		{
			"valid max value",
			fields{RssiAmbient: RSSIAmbientEnable},
			"RSSIAmbientEnable",
		},
		{
			"invalid value",
			fields{RssiAmbient: RSSIAmbientEnable + 1},
			"RSSIAmbient: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.rssiAmbientString(); got != tt.want {
				t.Errorf("Config.rssiAmbientString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_transmitPowerString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{TransmitPower: TransmitPowerUnavailable},
			"TransmitPowerUnavailable",
		},
		{
			"valid max value",
			fields{TransmitPower: TransmitPower0Dbm},
			"TransmitPower0Dbm",
		},
		{
			"invalid value",
			fields{TransmitPower: TransmitPower0Dbm + 1},
			"TransmitPower: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.transmitPowerString(); got != tt.want {
				t.Errorf("Config.transmitPowerString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_rssiByteString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{RssiByte: RSSIByteDisable},
			"RSSIByteDisable",
		},
		{
			"valid max value",
			fields{RssiByte: RSSIByteEnable},
			"RSSIByteEnable",
		},
		{
			"invalid value",
			fields{RssiByte: RSSIByteEnable + 1},
			"RSSIByte: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.rssiByteString(); got != tt.want {
				t.Errorf("Config.rssiByteString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_transmitMethodString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{TransmitMethod: TransmitMethodTransparent},
			"TransmitMethodTransparent",
		},
		{
			"valid max value",
			fields{TransmitMethod: TransmitMethodFixed},
			"TransmitMethodFixed",
		},
		{
			"invalid value",
			fields{TransmitMethod: TransmitMethodFixed + 1},
			"TransmitMethod: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.transmitMethodString(); got != tt.want {
				t.Errorf("Config.transmitMethodString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_worCycleSettingString(t *testing.T) {
	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"valid minimum value",
			fields{WorCycleSetting: WorCycleSetting500ms},
			"WorCycleSetting500ms",
		},
		{
			"valid max value",
			fields{WorCycleSetting: WorCycleSetting4000ms},
			"WorCycleSetting4000ms",
		},
		{
			"invalid value",
			fields{WorCycleSetting: WorCycleSetting4000ms + 1},
			"WorCycleSetting: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ModuleAddr:         tt.fields.ModuleAddr,
				UartSerialPortRate: tt.fields.UartSerialPortRate,
				AirDataRate:        tt.fields.AirDataRate,
				SubPacket:          tt.fields.SubPacket,
				RssiAmbient:        tt.fields.RssiAmbient,
				TransmitPower:      tt.fields.TransmitPower,
				Channel:            tt.fields.Channel,
				RssiByte:           tt.fields.RssiByte,
				TransmitMethod:     tt.fields.TransmitMethod,
				WorCycleSetting:    tt.fields.WorCycleSetting,
				EncryptionKey:      tt.fields.EncryptionKey,
				Version:            tt.fields.Version,
			}
			if got := c.worCycleSettingString(); got != tt.want {
				t.Errorf("Config.worCycleSettingString() = %v, want %v", got, tt.want)
			}
		})
	}
}
