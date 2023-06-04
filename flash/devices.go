package flash

import "time"

// A DeviceIdentifier can be passed to the Configure() method of a flash Device
// in order provide a means of discovery of device-specific attributes based on
// the JEDEC ID read from the device.
type DeviceIdentifier interface {
	// Identify returns an Attrs struct based on the provided JEDEC ID
	Identify(id JedecID) Attrs
}

// DeviceIdentifierFunc is a functional Identifier implementation
type DeviceIdentifierFunc func(id JedecID) Attrs

// Identify implements the Identifier interface
func (fn DeviceIdentifierFunc) Identify(id JedecID) Attrs {
	return fn(id)
}

// DefaultDeviceIndentifier is a DeviceIdentifier that is capable of recognizing
// JEDEC IDs for all of the known memory devices in this package. If you are
// have no way to be sure about the type of memory device that might be on a
// board you are targeting, this can be a good starting point to use.  The
// downside of using this function is that it will prevent the compiler from
// being able to mark any of the functions for the various devices as unused,
// resulting in larger code size.  If code size is a concern, and if you know
// ahead of time you are only dealing with a limited set of memory devices, it
// might be worthwhile to use your own implementation of a DeviceIdentifier
// that only references those devices, so that more methods are marked unused.
var DefaultDeviceIdentifier = DeviceIdentifierFunc(func(id JedecID) Attrs {
	switch id.Uint32() {
	case 0x010617:
		return S25FL064L()
	case 0x014015:
		return S25FL216K()
	case 0x1F4501:
		return AT25DF081A()
	case 0x856015:
		return P25Q16H()
	case 0xC22015:
		return MX25L1606()
	case 0xC22016:
		return MX25L3233F()
	case 0xC22817:
		return MX25R6435F()
	case 0xC84015:
		return GD25Q16C()
	case 0xC84017:
		return GD25Q64C()
	case 0xEF4015:
		return W25Q16JVIQ()
	case 0xEF4016:
		return W25Q32FV()
	case 0xEF4017:
		return W25Q64JVIQ()
	case 0xEF4018:
		return W25Q128JVSQ()
	case 0xEF6014:
		return W25Q80DL()
	case 0xEF6015:
		return W25Q16FW()
	case 0xEF6016:
		return W25Q32BV()
	case 0xEF7015:
		return W25Q16JVIM()
	case 0xEF7016:
		return W25Q32JVIM()
	case 0xEF7017:
		return W25Q64JVIM()
	case 0xEF7018:
		return W25Q128JVPM()
	default:
		return Attrs{JedecID: id}
	}
})

// Settings for Puya Semi P25Q16H 16MiB QSPI flash.
// Datasheet: https://www.puyasemi.com/uploadfiles/2018/08/20180807152503253.pdf
func P25Q16H() Attrs {
	return Attrs{
		TotalSize:           1 << 24, // 16 MiB
		StartUp:             12 * time.Microsecond,
		JedecID:             JedecID{0x85, 0x60, 0x15},
		MaxClockSpeedMHz:    104,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    true,
	}
}

