package gps

// UBX message classes
const (
	ubxClassACK = 0x05
)

// UBX ACK message IDs
const (
	ubxACK_NAK = 0x00 // Message not acknowledged
	ubxACK_ACK = 0x01 // Message acknowledged
)

// UBX sync characters
const (
	ubxSyncChar1 = 0xB5
	ubxSyncChar2 = 0x62
)

const (
	DynModePortable   = 0
	DynModeStationary = 2
	DynModePedestrian = 3
	DynModeAutomotive = 4
	DynModeSea        = 5
	DynModeAirborne1g = 6
	DynModeAirborne2g = 7
	DynModeAirborne4g = 8
	DynModeWristWatch = 9
	DynModeBike       = 10
)

const (
	FixMode2D   = 1
	FixMode3D   = 2
	FixModeAuto = 3
)

// from https://github.com/daedaleanai/ublox/blob/main/ubx/messages.go

// Message ubx-cfg-nav5

// CfgNav5 (Get/set) Navigation engine settings
// Class/Id 0x06 0x24 (36 bytes)
// See the Navigation Configuration Settings Description for a detailed description of how these settings affect receiver operation.
type CfgNav5 struct {
	Mask                  CfgNav5Mask // Parameters bitmask. Only the masked parameters will be applied.
	DynModel              byte        // Dynamic platform model: 0: portable 2: stationary 3: pedestrian 4: automotive 5: sea 6: airborne with <1g acceleration 7: airborne with <2g acceleration 8: airborne with <4g acceleration 9: wrist-worn watch (not supported in protocol versions less than 18) 10: bike (supported in protocol versions 19. 2)
	FixMode               byte        // Position fixing mode: 1: 2D only 2: 3D only 3: auto 2D/3D
	FixedAlt_me2          int32       // [1e-2 m] Fixed altitude (mean sea level) for 2D fix mode
	FixedAltVar_m2e4      uint32      // [1e-4 m^2] Fixed altitude variance for 2D mode
	MinElev_deg           int8        // [deg] Minimum elevation for a GNSS satellite to be used in NAV
	DrLimit_s             byte        // [s] Reserved
	PDop                  uint16      // Position DOP mask to use
	TDop                  uint16      // Time DOP mask to use
	PAcc_m                uint16      // [m] Position accuracy mask
	TAcc_m                uint16      // [m] Time accuracy mask
	StaticHoldThresh_cm_s byte        // [cm/s] Static hold threshold
	DgnssTimeout_s        byte        // [s] DGNSS timeout
	CnoThreshNumSVs       byte        // Number of satellites required to have C/N0 above cnoThresh for a fix to be attempted
	CnoThresh_dbhz        byte        // [dBHz] C/N0 threshold for deciding whether to attempt a fix
	Reserved1             [2]byte     // Reserved
	StaticHoldMaxDist_m   uint16      // [m] Static hold distance threshold (before quitting static hold)
	UtcStandard           byte        // UTC standard to be used: 0: Automatic; receiver selects based on GNSS configuration (see GNSS time bases) 3: UTC as operated by the U.S. Naval Observatory (USNO); derived from GPS time 5: UTC as combined from multiple European laboratories; derived from Galileo time 6: UTC as operated by the former Soviet Union (SU); derived from GLONASS time 7: UTC as operated by the National Time Service Center (NTSC), China; derived from BeiDou time (not supported in protocol versions less than 16).
	Reserved2             [5]byte     // Reserved
}

func (CfgNav5) classID() uint16 { return 0x2406 }

type CfgNav5Mask uint16

const (
	CfgNav5Dyn            CfgNav5Mask = 0x1   // Apply dynamic model settings
	CfgNav5MinEl          CfgNav5Mask = 0x2   // Apply minimum elevation settings
	CfgNav5PosFixMode     CfgNav5Mask = 0x4   // Apply fix mode settings
	CfgNav5DrLim          CfgNav5Mask = 0x8   // Reserved
	CfgNav5PosMask        CfgNav5Mask = 0x10  // Apply position mask settings
	CfgNav5TimeMask       CfgNav5Mask = 0x20  // Apply time mask settings
	CfgNav5StaticHoldMask CfgNav5Mask = 0x40  // Apply static hold settings
	CfgNav5DgpsMask       CfgNav5Mask = 0x80  // Apply DGPS settings
	CfgNav5CnoThreshold   CfgNav5Mask = 0x100 // Apply CNO threshold settings (cnoThresh, cnoThreshNumSVs)
	CfgNav5Utc            CfgNav5Mask = 0x400 // Apply UTC settings (not supported in protocol versions less than 16).
)

