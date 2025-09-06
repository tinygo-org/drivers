//nolint:funlen // ok for tests
package hx711

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // ok for test
var (
	// to speed up tests
	fastCfg = DeviceConfig{
		ResetDuration: 1 * time.Nanosecond,
		SettlingTime:  2 * time.Nanosecond,
		ReadTimeout:   3 * time.Nanosecond,
		ReadTriesMax:  2,
	}

	getValuesZero = [24]bool{
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
	}

	// first value is ready state
	readyValuesZero = append([]bool{false}, getValuesZero[:]...)
	// first value simulates busy-state
	busyReadyValuesZero = append([]bool{true}, readyValuesZero...)

	// 123dec, 0x7B, 0111 1011, MSB first needed
	getValues123 = [24]bool{
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, true, true, true, true, false, true, true,
	}

	// 7654321, 0x74CBB1, 0111 0100 1100 1011 1011 0001, MSB first
	getValues7654321 = [24]bool{
		false, true, true, true, false, true, false, false,
		true, true, false, false, true, false, true, true,
		true, false, true, true, false, false, false, true,
	}

	// the maximum for 24 bit: 0x7FFFFF, 8388607
	getValues4bitMax = [24]bool{
		false, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true,
	}

	// first value is ready state
	readyValues123      = append([]bool{false}, getValues123[:]...)
	readyValues7654321  = append([]bool{false}, getValues7654321[:]...)
	readyValues24bitMax = append([]bool{false}, getValues4bitMax[:]...)

	clkReset     = []bool{true, false}
	clk24BitData = [48]bool{
		true, false, true, false, true, false, true, false, true, false, true, false,
		true, false, true, false, true, false, true, false, true, false, true, false,
		true, false, true, false, true, false, true, false, true, false, true, false,
		true, false, true, false, true, false, true, false, true, false, true, false,
	}

	// last pulse for A@128 selection
	clkDataA128 = append(clk24BitData[:], true, false)
	// last pulses for B@32 selection
	clkDataB32 = append(clk24BitData[:], true, false, true, false)
	// last pulses for A@64 selection
	clkDataA64 = append(clk24BitData[:], true, false, true, false, true, false)
)

func TestNew(t *testing.T) {
	// arrange
	pClkMock := outputPinMock{}
	pClk := newOutPin(&pClkMock)
	pDtaMock := inputPinMock{}
	pDta := newInPin(&pDtaMock)
	// act
	d := New(pClk, pDta, A128)
	// assert
	assert.IsType(t, &Device{}, d)
	assert.NotNil(t, d.powerDownAndSckPinSetState)
	assert.NotNil(t, d.dataPinState)
	assert.NotNil(t, d.firstReadingCfg)
	assert.Equal(t, A128, d.firstReadingCfg.gainAndChanAfterRead)
	assert.Equal(t, int32(0), d.firstReadingCfg.offset)
	assert.Equal(t, int32(1), d.firstReadingCfg.calibrationFactorNum)
	assert.Equal(t, int32(1), d.firstReadingCfg.calibrationFactorDen)
	assert.NotNil(t, d.secondReadingCfg)
	assert.Equal(t, None, d.secondReadingCfg.gainAndChanAfterRead)
	assert.Equal(t, int32(0), d.secondReadingCfg.offset)
	assert.Equal(t, int32(1), d.secondReadingCfg.calibrationFactorNum)
	assert.Equal(t, int32(1), d.secondReadingCfg.calibrationFactorDen)
	assert.NotNil(t, d.mu)
	assert.NotNil(t, d.devCfg)
	assert.Equal(t, DefaultConfig, d.devCfg)
}

