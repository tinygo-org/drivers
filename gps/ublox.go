package gps

import (
	"time"
)

// FlightModeCmd is a UBX-CFG-NAV5 command to set the GPS into
// flight mode (airborne <1g)
var flightModeCmd = CfgNav5{
	Mask:                  CfgNav5Dyn | CfgNav5MinEl | CfgNav5PosFixMode,
	DynModel:              DynModeAirborne1g, // Airborne with <1g acceleration
	FixMode:               FixModeAuto,       // Auto 2D/3D
	MinElev_deg:           5,                 // Minimum elevation 5 degrees
	FixedAlt_me2:          0,                 // Not used
	FixedAltVar_m2e4:      0,                 // Not used
	PDop:                  100,               // 10.0
	TDop:                  100,               // 10.0
	PAcc_m:                5000,              // 5 meters
	TAcc_m:                5000,              // 5 meters
	StaticHoldThresh_cm_s: 0,                 // Not used
	DgnssTimeout_s:        0,                 // Not used
	CnoThreshNumSVs:       0,                 // Not used
	CnoThresh_dbhz:        0,                 // Not used
	StaticHoldMaxDist_m:   0,                 // Not used
	UtcStandard:           0,                 // Automatic
	Reserved1:             [2]byte{},
	Reserved2:             [5]byte{},
}

// SetFlightMode sends UBX-CFG-NAV5 command to set GPS into flight mode
func (d *Device) SetFlightMode() (err error) {
	flightModeCmd.Put42Bytes(d.buffer[:])
	return d.SendCommand(d.buffer[:42])
}

var (
	// GGA (time, lat/lng, altitude)
	messageRateGGACmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x00,
		Rate:     1, // Every position fix
	}
	// GLL (time, lat/lng)
	messageRateGLLCmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x01,
		Rate:     0, // Disabled
	}
	// GSA (satellite id list)
	messageRateGSACmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x02,
		Rate:     1, // Every position fix
	}
	// GSV (satellite locations)
	messageRateGSVCmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x03,
		Rate:     1, // Every position fix
	}
	// RMC (time, lat/lng, speed, course)
	messageRateRMCCmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x04,
		Rate:     1, // Every position fix
	}
	// VTG (speed, course)
	messageRateVTGCmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x05,
		Rate:     0, // Disabled
	}
	// ZDA (time, timezone)
	messageRateZDACmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x08,
		Rate:     0, // Disabled
	}
	// TXT (text transmission)
	messageRateTXTCmd = CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x41,
		Rate:     0, // Disabled
	}
)

// SetMessageRatesMinimal configures the GPS to output a minimal set of NMEA sentences
func SetMessageRatesMinimal(d *Device) (err error) {
	commands := []CfgMsg1{
		messageRateGSACmd,
		messageRateGGACmd,
		messageRateGLLCmd,
		messageRateGSVCmd,
		messageRateRMCCmd,
		messageRateVTGCmd,
		messageRateZDACmd,
		messageRateTXTCmd,
	}
	return setCfg1s(d, commands)
}

// SetMessageRatesAllEnabled configures the GPS to output all NMEA sentences
func SetMessageRatesAllEnabled(d *Device) (err error) {
	commands := []CfgMsg1{
		messageRateGSACmd,
		messageRateGGACmd,
		messageRateGLLCmd,
		messageRateGSVCmd,
		messageRateRMCCmd,
		messageRateVTGCmd,
		messageRateZDACmd,
		messageRateTXTCmd,
	}
	return setCfg1s(d, commands)
}

func setCfg1s(d *Device, commands []CfgMsg1) (err error) {
	var buf [9]byte
	for _, cmd := range commands {
		cmd.Put9Bytes(buf[:9])
		if err = d.SendCommand(buf[:9]); err != nil {
			return err
		}
	}
	return nil
}

// gnssDisableCmd is a UBX-CFG-GNSS command to disable all GNSS but GPS
// Needed for MAX8's, not needed for MAX7
var gnssDisableCmd = CfgGnss{
	MsgVer:      0x00,
	NumTrkChHw:  0x20, // 32 channels
	NumTrkChUse: 0x20,
	ConfigBlocks: []CfgGnssConfigBlocksType{
		{GnssId: 0, ResTrkCh: 8, MaxTrkCh: 16, Flags: CfgGnssEnable | 0x010000}, // GPS enabled
		{GnssId: 1, ResTrkCh: 1, MaxTrkCh: 3, Flags: 0x010000},                  // SBAS disabled
		{GnssId: 3, ResTrkCh: 8, MaxTrkCh: 16, Flags: 0x010000},                 // BeiDou disabled
		{GnssId: 5, ResTrkCh: 0, MaxTrkCh: 3, Flags: 0x010000},                  // QZSS disabled
		{GnssId: 6, ResTrkCh: 8, MaxTrkCh: 14, Flags: 0x010000},                 // GLONASS disabled
	},
}

// SetGNSSDisable sends UBX-CFG-GNSS command to disable all GNSS but GPS
func (d *Device) SetGNSSDisable() (err error) {
	err = gnssDisableCmd.Put(d.buffer[:])
	if err != nil {
		return err
	}
	return d.SendCommand(d.buffer[:])
}

// SendCommand sends a UBX command and waits for ACK/NAK response
func (d *Device) SendCommand(command []byte) error {
	// Calculate and append checksum
	checksummed := appendChecksum(command)
	d.WriteBytes(checksummed)

	start := time.Now()
	for time.Since(start) < time.Second {
		// Look for UBX sync sequence
		if d.readNextByte() != ubxSyncChar1 {
			continue
		}
		if d.readNextByte() != ubxSyncChar2 {
			continue
		}

		// Read message class and ID
		msgClass := d.readNextByte()
		msgID := d.readNextByte()

		// Check if it's an ACK class message
		if msgClass != ubxClassACK {
			continue
		}

		// Read length (2 bytes, little-endian) - ACK is always 2 bytes payload
		lenLo := d.readNextByte()
		lenHi := d.readNextByte()
		length := uint16(lenLo) | uint16(lenHi)<<8

		if length != 2 {
			continue
		}

		// Read ACK payload: class and ID of acknowledged message
		ackClass := d.readNextByte()
		ackID := d.readNextByte()

		// Verify ACK is for our command (command[2] = class, command[3] = ID)
		if ackClass != command[2] || ackID != command[3] {
			continue
		}

		if msgID == ubxACK_ACK {
			return nil
		}
		if msgID == ubxACK_NAK {
			return errGPSCommandRejected
		}
	}

	return errNoACKToGPSCommand
}

// appendChecksum calculates UBX checksum and appends it to the message
func appendChecksum(msg []byte) []byte {
	var ckA, ckB byte
	// Checksum covers class, ID, length, and payload (skip sync chars)
	for i := 2; i < len(msg); i++ {
		ckA += msg[i]
		ckB += ckA
	}
	return append(msg, ckA, ckB)
}