// Write CfgNav5 message to buffer
func (cfg CfgNav5) Write(buf []byte) (int, error) {
	copy(buf, []byte{0xb5, 0x62, byte(cfg.classID()), byte(cfg.classID() >> 8), 36, 0})

	buf[6] = byte(cfg.Mask)
	buf[7] = byte(cfg.Mask >> 8)
	buf[8] = cfg.DynModel
	buf[9] = cfg.FixMode
	buf[10] = byte(cfg.FixedAlt_me2)
	buf[11] = byte(cfg.FixedAlt_me2 >> 8)
	buf[12] = byte(cfg.FixedAlt_me2 >> 16)
	buf[13] = byte(cfg.FixedAlt_me2 >> 24)
	buf[14] = byte(cfg.FixedAltVar_m2e4)
	buf[15] = byte(cfg.FixedAltVar_m2e4 >> 8)
	buf[16] = byte(cfg.FixedAltVar_m2e4 >> 16)
	buf[17] = byte(cfg.FixedAltVar_m2e4 >> 24)
	buf[18] = byte(cfg.MinElev_deg)
	buf[19] = cfg.DrLimit_s
	buf[20] = byte(cfg.PDop)
	buf[21] = byte(cfg.PDop >> 8)
	buf[22] = byte(cfg.TDop)
	buf[23] = byte(cfg.TDop >> 8)
	buf[24] = byte(cfg.PAcc_m)
	buf[25] = byte(cfg.PAcc_m >> 8)
	buf[26] = byte(cfg.TAcc_m)
	buf[27] = byte(cfg.TAcc_m >> 8)
	buf[28] = cfg.StaticHoldThresh_cm_s
	buf[29] = cfg.DgnssTimeout_s
	buf[30] = cfg.CnoThreshNumSVs
	buf[31] = cfg.CnoThresh_dbhz
	copy(buf[32:34], cfg.Reserved1[:])
	buf[34] = byte(cfg.StaticHoldMaxDist_m)
	buf[35] = byte(cfg.StaticHoldMaxDist_m >> 8)
	buf[36] = cfg.UtcStandard
	copy(buf[37:42], cfg.Reserved2[:])

	return 42, nil
}

// Message ubx-cfg-msg

// CfgMsg1 (Get/set) Set message rate
// Class/Id 0x06 0x01 (3 bytes)
// Set message rate configuration for the current port. See also section How to change between protocols.
type CfgMsg1 struct {
	MsgClass byte // Message class
	MsgID    byte // Message identifier
	Rate     byte // Send rate on current port
}

func (CfgMsg1) classID() uint16 { return 0x0106 }

func (cfg CfgMsg1) Write(buf []byte) (int, error) {
	copy(buf, []byte{0xb5, 0x62, byte(cfg.classID()), byte(cfg.classID() >> 8), 3, 0})

	buf[6] = cfg.MsgClass
	buf[7] = cfg.MsgID
	buf[8] = cfg.Rate

	return 9, nil
}

// Message ubx-cfg-gnss

