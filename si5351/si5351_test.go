package si5351

import (
	"testing"
)

func TestSelectRDiv(t *testing.T) {
	d := &Device{}

	tests := []struct {
		name     string
		freq     Frequency
		wantDiv  uint8
		wantFreq Frequency
	}{
		{"4kHz", 4000 * FREQ_MULT, OUTPUT_CLK_DIV_128, 4000 * FREQ_MULT * 128},
		{"8kHz", 8000 * FREQ_MULT, OUTPUT_CLK_DIV_64, 8000 * FREQ_MULT * 64},
		{"16kHz", 16000 * FREQ_MULT, OUTPUT_CLK_DIV_32, 16000 * FREQ_MULT * 32},
		{"32kHz", 32000 * FREQ_MULT, OUTPUT_CLK_DIV_16, 32000 * FREQ_MULT * 16},
		{"64kHz", 64000 * FREQ_MULT, OUTPUT_CLK_DIV_8, 64000 * FREQ_MULT * 8},
		{"128kHz", 128000 * FREQ_MULT, OUTPUT_CLK_DIV_4, 128000 * FREQ_MULT * 4},
		{"256kHz", 256000 * FREQ_MULT, OUTPUT_CLK_DIV_2, 256000 * FREQ_MULT * 2},
		{"512kHz", 512000 * FREQ_MULT, OUTPUT_CLK_DIV_1, 512000 * FREQ_MULT},
		{"1MHz", 1000000 * FREQ_MULT, OUTPUT_CLK_DIV_1, 1000000 * FREQ_MULT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			freq := tt.freq
			freq, gotDiv := d.selectRDiv(freq)
			if gotDiv != tt.wantDiv {
				t.Errorf("selectRDiv() div = %v, want %v", gotDiv, tt.wantDiv)
			}
			if freq != tt.wantFreq {
				t.Errorf("selectRDiv() freq = %v, want %v", freq, tt.wantFreq)
			}
		})
	}
}

func TestSelectRDivMS67(t *testing.T) {
	d := &Device{}

	tests := []struct {
		name     string
		freq     Frequency
		wantDiv  uint8
		wantFreq Frequency
	}{
		{"4kHz", 4000 * FREQ_MULT, OUTPUT_CLK_DIV_128, 4000 * FREQ_MULT * 128},
		{"8kHz", 8000 * FREQ_MULT, OUTPUT_CLK_DIV_64, 8000 * FREQ_MULT * 64},
		{"16kHz", 16000 * FREQ_MULT, OUTPUT_CLK_DIV_32, 16000 * FREQ_MULT * 32},
		{"64kHz", 64000 * FREQ_MULT, OUTPUT_CLK_DIV_8, 64000 * FREQ_MULT * 8},
		{"256kHz", 256000 * FREQ_MULT, OUTPUT_CLK_DIV_2, 256000 * FREQ_MULT * 2},
		{"512kHz", 512000 * FREQ_MULT, OUTPUT_CLK_DIV_1, 512000 * FREQ_MULT},
		{"1MHz", 1000000 * FREQ_MULT, OUTPUT_CLK_DIV_1, 1000000 * FREQ_MULT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			freq := tt.freq
			freq, gotDiv := d.selectRDivMS67(freq)
			if gotDiv != tt.wantDiv {
				t.Errorf("selectRDivMS67() div = %v, want %v", gotDiv, tt.wantDiv)
			}
			if freq != tt.wantFreq {
				t.Errorf("selectRDivMS67() freq = %v, want %v", freq, tt.wantFreq)
			}
		})
	}
}

func TestCalculatePLL(t *testing.T) {
	d := &Device{}
	d.crystalFreq[0] = 25000000

	tests := []struct {
		name    string
		freq    Frequency
		wantMin Frequency
		wantMax Frequency
	}{
		{"600MHz", 600000000 * FREQ_MULT, 599000000 * FREQ_MULT, 601000000 * FREQ_MULT},
		{"750MHz", 750000000 * FREQ_MULT, 749000000 * FREQ_MULT, 751000000 * FREQ_MULT},
		{"900MHz", 900000000 * FREQ_MULT, 899000000 * FREQ_MULT, 901000000 * FREQ_MULT},
		{"BelowMin", 500000000 * FREQ_MULT, 600000000 * FREQ_MULT, 600000000 * FREQ_MULT},
		{"AboveMax", 1000000000 * FREQ_MULT, 900000000 * FREQ_MULT, 900000000 * FREQ_MULT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reg := d.CalculatePLL(PLL_A, tt.freq, 0, false)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculatePLL() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
			if reg.p1 == 0 || reg.p3 == 0 {
				t.Errorf("CalculatePLL() invalid register values: p1=%v, p2=%v, p3=%v", reg.p1, reg.p2, reg.p3)
			}
		})
	}
}