func TestNew_gainAndChannelCfg(t *testing.T) {
	defaultReadingConfig := readingConfig{calibrationFactorNum: 1, calibrationFactorDen: 1}
	tests := map[string]struct {
		simGainAndChannelCfg         GainAndChannelCfg
		wantFirstReadingGainAndChan  GainAndChannelCfg
		wantSecondReadingGainAndChan GainAndChannelCfg
	}{
		"new_none_is_a128": {
			simGainAndChannelCfg:         None,
			wantFirstReadingGainAndChan:  A128,
			wantSecondReadingGainAndChan: None,
		},
		"new_a128": {
			simGainAndChannelCfg:         A128,
			wantFirstReadingGainAndChan:  A128,
			wantSecondReadingGainAndChan: None,
		},
		"new_a64": {
			simGainAndChannelCfg:         A64,
			wantFirstReadingGainAndChan:  A64,
			wantSecondReadingGainAndChan: None,
		},
		"new_b32": {
			simGainAndChannelCfg:         B32,
			wantFirstReadingGainAndChan:  B32,
			wantSecondReadingGainAndChan: None,
		},
		"new_a128b32": {
			simGainAndChannelCfg:         A128B32,
			wantFirstReadingGainAndChan:  B32,
			wantSecondReadingGainAndChan: A128,
		},
		"new_a64b32": {
			simGainAndChannelCfg:         A64B32,
			wantFirstReadingGainAndChan:  B32,
			wantSecondReadingGainAndChan: A64,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			pClkMock := outputPinMock{}
			pClk := newOutPin(&pClkMock)
			pDtaMock := inputPinMock{}
			pDta := newInPin(&pDtaMock)
			wantFirstReadingCfg := defaultReadingConfig
			wantFirstReadingCfg.gainAndChanAfterRead = tc.wantFirstReadingGainAndChan
			wantSecondReadingCfg := defaultReadingConfig
			wantSecondReadingCfg.gainAndChanAfterRead = tc.wantSecondReadingGainAndChan
			// act
			d := New(pClk, pDta, tc.simGainAndChannelCfg)
			// assert
			assert.Equal(t, wantFirstReadingCfg, d.firstReadingCfg)
			assert.Equal(t, wantSecondReadingCfg, d.secondReadingCfg)
		})
	}
}

func TestConfigure(t *testing.T) {
	const tries = 5
	devCfg := DeviceConfig{
		ResetDuration: minResetDuration, // to prevent messages
		SettlingTime:  2 * time.Nanosecond,
		ReadTimeout:   3 * time.Nanosecond,
		ReadTriesMax:  tries,
	}

	tests := map[string]struct {
		simDevConfig         *DeviceConfig
		simGainAndChannelCfg GainAndChannelCfg
		simPinDtaGetValues   []bool
		simClkPinSetErr      error
		simDtaPinGetErr      error
		wantDevConfig        DeviceConfig
		wantPinClkSetCalled  []bool
		wantErr              string
	}{
		"configure_none_is_a128": {
			simDevConfig:         &devCfg,
			simGainAndChannelCfg: None,
			simPinDtaGetValues:   busyReadyValuesZero,
			wantDevConfig:        devCfg,
			wantPinClkSetCalled:  append(clkReset, clkDataA128...),
		},
		"configure_a128": {
			simDevConfig:         &devCfg,
			simGainAndChannelCfg: A128,
			simPinDtaGetValues:   busyReadyValuesZero,
			wantDevConfig:        devCfg,
			wantPinClkSetCalled:  append(clkReset, clkDataA128...),
		},
		"configure_b32": {
			simDevConfig:         &devCfg,
			simGainAndChannelCfg: B32,
			simPinDtaGetValues:   busyReadyValuesZero,
			wantDevConfig:        devCfg,
			wantPinClkSetCalled:  append(clkReset, clkDataB32...),
		},
		"configure_a64_keep_default_config": {
			simDevConfig:         nil,
			simGainAndChannelCfg: A64,
			simPinDtaGetValues:   busyReadyValuesZero,
			wantDevConfig:        DefaultConfig,
			wantPinClkSetCalled:  append(clkReset, clkDataA64...),
		},
		"configure_a128b32_with_adjustement": {
			simDevConfig: &DeviceConfig{
				ResetDuration: 59 * time.Microsecond, // forces adjustment
				SettlingTime:  1 * time.Nanosecond,
				ReadTimeout:   2 * time.Nanosecond,
				ReadTriesMax:  0, // forces adjustment
			},
			simGainAndChannelCfg: A128B32,
			// must contain no busy state, because only 1 try is allowed
			simPinDtaGetValues: append(readyValuesZero, readyValuesZero...),
			wantDevConfig: DeviceConfig{
				ResetDuration: minResetDuration,
				SettlingTime:  1 * time.Nanosecond,
				ReadTimeout:   2 * time.Nanosecond,
				ReadTriesMax:  1,
			},
			wantPinClkSetCalled: append(append(clkReset, clkDataB32...), clkDataA128...),
		},
		"configure_a64b32": {
			simDevConfig:         &devCfg,
			simGainAndChannelCfg: A64B32,
			simPinDtaGetValues:   append(busyReadyValuesZero, readyValuesZero...),
			wantDevConfig:        devCfg,
			wantPinClkSetCalled:  append(append(clkReset, clkDataB32...), clkDataA64...),
		},
		"error_set_clk_pin": {
			simDevConfig:        &devCfg,
			simClkPinSetErr:     errors.New("set clk err"),
			wantDevConfig:       devCfg,
			wantPinClkSetCalled: []bool{true},
			wantErr:             "set clk err",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			pClkMock := outputPinMock{
				setErr: tc.simClkPinSetErr,
			}
			pClk := newOutPin(&pClkMock)
			pDtaMock := inputPinMock{
				getSimReturn: tc.simPinDtaGetValues,
				getErr:       tc.simDtaPinGetErr,
			}
			pDta := newInPin(&pDtaMock)
			d := New(pClk, pDta, tc.simGainAndChannelCfg)
			// act
			err := d.Configure(tc.simDevConfig)
			// assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantDevConfig, d.devCfg)
			assert.Equal(t, tc.wantPinClkSetCalled, pClkMock.setCalled)
			assert.Equal(t, len(tc.simPinDtaGetValues), pDtaMock.getCalled)
		})
	}
}