// CfgGnss (Get/set) GNSS system configuration
// Class/Id 0x06 0x3e (4 + N*8 bytes)
// Gets or sets the GNSS system channel sharing configuration. If the receiver is sent a valid new configuration, it will respond with a UBX-ACK- ACK message and immediately change to the new configuration. Otherwise the receiver will reject the request, by issuing a UBX-ACK-NAK and continuing operation with the previous configuration. Configuration requirements:  It is necessary for at least one major GNSS to be enabled, after applying the  new configuration to the current one.  It is also required that at least 4 tracking channels are available to each  enabled major GNSS, i.e. maxTrkCh must have a minimum value of 4 for each  enabled major GNSS.  The number of tracking channels in use must not exceed the number of  tracking channels available in hardware, and the sum of all reserved tracking  channels needs to be less than or equal to the number of tracking channels in  use. Notes:  To avoid cross-correlation issues, it is recommended that GPS and QZSS are  always both enabled or both disabled.  Polling this message returns the configuration of all supported GNSS, whether  enabled or not; it may also include GNSS unsupported by the particular  product, but in such cases the enable flag will always be unset.  See section GNSS Configuration for a discussion of the use of this message.  See section Satellite Numbering for a description of the GNSS IDs available.  Configuration specific to the GNSS system can be done via other messages (e.  g. UBX-CFG-SBAS).
type CfgGnss struct {
	MsgVer          byte                       // Message version (0x00 for this version)
	NumTrkChHw      byte                       // Number of tracking channels available in hardware (read only)
	NumTrkChUse     byte                       // (Read only in protocol versions greater than 23) Number of tracking channels to use. Must be > 0, <= numTrkChHw. If 0xFF, then number of tracking channels to use will be set to numTrkChHw.
	NumConfigBlocks byte                       `len:"ConfigBlocks"` // Number of configuration blocks following
	ConfigBlocks    []*CfgGnssConfigBlocksType // len: NumConfigBlocks
}

func (CfgGnss) classID() uint16 { return 0x3e06 }

type CfgGnssConfigBlocksType struct {
	GnssId    byte         // System identifier (see Satellite Numbering )
	ResTrkCh  byte         // (Read only in protocol versions greater than 23) Number of reserved (minimum) tracking channels for this system.
	MaxTrkCh  byte         // (Read only in protocol versions greater than 23) Maximum number of tracking channels used for this system. Must be > 0, >= resTrkChn, <= numTrkChUse and <= maximum number of tracking channels supported for this system.
	Reserved1 byte         // Reserved
	Flags     CfgGnssFlags // Bitfield of flags. At least one signal must be configured in every enabled system.
}

type CfgGnssFlags uint32

const (
	CfgGnssEnable     CfgGnssFlags = 0x1      // Enable this system
	CfgGnssSigCfgMask CfgGnssFlags = 0xff0000 // Signal configuration mask When gnssId is 0 (GPS) 0x01 = GPS L1C/A 0x10 = GPS L2C 0x20 = GPS L5 When gnssId is 1 (SBAS) 0x01 = SBAS L1C/A When gnssId is 2 (Galileo) 0x01 = Galileo E1 (not supported in protocol versions less than 18) 0x10 = Galileo E5a 0x20 = Galileo E5b When gnssId is 3 (BeiDou) 0x01 = BeiDou B1I 0x10 = BeiDou B2I 0x80 = BeiDou B2A When gnssId is 4 (IMES) 0x01 = IMES L1 When gnssId is 5 (QZSS) 0x01 = QZSS L1C/A 0x04 = QZSS L1S 0x10 = QZSS L2C 0x20 = QZSS L5 When gnssId is 6 (GLONASS) 0x01 = GLONASS L1 0x10 = GLONASS L2
)

// Write CfgGnss message to buffer
func (cfg CfgGnss) Write(buf []byte) (int, error) {
	copy(buf, []byte{0xb5, 0x62, byte(cfg.classID()), byte(cfg.classID() >> 8), 4 + byte(len(cfg.ConfigBlocks))*8, 0})

	buf[6] = cfg.MsgVer
	buf[7] = cfg.NumTrkChHw
	buf[8] = cfg.NumTrkChUse
	buf[9] = byte(len(cfg.ConfigBlocks))

	offset := 10
	for _, block := range cfg.ConfigBlocks {
		buf[offset] = block.GnssId
		buf[offset+1] = block.ResTrkCh
		buf[offset+2] = block.MaxTrkCh
		buf[offset+3] = block.Reserved1
		buf[offset+4] = byte(block.Flags)
		buf[offset+5] = byte(block.Flags >> 8)
		buf[offset+6] = byte(block.Flags >> 16)
		buf[offset+7] = byte(block.Flags >> 24)
		offset += 8
	}

	return offset, nil
}
