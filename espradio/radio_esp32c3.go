//go:build esp32c3

package espradio

import "device/esp"

func initHardware() error {
	// See:
	// https://github.com/esp-rs/esp-wifi/blob/v0.2.0/esp-wifi/src/common_adapter/common_adapter_esp32c3.rs#L18

	const (
		SYSTEM_WIFIBB_RST       = 1 << 0
		SYSTEM_FE_RST           = 1 << 1
		SYSTEM_WIFIMAC_RST      = 1 << 2
		SYSTEM_BTBB_RST         = 1 << 3  // Bluetooth Baseband
		SYSTEM_BTMAC_RST        = 1 << 4  // deprecated
		SYSTEM_RW_BTMAC_RST     = 1 << 9  // Bluetooth MAC
		SYSTEM_RW_BTMAC_REG_RST = 1 << 11 // Bluetooth MAC Regsiters
		SYSTEM_BTBB_REG_RST     = 1 << 13 // Bluetooth Baseband Registers
	)

	const MODEM_RESET_FIELD_WHEN_PU = SYSTEM_WIFIBB_RST |
		SYSTEM_FE_RST |
		SYSTEM_WIFIMAC_RST |
		SYSTEM_BTBB_RST |
		SYSTEM_BTMAC_RST |
		SYSTEM_RW_BTMAC_RST |
		SYSTEM_RW_BTMAC_REG_RST |
		SYSTEM_BTBB_REG_RST

	esp.RTC_CNTL.DIG_PWC.ClearBits(esp.RTC_CNTL_DIG_PWC_WIFI_FORCE_PD)
	esp.APB_CTRL.WIFI_RST_EN.SetBits(MODEM_RESET_FIELD_WHEN_PU)
	esp.APB_CTRL.WIFI_RST_EN.ClearBits(MODEM_RESET_FIELD_WHEN_PU)
	esp.RTC_CNTL.DIG_ISO.ClearBits(esp.RTC_CNTL_DIG_ISO_FORCE_OFF)

	return nil
}

// This is the value used for the ESP32-C3, see:
// https://github.com/esp-rs/esp-wifi/blob/v0.2.0/esp-wifi/src/timer/riscv.rs#L28
const ticksPerSecond = 16_000_000