func TestZero(t *testing.T) {
	tests := map[string]struct {
		secondReading         bool
		simGainAndChannelCfg  GainAndChannelCfg
		simPinDtaGetValues    []bool
		simClkPinSetErr       error
		wantFirstReadingOffs  int32
		wantSecondReadingOffs int32
		wantErr               string
	}{
		"zero_first": {
			simGainAndChannelCfg:  A128,
			simPinDtaGetValues:    readyValues123,
			wantFirstReadingOffs:  -123,
			wantSecondReadingOffs: 0,
		},
		"zero_second": {
			secondReading:         true,
			simGainAndChannelCfg:  A128B32,
			simPinDtaGetValues:    append(readyValues123, readyValues7654321...),
			wantFirstReadingOffs:  0,
			wantSecondReadingOffs: -7654321,
		},
		"error_zero_no_second": {
			secondReading:        true,
			simGainAndChannelCfg: A128,
			wantErr:              "no zero possible, second reading is not configured",
		},
		"error_get_dta_timeout": {
			simPinDtaGetValues: []bool{true, true},
			wantErr:            "timeout reached for HX711 on wait for ready state (1)",
		},
		"error_zero_read": {
			simPinDtaGetValues:   []bool{false},
			simClkPinSetErr:      errors.New("set clk on zero err"),
			simGainAndChannelCfg: A128,
			wantErr:              "set clk on zero err",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			pClkMock := outputPinMock{
				setErr: tc.simClkPinSetErr,
			}
			pClk := newOutPin(&pClkMock)
			pDtaMock := inputPinMock{getSimReturn: tc.simPinDtaGetValues}
			pDta := newInPin(&pDtaMock)
			d := New(pClk, pDta, tc.simGainAndChannelCfg)
			d.devCfg = fastCfg
			// act
			err := d.Zero(tc.secondReading)
			// assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, len(tc.simPinDtaGetValues), pDtaMock.getCalled)
			assert.Equal(t, tc.wantFirstReadingOffs, d.firstReadingCfg.offset)
			assert.Equal(t, int32(1), d.firstReadingCfg.calibrationFactorNum)
			assert.Equal(t, int32(1), d.firstReadingCfg.calibrationFactorDen)
			assert.Equal(t, tc.wantSecondReadingOffs, d.secondReadingCfg.offset)
			assert.Equal(t, int32(1), d.secondReadingCfg.calibrationFactorNum)
			assert.Equal(t, int32(1), d.secondReadingCfg.calibrationFactorDen)
		})
	}
}

