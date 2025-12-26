package gps

import (
	"testing"
)

func TestCfgNav5ClassID(t *testing.T) {
	cfg := CfgNav5{}
	if got := cfg.classID(); got != 0x2406 {
		t.Errorf("expected 0x2406, got 0x%04x", got)
	}
}

func TestCfgNav5Write(t *testing.T) {
	cfg := CfgNav5{
		Mask:                  CfgNav5Dyn | CfgNav5MinEl,
		DynModel:              4,
		FixMode:               3,
		FixedAlt_me2:          10000,
		FixedAltVar_m2e4:      10000,
		MinElev_deg:           5,
		DrLimit_s:             0,
		PDop:                  250,
		TDop:                  250,
		PAcc_m:                100,
		TAcc_m:                300,
		StaticHoldThresh_cm_s: 50,
		DgnssTimeout_s:        60,
		CnoThreshNumSVs:       3,
		CnoThresh_dbhz:        35,
		Reserved1:             [2]byte{0, 0},
		StaticHoldMaxDist_m:   200,
		UtcStandard:           0,
		Reserved2:             [5]byte{0, 0, 0, 0, 0},
	}

	buf := make([]byte, 64)
	n, err := cfg.Write(buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 42 {
		t.Errorf("expected 42 bytes written, got %d", n)
	}

	// Check sync chars
	if buf[0] != 0xb5 || buf[1] != 0x62 {
		t.Errorf("expected sync chars 0xb5 0x62, got 0x%02x 0x%02x", buf[0], buf[1])
	}

	// Check class/id (little-endian)
	if buf[2] != 0x06 || buf[3] != 0x24 {
		t.Errorf("expected class/id 0x06 0x24, got 0x%02x 0x%02x", buf[2], buf[3])
	}

	// Check length
	if buf[4] != 36 || buf[5] != 0 {
		t.Errorf("expected length 36, got %d", uint16(buf[4])|uint16(buf[5])<<8)
	}

	// Check Mask (little-endian)
	mask := uint16(buf[6]) | uint16(buf[7])<<8
	if mask != uint16(CfgNav5Dyn|CfgNav5MinEl) {
		t.Errorf("expected mask 0x03, got 0x%04x", mask)
	}

	// Check DynModel
	if buf[8] != 4 {
		t.Errorf("expected DynModel 4, got %d", buf[8])
	}

	// Check FixMode
	if buf[9] != 3 {
		t.Errorf("expected FixMode 3, got %d", buf[9])
	}

	// Check FixedAlt_me2 (little-endian int32)
	fixedAlt := int32(buf[10]) | int32(buf[11])<<8 | int32(buf[12])<<16 | int32(buf[13])<<24
	if fixedAlt != 10000 {
		t.Errorf("expected FixedAlt_me2 10000, got %d", fixedAlt)
	}
}

func TestCfgGnssClassID(t *testing.T) {
	cfg := CfgGnss{}
	if got := cfg.classID(); got != 0x3e06 {
		t.Errorf("expected 0x3e06, got 0x%04x", got)
	}
}

func TestCfgGnssWrite(t *testing.T) {
	testCases := []struct {
		name           string
		cfg            CfgGnss
		expectedLen    int
		expectedBlocks byte
	}{
		{
			name: "no config blocks",
			cfg: CfgGnss{
				MsgVer:       0,
				NumTrkChHw:   32,
				NumTrkChUse:  32,
				ConfigBlocks: nil,
			},
			expectedLen:    10,
			expectedBlocks: 0,
		},
		{
			name: "one config block",
			cfg: CfgGnss{
				MsgVer:      0,
				NumTrkChHw:  32,
				NumTrkChUse: 32,
				ConfigBlocks: []*CfgGnssConfigBlocksType{
					{GnssId: 0, ResTrkCh: 8, MaxTrkCh: 16, Flags: CfgGnssEnable | 0x010000},
				},
			},
			expectedLen:    18,
			expectedBlocks: 1,
		},
		{
			name: "two config blocks",
			cfg: CfgGnss{
				MsgVer:      0,
				NumTrkChHw:  32,
				NumTrkChUse: 32,
				ConfigBlocks: []*CfgGnssConfigBlocksType{
					{GnssId: 0, ResTrkCh: 8, MaxTrkCh: 16, Flags: CfgGnssEnable | 0x010000},
					{GnssId: 6, ResTrkCh: 8, MaxTrkCh: 14, Flags: CfgGnssEnable | 0x010000},
				},
			},
			expectedLen:    26,
			expectedBlocks: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := make([]byte, 64)
			n, err := tc.cfg.Write(buf)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if n != tc.expectedLen {
				t.Errorf("expected %d bytes written, got %d", tc.expectedLen, n)
			}

			// Check sync chars
			if buf[0] != 0xb5 || buf[1] != 0x62 {
				t.Errorf("expected sync chars 0xb5 0x62, got 0x%02x 0x%02x", buf[0], buf[1])
			}

			// Check class/id (little-endian)
			if buf[2] != 0x06 || buf[3] != 0x3e {
				t.Errorf("expected class/id 0x06 0x3e, got 0x%02x 0x%02x", buf[2], buf[3])
			}

			// Check number of config blocks
			if buf[9] != tc.expectedBlocks {
				t.Errorf("expected %d config blocks, got %d", tc.expectedBlocks, buf[9])
			}
		})
	}
}

func TestCfgGnssWriteBlockContent(t *testing.T) {
	cfg := CfgGnss{
		MsgVer:      0,
		NumTrkChHw:  32,
		NumTrkChUse: 32,
		ConfigBlocks: []*CfgGnssConfigBlocksType{
			{GnssId: 0, ResTrkCh: 8, MaxTrkCh: 16, Reserved1: 0, Flags: CfgGnssEnable | 0x010000},
		},
	}

	buf := make([]byte, 64)
	_, _ = cfg.Write(buf)

	// Check first block at offset 10
	if buf[10] != 0 {
		t.Errorf("expected GnssId 0, got %d", buf[10])
	}
	if buf[11] != 8 {
		t.Errorf("expected ResTrkCh 8, got %d", buf[11])
	}
	if buf[12] != 16 {
		t.Errorf("expected MaxTrkCh 16, got %d", buf[12])
	}

	// Check flags (little-endian uint32)
	flags := uint32(buf[14]) | uint32(buf[15])<<8 | uint32(buf[16])<<16 | uint32(buf[17])<<24
	expectedFlags := uint32(CfgGnssEnable | 0x010000)
	if flags != expectedFlags {
		t.Errorf("expected flags 0x%08x, got 0x%08x", expectedFlags, flags)
	}
}