// Settings for the Cypress (was Spansion) S25FL064L 8MiB SPI flash.
// Datasheet: http://www.cypress.com/file/316661/download
func S25FL064L() Attrs {
	return Attrs{
		TotalSize:           1 << 23, // 8 MiB
		StartUp:             300 * time.Microsecond,
		JedecID:             JedecID{0x01, 0x60, 0x17},
		MaxClockSpeedMHz:    108,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Cypress (was Spansion) S25FL116K 2MiB SPI flash.
// Datasheet: http://www.cypress.com/file/196886/download
func S25FL116K() Attrs {
	return Attrs{
		TotalSize:           1 << 21, // 2 MiB
		StartUp:             10000 * time.Microsecond,
		JedecID:             JedecID{0x01, 0x40, 0x15},
		MaxClockSpeedMHz:    108,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  false,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Cypress (was Spansion) S25FL216K 2MiB SPI flash.
// Datasheet: http://www.cypress.com/file/197346/download
func S25FL216K() Attrs {
	return Attrs{
		TotalSize:           1 << 21, // 2 MiB
		StartUp:             10000 * time.Microsecond,
		JedecID:             JedecID{0x01, 0x40, 0x15},
		MaxClockSpeedMHz:    65,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  false,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Adesto Tech AT25DF081A 1MiB SPI flash. Its on the SAMD21
// Xplained board.
// Datasheet: https://www.adestotech.com/wp-content/uploads/doc8715.pdf
func AT25DF081A() Attrs {
	return Attrs{
		TotalSize:           1 << 20, // 1 MiB
		StartUp:             10000 * time.Microsecond,
		JedecID:             JedecID{0x1F, 0x45, 0x01},
		MaxClockSpeedMHz:    85,
		QuadEnableBitMask:   0x00,
		HasSectorProtection: true,
		SupportsFastRead:    true,
		SupportsQSPI:        false,
		SupportsQSPIWrites:  false,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Macronix MX25L1606 2MiB SPI flash.
// Datasheet:
func MX25L1606() Attrs {
	return Attrs{
		TotalSize:           1 << 21, // 2 MiB,
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xC2, 0x20, 0x15},
		MaxClockSpeedMHz:    8,
		QuadEnableBitMask:   0x40,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    true,
	}
}

// Settings for the Macronix MX25L3233F 4MiB SPI flash.
// Datasheet:
// http://www.macronix.com/Lists/Datasheet/Attachments/7426/MX25L3233F,%203V,%2032Mb,%20v1.6.pdf
func MX25L3233F() Attrs {
	return Attrs{
		TotalSize:           1 << 22, // 4 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xC2, 0x20, 0x16},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x40,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Macronix MX25R6435F 8MiB SPI flash.
// Datasheet:
// http://www.macronix.com/Lists/Datasheet/Attachments/7428/MX25R6435F,%20Wide%20Range,%2064Mb,%20v1.4.pdf
// By default its in lower power mode which can only do 8mhz. In high power mode
// it can do 80mhz.
func MX25R6435F() Attrs {
	return Attrs{
		TotalSize:           1 << 23, // 8 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xC2, 0x28, 0x17},
		MaxClockSpeedMHz:    8,
		QuadEnableBitMask:   0x40,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    true,
	}
}

// Settings for the Gigadevice GD25Q16C 2MiB SPI flash.
// Datasheet: http://www.gigadevice.com/datasheet/gd25q16c/
func GD25Q16C() Attrs {
	return Attrs{
		TotalSize:          1 << 21, // 2 MiB
		StartUp:            5000 * time.Microsecond,
		JedecID:            JedecID{0xC8, 0x40, 0x15},
		MaxClockSpeedMHz:   104,
		QuadEnableBitMask:  0x02,
		SupportsFastRead:   true,
		SupportsQSPI:       true,
		SupportsQSPIWrites: true,
		WriteStatusSplit:   false,
		SingleStatusByte:   false,
	}
}

// Settings for the Gigadevice GD25Q64C 8MiB SPI flash.
// Datasheet: http://www.elm-tech.com/en/products/spi-flash-memory/gd25q64/gd25q64.pdf
func GD25Q64C() Attrs {
	return Attrs{
		TotalSize:           1 << 23, // 8 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xC8, 0x40, 0x17},
		MaxClockSpeedMHz:    104,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    true,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q16JV-IQ 2MiB SPI flash. Note that JV-IM has a
// different .memory_type (0x70) Datasheet:
// https://www.winbond.com/resource-files/w25q16jv%20spi%20revf%2005092017.pdf
func W25Q16JVIQ() Attrs {
	return Attrs{
		TotalSize:           1 << 21, // 2 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x40, 0x15},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q16FW 2MiB SPI flash.
// Datasheet:
// https://www.winbond.com/resource-files/w25q16fw%20revj%2005182017%20sfdp.pdf
func W25Q16FW() Attrs {
	return Attrs{
		TotalSize:           1 << 21, // 2 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x60, 0x15},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q16JV-IM 2MiB SPI flash. Note that JV-IQ has a
// different .memory_type (0x40) Datasheet:
// https://www.winbond.com/resource-files/w25q16jv%20spi%20revf%2005092017.pdf
func W25Q16JVIM() Attrs {
	return Attrs{
		TotalSize:           1 << 21, // 2 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x70, 0x15},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q32BV 4MiB SPI flash.
// Datasheet:
// https://www.winbond.com/resource-files/w25q32bv_revi_100413_wo_automotive.pdf
func W25Q32BV() Attrs {
	return Attrs{
		TotalSize:           1 << 22, // 4 MiB
		StartUp:             10000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x60, 0x16},
		MaxClockSpeedMHz:    104,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  false,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q32JV-IM 4MiB SPI flash.
// Datasheet:
// https://www.winbond.com/resource-files/w25q32jv%20revg%2003272018%20plus.pdf
func W25Q32JVIM() Attrs {
	return Attrs{
		TotalSize:           1 << 22, // 4 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x70, 0x16},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q64JV-IM 8MiB SPI flash. Note that JV-IQ has a
// different .memory_type (0x40) Datasheet:
// http://www.winbond.com/resource-files/w25q64jv%20revj%2003272018%20plus.pdf
func W25Q64JVIM() Attrs {
	return Attrs{
		TotalSize:           1 << 23, // 8 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x70, 0x17},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q64JV-IQ 8MiB SPI flash. Note that JV-IM has a
// different .memory_type (0x70) Datasheet:
// http://www.winbond.com/resource-files/w25q64jv%20revj%2003272018%20plus.pdf
func W25Q64JVIQ() Attrs {
	return Attrs{
		TotalSize:           1 << 23, // 8 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x40, 0x17},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q80DL 1MiB SPI flash.
// Datasheet:
// https://www.winbond.com/resource-files/w25q80dv%20dl_revh_10022015.pdf
func W25Q80DL() Attrs {
	return Attrs{
		TotalSize:           1 << 20, // 1 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x60, 0x14},
		MaxClockSpeedMHz:    104,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  false,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q128JV-SQ 16MiB SPI flash. Note that JV-IM has a
// different .memory_type (0x70) Datasheet:
// https://www.winbond.com/resource-files/w25q128jv%20revf%2003272018%20plus.pdf
func W25Q128JVSQ() Attrs {
	return Attrs{
		TotalSize:           1 << 24, // 16 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x40, 0x18},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q128JV-PM 16MiB SPI flash. Note that JV-IM has a
// different .memory_type (0x70) Datasheet:
// https://www.winbond.com/resource-files/w25q128jv%20revf%2003272018%20plus.pdf
func W25Q128JVPM() Attrs {
	return Attrs{
		TotalSize:           1 << 24, // 16 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x70, 0x18},
		MaxClockSpeedMHz:    133,
		QuadEnableBitMask:   0x02,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        true,
		SupportsQSPIWrites:  true,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}

// Settings for the Winbond W25Q32FV 4MiB SPI flash.
// Datasheet:http://www.winbond.com/resource-files/w25q32fv%20revj%2006032016.pdf?__locale=en
func W25Q32FV() Attrs {
	return Attrs{
		TotalSize:           1 << 22, // 4 MiB
		StartUp:             5000 * time.Microsecond,
		JedecID:             JedecID{0xEF, 0x40, 0x16},
		MaxClockSpeedMHz:    104,
		QuadEnableBitMask:   0x00,
		HasSectorProtection: false,
		SupportsFastRead:    true,
		SupportsQSPI:        false,
		SupportsQSPIWrites:  false,
		WriteStatusSplit:    false,
		SingleStatusByte:    false,
	}
}