func TestCalibrate(t *testing.T) {
	tests := map[string]struct {
		setValue                int32
		secondReading           bool
		simGainAndChannelCfg    GainAndChannelCfg
		simPinDtaGetValues      []bool
		simOffset               int32
		simClkPinSetErr         error
		wantFirstReadingCalFac  float32
		wantSecondReadingCalFac float32
		wantErr                 string
	}{
		"calibrate_first": {
			setValue:                123 * 2,
			simGainAndChannelCfg:    A128,
			simPinDtaGetValues:      readyValues123,
			wantFirstReadingCalFac:  2,
			wantSecondReadingCalFac: 1,
		},
		"calibrate_second": {
			setValue:                7654321 * 3,
			secondReading:           true,
			simGainAndChannelCfg:    A128B32,
			simPinDtaGetValues:      append(readyValues123, readyValues7654321...),
			wantFirstReadingCalFac:  1,
			wantSecondReadingCalFac: 3,
		},
		"result_near_zero": {
			// this can happen, if the set value is much smaller than the read value, practically this can be prevented by
			// selecting another unit for the set value (and measurement), e.g. use 0.1 gram instead of 1e-07 tons
			setValue:                1,
			simGainAndChannelCfg:    A128,
			simPinDtaGetValues:      readyValues24bitMax,
			simOffset:               math.MaxInt32,
			wantFirstReadingCalFac:  -4.6748744e-10,
			wantSecondReadingCalFac: 1,
		},
		"error_calibrate_no_second": {
			secondReading:           true,
			simGainAndChannelCfg:    A128,
			wantFirstReadingCalFac:  1,
			wantSecondReadingCalFac: 1,
			wantErr:                 "no calibration possible, second reading is not configured",
		},
		"error_calibrate_read": {
			simPinDtaGetValues:      []bool{false},
			simClkPinSetErr:         errors.New("set clk on calibrate err"),
			simGainAndChannelCfg:    A128,
			wantFirstReadingCalFac:  1,
			wantSecondReadingCalFac: 1,
			wantErr:                 "set clk on calibrate err",
		},
		"error_setvalue_zero": {
			setValue:                0,
			simGainAndChannelCfg:    A128,
			simPinDtaGetValues:      readyValues123,
			wantFirstReadingCalFac:  1,
			wantSecondReadingCalFac: 1,
			wantErr:                 "set value needs to be <> 0",
		},
		"error_read_exactly_zero": {
			setValue:                1,
			simGainAndChannelCfg:    A128,
			simPinDtaGetValues:      readyValuesZero,
			wantFirstReadingCalFac:  1,
			wantSecondReadingCalFac: 1,
			wantErr:                 "read value is exactly 0",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			pClkMock := outputPinMock{
				setErr: tc.simClkPinSetErr,
			}
			pClk := newOutPin(&pClkMock)
			pDtaMock := inputPinMock{getSimReturn: tc.simPinDtaGetValues}
			pDta := newInPin(&pDtaMock)
			d := New(pClk, pDta, tc.simGainAndChannelCfg)
			d.devCfg = fastCfg
			d.firstReadingCfg.offset = tc.simOffset
			// act
			err := d.Calibrate(tc.setValue, tc.secondReading)
			// assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, len(tc.simPinDtaGetValues), pDtaMock.getCalled)
			gotFirstReadingCalFac := d.firstReadingCfg.calibrationFactorNum / d.firstReadingCfg.calibrationFactorDen
			assert.InDelta(t, tc.wantFirstReadingCalFac, gotFirstReadingCalFac, 0.00001)
			gotSecondReadingCalFac := d.secondReadingCfg.calibrationFactorNum / d.secondReadingCfg.calibrationFactorDen
			assert.InDelta(t, tc.wantSecondReadingCalFac, gotSecondReadingCalFac, 0.00001)
		})
	}
}

