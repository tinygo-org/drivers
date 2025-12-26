package gps

import (
	"testing"
)

func TestAppendChecksum(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "simple message",
			input:    []byte{0xB5, 0x62, 0x06, 0x24, 0x00, 0x00},
			expected: []byte{0xB5, 0x62, 0x06, 0x24, 0x00, 0x00, 0x2A, 0x84},
		},
		{
			name:     "CFG-NAV5 header only",
			input:    []byte{0xB5, 0x62, 0x06, 0x24, 0x24, 0x00},
			expected: []byte{0xB5, 0x62, 0x06, 0x24, 0x24, 0x00, 0x4E, 0xCC},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := appendChecksum(tc.input)

			if len(result) != len(tc.expected) {
				t.Errorf("expected length %d, got %d", len(tc.expected), len(result))
				return
			}

			// Check checksum bytes (last two bytes)
			ckA := result[len(result)-2]
			ckB := result[len(result)-1]
			expectedCkA := tc.expected[len(tc.expected)-2]
			expectedCkB := tc.expected[len(tc.expected)-1]

			if ckA != expectedCkA || ckB != expectedCkB {
				t.Errorf("expected checksum 0x%02X 0x%02X, got 0x%02X 0x%02X",
					expectedCkA, expectedCkB, ckA, ckB)
			}
		})
	}
}

func TestAppendChecksumPreservesOriginal(t *testing.T) {
	input := []byte{0xB5, 0x62, 0x06, 0x24, 0x00, 0x00}
	original := make([]byte, len(input))
	copy(original, input)

	result := appendChecksum(input)

	// Verify original bytes are preserved
	for i := range input {
		if result[i] != original[i] {
			t.Errorf("byte %d changed: expected 0x%02X, got 0x%02X", i, original[i], result[i])
		}
	}

	// Verify two bytes were appended
	if len(result) != len(input)+2 {
		t.Errorf("expected length %d, got %d", len(input)+2, len(result))
	}
}

func TestFlightModeCmdConfig(t *testing.T) {
	// Verify FlightModeCmd has expected values
	if FlightModeCmd.DynModel != 6 {
		t.Errorf("expected DynModel 6 (airborne <1g), got %d", FlightModeCmd.DynModel)
	}

	if FlightModeCmd.FixMode != 3 {
		t.Errorf("expected FixMode 3 (auto 2D/3D), got %d", FlightModeCmd.FixMode)
	}

	expectedMask := CfgNav5Dyn | CfgNav5MinEl | CfgNav5PosFixMode
	if FlightModeCmd.Mask != expectedMask {
		t.Errorf("expected Mask 0x%04X, got 0x%04X", expectedMask, FlightModeCmd.Mask)
	}

	if FlightModeCmd.MinElev_deg != 5 {
		t.Errorf("expected MinElev_deg 5, got %d", FlightModeCmd.MinElev_deg)
	}
}

func TestGNSSDisableCmdConfig(t *testing.T) {
	// Verify GNSSDisableCmd has expected structure
	if GNSSDisableCmd.MsgVer != 0 {
		t.Errorf("expected MsgVer 0, got %d", GNSSDisableCmd.MsgVer)
	}

	if GNSSDisableCmd.NumTrkChHw != 0x20 {
		t.Errorf("expected NumTrkChHw 0x20, got 0x%02X", GNSSDisableCmd.NumTrkChHw)
	}

	if len(GNSSDisableCmd.ConfigBlocks) != 5 {
		t.Errorf("expected 5 config blocks, got %d", len(GNSSDisableCmd.ConfigBlocks))
		return
	}

	// Verify GPS is enabled
	gpsBlock := GNSSDisableCmd.ConfigBlocks[0]
	if gpsBlock.GnssId != 0 {
		t.Errorf("expected first block GnssId 0 (GPS), got %d", gpsBlock.GnssId)
	}
	if gpsBlock.Flags&CfgGnssEnable == 0 {
		t.Error("expected GPS to be enabled")
	}

	// Verify other GNSS are disabled
	for i := 1; i < len(GNSSDisableCmd.ConfigBlocks); i++ {
		block := GNSSDisableCmd.ConfigBlocks[i]
		if block.Flags&CfgGnssEnable != 0 {
			t.Errorf("expected block %d (GnssId %d) to be disabled", i, block.GnssId)
		}
	}
}