func TestCalculateMultisynth(t *testing.T) {
	d := &Device{}

	tests := []struct {
		name    string
		freq    Frequency
		pllFreq Frequency
		wantDiv bool
	}{
		{"10MHz from 800MHz", 10000000 * FREQ_MULT, 800000000 * FREQ_MULT, false},
		{"1MHz from 800MHz", 1000000 * FREQ_MULT, 800000000 * FREQ_MULT, false},
		{"Auto PLL 10MHz", 10000000 * FREQ_MULT, 0, false},
		{"150MHz DivBy4", 150000000 * FREQ_MULT, 600000000 * FREQ_MULT, true},
		{"BelowMin", 100000 * FREQ_MULT, 800000000 * FREQ_MULT, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reg := d.CalculateMultisynth(tt.freq, tt.pllFreq)
			if tt.pllFreq == 0 {
				// Auto mode should return a valid PLL frequency
				if got < PLL_VCO_MIN*FREQ_MULT || got > PLL_VCO_MAX*FREQ_MULT {
					t.Errorf("CalculateMultisynth() returned invalid PLL freq %v", got)
				}
			}
			if reg.p3 == 0 {
				t.Errorf("CalculateMultisynth() p3 should not be 0")
			}
		})
	}
}

func TestMultisynth67Calc(t *testing.T) {
	d := &Device{}

	tests := []struct {
		name    string
		freq    Frequency
		pllFreq Frequency
		wantErr bool
	}{
		{"10MHz Auto", 10000000 * FREQ_MULT, 0, false},
		{"100MHz Auto", 100000000 * FREQ_MULT, 0, false},
		{"100MHz from 800MHz", 100000000 * FREQ_MULT, 800000000 * FREQ_MULT, false},
		{"Invalid Division", 10000000 * FREQ_MULT, 777000000 * FREQ_MULT, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reg := d.multisynth67Calc(tt.freq, tt.pllFreq)
			if tt.pllFreq == 0 {
				if got < PLL_VCO_MIN*FREQ_MULT || got > PLL_VCO_MAX*FREQ_MULT {
					t.Errorf("multisynth67Calc() returned invalid PLL freq %v", got)
				}
			} else if tt.wantErr {
				if got != 0 {
					t.Errorf("multisynth67Calc() should return 0 for invalid division, got %v", got)
				}
			}
			if reg.p1 == 0 && !tt.wantErr {
				t.Errorf("multisynth67Calc() p1 should not be 0")
			}
		})
	}
}

func TestSetCorrection(t *testing.T) {
	// Skip this test as it requires a mock I2C bus
	t.Skip("Requires mock I2C bus implementation")
}

func TestSetRefFreq(t *testing.T) {
	d := &Device{}

	tests := []struct {
		name     string
		freq     CrystalFrequency
		wantFreq CrystalFrequency
		wantDiv  uint8
	}{
		{"25MHz", 25000000, 25000000, CLKIN_DIV_1},
		{"50MHz", 50000000, 25000000, CLKIN_DIV_2},
		{"100MHz", 100000000, 25000000, CLKIN_DIV_4},
		{"30MHz", 30000000, 30000000, CLKIN_DIV_1},
		{"60MHz", 60000000, 30000000, CLKIN_DIV_2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d.SetReferenceFrequency(PLLInputClockIn, tt.freq)
			if d.crystalFreq[PLLInputClockIn] != tt.wantFreq {
				t.Errorf("SetReferenceFrequency() freq = %v, want %v", d.crystalFreq[PLLInputClockIn], tt.wantFreq)
			}
			if d.clkinDiv != tt.wantDiv {
				t.Errorf("SetReferenceFrequency() clkinDiv = %v, want %v", d.clkinDiv, tt.wantDiv)
			}
		})
	}
}

func TestGetCorrection(t *testing.T) {
	d := &Device{}
	d.refCorrection[PLLInputXO] = 5000
	d.refCorrection[PLLInputClockIn] = -3000

	if got := d.GetCorrection(PLLInputXO); got != 5000 {
		t.Errorf("GetCorrection(PLLInputXO) = %v, want 5000", got)
	}
	if got := d.GetCorrection(PLLInputClockIn); got != -3000 {
		t.Errorf("GetCorrection(PLLInputClockIn) = %v, want -3000", got)
	}
}