func TestUpdate_Values(t *testing.T) {
	tests := map[string]struct {
		simGainAndChannelCfg GainAndChannelCfg
		simPinDtaGetValues   []bool
		simClkPinSetErr      error
		wantFirstReadingRaw  float32
		wantSecondReadingRaw float32
		wantFirstReading     float32
		wantSecondReading    float32
		wantErr              string
	}{
		"update_single": {
			simGainAndChannelCfg: A128,
			simPinDtaGetValues:   readyValues123,
			wantFirstReadingRaw:  123,
			wantFirstReading:     (123 - 5000) * 2,
			wantSecondReading:    4000, // because we set the offset to 1000 below
		},
		"update_double": {
			simGainAndChannelCfg: A128B32,
			simPinDtaGetValues:   append(readyValues7654321, readyValues123...),
			wantFirstReadingRaw:  7654321,
			wantSecondReadingRaw: 123,
			wantFirstReading:     (7654321 - 5000) * 2,
			wantSecondReading:    (123 + 1000) * 4,
		},
		"error_update_read": {
			simGainAndChannelCfg: A128,
			simPinDtaGetValues:   []bool{false},
			simClkPinSetErr:      errors.New("set clk on update err"),
			wantFirstReading:     -5000 * 2,
			wantSecondReading:    +1000 * 4,
			wantErr:              "set clk on update err",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			pClkMock := outputPinMock{
				setErr: tc.simClkPinSetErr,
			}
			pClk := newOutPin(&pClkMock)
			pDtaMock := inputPinMock{getSimReturn: tc.simPinDtaGetValues}
			pDta := newInPin(&pDtaMock)
			d := New(pClk, pDta, tc.simGainAndChannelCfg)
			d.devCfg = fastCfg
			// small adjustments for calibration (raw values not affected)
			d.firstReadingCfg.offset = -5000
			d.firstReadingCfg.calibrationFactorNum = 2
			d.secondReadingCfg.offset = 1000
			d.secondReadingCfg.calibrationFactorNum = 4
			// act
			err := d.Update(0)
			got1, got2 := d.Values()
			// assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, len(tc.simPinDtaGetValues), pDtaMock.getCalled)
			assert.InDelta(t, tc.wantFirstReadingRaw, d.firstRawValue, 0.00001)
			assert.InDelta(t, tc.wantSecondReadingRaw, d.secondRawValue, 0.00001)
			assert.InDelta(t, tc.wantFirstReading, got1, 0.00001)
			assert.InDelta(t, tc.wantSecondReading, got2, 0.00001)
		})
	}
}

func TestOffsetAndCalibrationFactor_set_get(t *testing.T) {
	tests := map[string]struct {
		secondReading bool
		simOffs       int32
		simCalFac     float32
		wantOffs      int32
		wantCalFac    float32
		wantErr       string
	}{
		"set_first": {
			simOffs:    345,
			simCalFac:  567.89,
			wantOffs:   345,
			wantCalFac: 567.89,
		},
		"set_second": {
			secondReading: true,
			simOffs:       45,
			simCalFac:     67.89,
			wantOffs:      45,
			wantCalFac:    67.89,
		},
		"error_too_small": {
			simOffs:   5,
			simCalFac: float32(math.MinInt32) - 128.1, // float32 has a conversion error of 128 for "math.MinInt32"
			wantOffs:  0,                              // factor is NaN, because denominator not set
			wantErr:   "input value exceeds -int32 range",
		},
		"error_too_big": {
			simOffs:   45,
			simCalFac: float32(math.MaxInt32) - 64, // float32 has a conversion error of -64 for "math.MaxInt32",
			wantOffs:  0,                           // factor is NaN, because denominator not set
			wantErr:   "input value exceeds +int32 range",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			d := Device{mu: &sync.Mutex{}}
			// act
			err := d.SetOffsetAndCalibrationFactor(tc.simOffs, tc.simCalFac, tc.secondReading)
			gotOffs, gotCalFac := d.OffsetAndCalibrationFactor(tc.secondReading)
			// assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tc.wantCalFac, gotCalFac, 0)
			}
			assert.Equal(t, tc.wantOffs, gotOffs)
			if tc.secondReading {
				assert.Equal(t, int32(0), d.firstReadingCfg.offset)
				assert.Equal(t, int32(0), d.firstReadingCfg.calibrationFactorNum)
				assert.Equal(t, int32(0), d.firstReadingCfg.calibrationFactorDen)
			} else {
				assert.Equal(t, int32(0), d.secondReadingCfg.offset)
				assert.Equal(t, int32(0), d.secondReadingCfg.calibrationFactorNum)
				assert.Equal(t, int32(0), d.secondReadingCfg.calibrationFactorDen)
			}
		})
	}
}

type outputPinMock struct {
	setCalled []bool
	setErr    error
}

func newOutPin(opm *outputPinMock) pinSetState {
	return opm.setState
}

func (opm *outputPinMock) setState(val bool) error {
	opm.setCalled = append(opm.setCalled, val)
	return opm.setErr
}

type inputPinMock struct {
	getCalled    int
	getSimReturn []bool
	getErr       error
}

func newInPin(ipm *inputPinMock) pinState {
	return ipm.state
}

func (ipm *inputPinMock) state() (bool, error) {
	ipm.getCalled++
	if len(ipm.getSimReturn) < ipm.getCalled {
		return false, errors.New("error get, no value")
	}

	return ipm.getSimReturn[ipm.getCalled-1], ipm.getErr
}