func TestFlightModeCmdWrite(t *testing.T) {
	buf := make([]byte, 64)
	n, err := FlightModeCmd.Write(buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if n != 42 {
		t.Errorf("expected 42 bytes, got %d", n)
	}

	// Verify sync chars
	if buf[0] != 0xB5 || buf[1] != 0x62 {
		t.Errorf("expected sync 0xB5 0x62, got 0x%02X 0x%02X", buf[0], buf[1])
	}

	// Verify class/id
	if buf[2] != 0x06 || buf[3] != 0x24 {
		t.Errorf("expected class/id 0x06 0x24, got 0x%02X 0x%02X", buf[2], buf[3])
	}

	// Verify DynModel at offset 8
	if buf[8] != 6 {
		t.Errorf("expected DynModel 6, got %d", buf[8])
	}
}

func TestGNSSDisableCmdWrite(t *testing.T) {
	buf := make([]byte, 64)
	n, err := GNSSDisableCmd.Write(buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 6 header + 4 payload header + 5*8 blocks = 50 bytes
	expectedLen := 6 + 4 + 5*8
	if n != expectedLen {
		t.Errorf("expected %d bytes, got %d", expectedLen, n)
	}

	// Verify sync chars
	if buf[0] != 0xB5 || buf[1] != 0x62 {
		t.Errorf("expected sync 0xB5 0x62, got 0x%02X 0x%02X", buf[0], buf[1])
	}

	// Verify class/id
	if buf[2] != 0x06 || buf[3] != 0x3E {
		t.Errorf("expected class/id 0x06 0x3E, got 0x%02X 0x%02X", buf[2], buf[3])
	}

	// Verify number of blocks
	if buf[9] != 5 {
		t.Errorf("expected 5 blocks, got %d", buf[9])
	}
}

func TestChecksumCalculation(t *testing.T) {
	// Test with known UBX message and expected checksum
	// This is a minimal CFG-NAV5 poll message
	msg := []byte{0xB5, 0x62, 0x06, 0x24, 0x00, 0x00}

	result := appendChecksum(msg)

	// Verify checksum by recalculating
	var ckA, ckB byte
	for i := 2; i < len(msg); i++ {
		ckA += msg[i]
		ckB += ckA
	}

	if result[6] != ckA || result[7] != ckB {
		t.Errorf("checksum mismatch: expected 0x%02X 0x%02X, got 0x%02X 0x%02X",
			ckA, ckB, result[6], result[7])
	}
}

func TestMessageRateCmdConfigs(t *testing.T) {
	testCases := []struct {
		name     string
		cmd      CfgMsg1
		msgClass byte
		msgID    byte
		rate     byte
	}{
		{"GGA", MessageRateGGACmd, 0xF0, 0x00, 1},
		{"GLL", MessageRateGLLCmd, 0xF0, 0x01, 0},
		{"GSA", MessageRateGSACmd, 0xF0, 0x02, 1},
		{"GSV", MessageRateGSVCmd, 0xF0, 0x03, 1},
		{"RMC", MessageRateRMCCmd, 0xF0, 0x04, 1},
		{"VTG", MessageRateVTGCmd, 0xF0, 0x05, 0},
		{"ZDA", MessageRateZDACmd, 0xF0, 0x08, 0},
		{"TXT", MessageRateTXTCmd, 0xF0, 0x41, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd.MsgClass != tc.msgClass {
				t.Errorf("expected MsgClass 0x%02X, got 0x%02X", tc.msgClass, tc.cmd.MsgClass)
			}
			if tc.cmd.MsgID != tc.msgID {
				t.Errorf("expected MsgID 0x%02X, got 0x%02X", tc.msgID, tc.cmd.MsgID)
			}
			if tc.cmd.Rate != tc.rate {
				t.Errorf("expected Rate %d, got %d", tc.rate, tc.cmd.Rate)
			}
		})
	}
}

func TestCfgMsg1Write(t *testing.T) {
	cmd := CfgMsg1{
		MsgClass: 0xF0,
		MsgID:    0x00,
		Rate:     1,
	}

	buf := make([]byte, 16)
	n, err := cmd.Write(buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if n != 9 {
		t.Errorf("expected 9 bytes, got %d", n)
	}

	// Verify sync chars
	if buf[0] != 0xB5 || buf[1] != 0x62 {
		t.Errorf("expected sync 0xB5 0x62, got 0x%02X 0x%02X", buf[0], buf[1])
	}

	// Verify class/id (0x06 0x01 for CFG-MSG)
	if buf[2] != 0x06 || buf[3] != 0x01 {
		t.Errorf("expected class/id 0x06 0x01, got 0x%02X 0x%02X", buf[2], buf[3])
	}

	// Verify length (3 bytes payload)
	if buf[4] != 3 || buf[5] != 0 {
		t.Errorf("expected length 3, got %d", uint16(buf[4])|uint16(buf[5])<<8)
	}

	// Verify payload
	if buf[6] != 0xF0 {
		t.Errorf("expected MsgClass 0xF0, got 0x%02X", buf[6])
	}
	if buf[7] != 0x00 {
		t.Errorf("expected MsgID 0x00, got 0x%02X", buf[7])
	}
	if buf[8] != 1 {
		t.Errorf("expected Rate 1, got %d", buf[8])
	}
}

func TestCfgMsg1ClassID(t *testing.T) {
	cmd := CfgMsg1{}
	if got := cmd.classID(); got != 0x0106 {
		t.Errorf("expected 0x0106, got 0x%04x", got)
	}
}

func TestMinimalMessageRatesConfig(t *testing.T) {
	// Verify the minimal config has correct rates set
	// GGA and RMC should be enabled (rate=1), others disabled (rate=0)
	expectedRates := map[byte]byte{
		0x00: 1, // GGA - enabled
		0x01: 0, // GLL - disabled
		0x02: 1, // GSA - enabled
		0x03: 1, // GSV - enabled
		0x04: 1, // RMC - enabled
		0x05: 0, // VTG - disabled
		0x08: 0, // ZDA - disabled
		0x41: 0, // TXT - disabled
	}

	commands := []CfgMsg1{
		MessageRateGGACmd,
		MessageRateGLLCmd,
		MessageRateGSACmd,
		MessageRateGSVCmd,
		MessageRateRMCCmd,
		MessageRateVTGCmd,
		MessageRateZDACmd,
		MessageRateTXTCmd,
	}

	for _, cmd := range commands {
		expectedRate, ok := expectedRates[cmd.MsgID]
		if !ok {
			t.Errorf("unexpected MsgID 0x%02X", cmd.MsgID)
			continue
		}
		if cmd.Rate != expectedRate {
			t.Errorf("MsgID 0x%02X: expected rate %d, got %d", cmd.MsgID, expectedRate, cmd.Rate)
		}
	}
}

func TestAllMessageRatesWriteCorrectBytes(t *testing.T) {
	// Test that each message rate command writes the correct bytes
	commands := []CfgMsg1{
		MessageRateGGACmd,
		MessageRateGLLCmd,
		MessageRateGSACmd,
		MessageRateGSVCmd,
		MessageRateRMCCmd,
		MessageRateVTGCmd,
		MessageRateZDACmd,
		MessageRateTXTCmd,
	}

	for _, cmd := range commands {
		buf := make([]byte, 16)
		n, err := cmd.Write(buf)

		if err != nil {
			t.Errorf("MsgID 0x%02X: unexpected error: %v", cmd.MsgID, err)
			continue
		}

		if n != 9 {
			t.Errorf("MsgID 0x%02X: expected 9 bytes, got %d", cmd.MsgID, n)
		}

		// Verify MsgClass in payload
		if buf[6] != 0xF0 {
			t.Errorf("MsgID 0x%02X: expected MsgClass 0xF0, got 0x%02X", cmd.MsgID, buf[6])
		}

		// Verify MsgID in payload
		if buf[7] != cmd.MsgID {
			t.Errorf("expected MsgID 0x%02X in payload, got 0x%02X", cmd.MsgID, buf[7])
		}

		// Verify Rate in payload
		if buf[8] != cmd.Rate {
			t.Errorf("MsgID 0x%02X: expected Rate %d, got %d", cmd.MsgID, cmd.Rate, buf[8])
		}
	}
}

func TestSetMessageRatesAllEnabledModifiesRate(t *testing.T) {
	// Verify that when we copy a command and set Rate=1, it works correctly
	cmd := MessageRateGLLCmd // This one is disabled by default
	if cmd.Rate != 0 {
		t.Errorf("expected GLL default rate 0, got %d", cmd.Rate)
	}

	// Simulate what SetMessageRatesAllEnabled does
	cmd.Rate = 1

	buf := make([]byte, 16)
	_, _ = cmd.Write(buf)

	if buf[8] != 1 {
		t.Errorf("expected Rate 1 in buffer, got %d", buf[8])
	}
}
