// Package bme68x provides a driver for the BME680/BME688 low power gas,
// pressure, temperature and humidity sensor by Bosch.
//
// Datasheet:
// https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bme680-ds001.pdf
package bme68x

import (
	"errors"
	"fmt"
	"math"
	"time"

	"tinygo.org/x/drivers"
)

const (
	// The default I2C address which this device listens to.
	Address = 0x77
	// AmbientTemperature is the ambient temperature in deg C used for defining the heater temperature.
	AmbientTemperature = 25.0
	// TargetTemperature is the target temperature in deg C used for defining the heater temperature.
	TargetTemperature uint16 = 320
	// TargetHeatrDuration is the target heater duration in ms used for defining the heater temperature.
	TargetHeatrDuration uint16 = 150
	// MeasOffset is the offset for the measurement phase
	MeasOffset uint32 = 1963
	// MeasDur is the duration of the TPH switching phase
	MeasDur uint32 = 1908
	// GasDur is the duration of the gas measurement phase
	GasDur uint32 = 2385
	// WakeUpDur is the duration of the wake up phase in ms
	WakeUpDur = 1000
	// MaxDuration is the maximum duration for gas wait
	MaxDuration = 0xFF
	// PeriodPoll is thed default period for polling the sensor in µs
	PeriodPoll uint32 = 10000
	// PeriodReset is the period for resetting the sensor in µs
	PeriodReset uint32 = 10000 // 10000 µs
)

var (
	osToMeasCycles = [6]uint8{0, 1, 2, 4, 8, 16}
	lookupK1Range  = [16]float32{
		0.0, 0.0, 0.0, 0.0, 0.0, -1.0, 0.0, -0.8, 0.0, 0.0, -0.2, -0.5, 0.0, -1.0, 0.0, 0.0,
	}
	lookupK2Range = [16]float32{
		0.0, 0.0, 0.0, 0.0, 0.1, 0.7, 0.0, -0.8, -0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}
)

type (
	// calibrationCoefficients reads at startup and stores the calibration coefficients.
	calibrationCoefficients struct {
		// temperature related coefficients
		t1 uint16
		t2 int16
		t3 int8

		// pressure related coefficients
		p1  uint16
		p2  int16
		p3  int8
		p4  int16
		p5  int16
		p6  int8
		p7  int8
		p8  int16
		p9  int16
		p10 int8

		// humidity related coefficients
		h1 uint16
		h2 uint16
		h3 int8
		h4 int8
		h5 int8
		h6 uint8
		h7 int8

		// gas related coefficients
		g1 int8
		g2 int16
		g3 int8

		// other coefficients
		resHeatRange uint8
		resHeatVal   int8
		rangeSwErr   int8
	}

	Config struct {
		Pressure    Oversampling
		Temperature Oversampling
		Humidity    Oversampling
		IIR         FilterCoefficient
		ODR         ODR
		// HeatrTemp is the target temperature in degree Celsius.
		HeatrTemp uint16
		// HeatrDur is the gas wait period.
		HeatrDur uint16
		// HeatrEnable enables gas measurement.
		HeatrEnable        bool
		AmbientTemperature int8
		PeriodPoll         uint32
		mode               Mode
	}

	Device struct {
		bus                     bus
		address                 uint16
		chipID                  byte
		calibrationCoefficients calibrationCoefficients
		config                  *Config
		measStart               int64
		measPeriod              uint16

		// Status contains new_data, gasm_valid and heat_stab bits.
		Status byte
		// GasIndex is the index of the heater profile used.
		GasIndex uint8
		// MeasIndex is the measurement index to track order.
		MeasIndex uint8
		// ResHeat is the heater resistance.
		ResHeat uint8
		// GasWait is the gas wait period.
		GasWait uint8
		// TemperatureFine is the intermediate temperature coefficient.
		TemperatureFine float32
		// Temperature is the temperature in degree Celsius.
		Temperature float32
		// Pressure is the pressure in Pascal.
		Pressure float32
		// Humidity is the relative humidity in percent (x1000).
		Humidity float32
		// GasResistance is the gas resistance in Ohms.
		GasResistance float32
		// Idac is the current DAC
		Idac uint8
		// VariantID is the variant ID.
		VariantID uint8
	}

	// bus is the interface for the I2C and SPI bus.
	bus interface {
		Reset(addr uint16) error
		Read(addr uint16, reg uint8, data []byte) error
		Write(addr uint16, reg []uint8, data []byte) error
	}
)

// NewI2C creates a new BME68x connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func NewI2C(bus drivers.I2C, opts ...Option) *Device {
	return new(&i2c{
		bus: bus,
	}, opts...)
}

// NewSPI creates a new BME68x connection. The SPI bus must already be
// configured. It also requires a CS pin to be used as the chip select.
//
// This function only creates the Device object, it does not touch the device.
func NewSPI(bus drivers.SPI, opts ...Option) *Device {
	return new(&spi{
		bus: bus,
	}, opts...)
}

func new(bus bus, opts ...Option) *Device {
	device := &Device{
		address: Address,
		bus:     bus,
		config: &Config{
			mode:               ModeForced,
			Temperature:        Sampling8X,
			Humidity:           Sampling2X,
			Pressure:           Sampling4X,
			IIR:                Coeff4,
			ODR:                ODR_NONE,
			HeatrEnable:        true,
			HeatrTemp:          TargetTemperature,
			HeatrDur:           TargetHeatrDuration,
			AmbientTemperature: AmbientTemperature,
			PeriodPoll:         PeriodPoll,
		},
	}

	for _, option := range opts {
		option(device)
	}

	return device
}

// Configure sets up the device for communication.
func (d *Device) Configure() error {
	connected, err := d.Connected()
	if err != nil {
		return fmt.Errorf("device not found or not connected: %w", err)
	}

	if !connected {
		return errors.New("device not found or not connected")
	}

	if err := d.Reset(); err != nil {
		return fmt.Errorf("failed to reset device: %w", err)
	}

	if err := d.readChipID(); err != nil {
		return fmt.Errorf("failed to read chip ID: %w", err)
	}

	if err := d.readVariantID(); err != nil {
		return fmt.Errorf("failed to read variant ID: %w", err)
	}

	if err := d.readCalibrationData(); err != nil {
		return fmt.Errorf("failed to read calibration data: %w", err)
	}

	if err := d.applyConfig(); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}

	if err := d.applyGasConfig(); err != nil {
		return fmt.Errorf("failed to apply gas config: %w", err)
	}

	return nil
}

func (d *Device) readChipID() error {
	var data [1]byte
	if err := d.bus.Read(d.address, REG_CHIP_ID, data[:]); err != nil {
		return err
	}
	d.chipID = data[0]
	return nil
}

func (d *Device) readVariantID() error {
	var data [1]byte
	if err := d.bus.Read(d.address, REG_VARIANT_ID, data[:]); err != nil {
		return err
	}
	d.VariantID = data[0]
	return nil
}

func (d *Device) readCalibrationData() error {
	var data [42]byte

	// read the calibration data
	if err := d.bus.Read(d.address, REG_COEFF1, data[:23]); err != nil {
		return err
	}
	if err := d.bus.Read(d.address, REG_COEFF2, data[23:37]); err != nil {
		return err
	}
	if err := d.bus.Read(d.address, REG_COEFF3, data[37:]); err != nil {
		return err
	}

	// temperature related coefficients
	d.calibrationCoefficients.t1 = parseByte[uint16](data[32], data[31])
	d.calibrationCoefficients.t2 = parseByte[int16](data[1], data[0])
	d.calibrationCoefficients.t3 = int8(data[2])

	// pressure related coefficients
	d.calibrationCoefficients.p1 = parseByte[uint16](data[5], data[4])
	d.calibrationCoefficients.p2 = parseByte[int16](data[7], data[6])
	d.calibrationCoefficients.p3 = int8(data[8])
	d.calibrationCoefficients.p4 = parseByte[int16](data[11], data[10])
	d.calibrationCoefficients.p5 = parseByte[int16](data[13], data[12])
	d.calibrationCoefficients.p6 = int8(data[15])
	d.calibrationCoefficients.p7 = int8(data[14])
	d.calibrationCoefficients.p8 = parseByte[int16](data[19], data[18])
	d.calibrationCoefficients.p9 = parseByte[int16](data[21], data[20])
	d.calibrationCoefficients.p10 = int8(data[22])

	// humidity related coefficients
	d.calibrationCoefficients.h1 = uint16(data[25])<<4 | uint16(data[24])&0x0F
	d.calibrationCoefficients.h2 = uint16(data[23])<<4 | uint16(data[24])>>4
	d.calibrationCoefficients.h3 = int8(data[26])
	d.calibrationCoefficients.h4 = int8(data[27])
	d.calibrationCoefficients.h5 = int8(data[28])
	d.calibrationCoefficients.h6 = data[29]
	d.calibrationCoefficients.h7 = int8(data[30])

	// gas heater related coefficients
	d.calibrationCoefficients.g1 = int8(data[35])
	d.calibrationCoefficients.g2 = parseByte[int16](data[34], data[33])
	d.calibrationCoefficients.g3 = int8(data[36])

	// other coefficients
	d.calibrationCoefficients.resHeatRange = (data[39] & 0x30) / 16
	d.calibrationCoefficients.resHeatVal = int8(data[37])
	d.calibrationCoefficients.rangeSwErr = int8(data[41]&0xF0) / 16

	return nil
}

// Reset does a soft reset by writing 0xB6 to the reset register.
func (d *Device) Reset() error {
	return d.bus.Reset(d.address)
}

// Connected checks if the device is connected by reading the chip ID.
// It returns true if the chip ID matches the expected value.
func (d *Device) Connected() (bool, error) {
	if err := d.readChipID(); err != nil {
		return false, err
	}

	return d.chipID == CHIP_ID, nil
}

// Mode returns the current mode of the sensor.
func (d *Device) Mode() (Mode, error) {
	var data [1]byte
	if err := d.bus.Read(d.address, REG_CTRL_MEAS, data[:]); err != nil {
		return ModeSleep, err
	}

	d.config.mode = Mode(data[0] & MODE_MSK)

	return d.config.mode, nil
}

// SetMode sets the mode of the sensor.
func (d *Device) SetMode(mode Mode) error {
	d.config.mode = mode

	var (
		tmpPowerMode [1]byte
		powerMode    byte
	)

	for ok := true; ok; ok = (powerMode != byte(ModeSleep)) {
		// read the current power mode
		if err := d.bus.Read(d.address, REG_CTRL_MEAS, tmpPowerMode[:]); err != nil {
			return err
		}

		// put to sleep before changing mode
		powerMode = (tmpPowerMode[0] & MODE_MSK)
		if powerMode != byte(ModeSleep) {
			tmpPowerMode[0] &= ^MODE_MSK
			if err := d.bus.Write(d.address, []uint8{REG_CTRL_MEAS}, tmpPowerMode[:]); err != nil {
				return err
			}

			time.Sleep(time.Duration(d.config.PeriodPoll) * time.Microsecond)
		}
	}

	// already in sleep
	if mode != ModeSleep {
		tmpPowerMode[0] = (tmpPowerMode[0] & ^MODE_MSK) | (byte(mode) & MODE_MSK)
		if err := d.bus.Write(d.address, []uint8{REG_CTRL_MEAS}, tmpPowerMode[:]); err != nil {
			return err
		}
	}

	return nil
}

// SetTemperatureOversampling sets the temperature oversampling.
func (d *Device) SetTemperatureOversampling(os Oversampling) error {
	d.config.Temperature = os

	if err := d.applyConfig(); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}

	return nil
}

// SetPressureOversampling sets the pressure oversampling.
func (d *Device) SetPressureOversampling(os Oversampling) error {
	d.config.Pressure = os

	if err := d.applyConfig(); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}

	return nil
}

// SetHumidityOversampling sets the humidity oversampling.
func (d *Device) SetHumidityOversampling(os Oversampling) error {
	d.config.Humidity = os

	if err := d.applyConfig(); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}

	return nil
}

// SetIIRFilter sets the IIR filter coefficient.
func (d *Device) SetIIRFilter(fc FilterCoefficient) error {
	d.config.IIR = fc

	if err := d.applyConfig(); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}

	return nil
}

// SetODR sets the ODR (output data rate) of the sensor.
func (d *Device) SetODR(odr ODR) error {
	d.config.ODR = odr

	if err := d.applyConfig(); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}

	return nil
}

func (d *Device) SetGasHeater(temp, dur uint16, enable bool) error {
	d.config.HeatrTemp = temp
	d.config.HeatrDur = dur
	d.config.HeatrEnable = enable

	if err := d.applyGasConfig(); err != nil {
		return fmt.Errorf("failed to apply gas config: %w", err)
	}

	return nil
}

// applyConfig sets oversampling and filter configuration.
func (d *Device) applyConfig() error {
	currentMode, err := d.Mode()
	if err != nil {
		return err
	}

	// configure only in the sleep mode
	if err := d.SetMode(ModeSleep); err != nil {
		return err
	}

	// read the current configuration
	var data [5]byte
	if err := d.bus.Read(d.address, REG_CTRL_GAS_1, data[:]); err != nil {
		return err
	}

	// set bits
	data[4] = (data[4] & ^FILTER_MSK) | ((byte(d.config.IIR) << FILTER_POS) & FILTER_MSK)
	data[3] = (data[3] & ^OST_MSK) | ((byte(d.config.Temperature) << OST_POS) & OST_MSK)
	data[3] = (data[3] & ^OSP_MSK) | ((byte(d.config.Pressure) << OSP_POS) & OSP_MSK)
	data[1] = (data[1] & ^OSH_MSK) | (byte(d.config.Humidity) & OSH_MSK)

	var odr20 ODR
	odr3 := 1

	if d.config.ODR != ODR_NONE {
		odr20 = d.config.ODR
		odr3 = 0
	}

	data[4] = (data[4] & ^ODR20_MSK) | ((byte(odr20) << ODR20_POS) & ODR20_MSK)
	data[0] = (data[0] & ^ODR3_MSK) | ((byte(odr3) << ODR3_POS) & ODR3_MSK)

	// write the new configuration
	// register data starting from REG_CTRL_GAS_1(0x71) up to REG_CONFIG(0x75)
	if err := d.bus.Write(
		d.address,
		[]uint8{REG_CTRL_GAS_1, REG_CTRL_HUM, 0x73, REG_CTRL_MEAS, REG_CONFIG},
		data[:],
	); err != nil {
		return err
	}

	// restore the previous mode
	if currentMode != ModeSleep {
		if err := d.SetMode(currentMode); err != nil {
			return err
		}
	}

	return nil
}

// applyGasConfig sets the gas configuration of the sensor.
func (d *Device) applyGasConfig() error {
	if d.config.HeatrTemp == 0 || d.config.HeatrDur == 0 {
		d.config.HeatrEnable = false
	}

	// configure only in the sleep mode
	if err := d.SetMode(ModeSleep); err != nil {
		return err
	}

	if err := d.applyHeatrConfig(); err != nil {
		return err
	}

	var hctrl, runGas byte
	var ctrlGasData [2]byte
	var nbConv byte = 0

	// read the current configuration
	if err := d.bus.Read(d.address, REG_CTRL_GAS_0, ctrlGasData[:]); err != nil {
		return err
	}

	if d.config.HeatrEnable {
		hctrl = ENABLE_HEATER

		if d.VariantID == VARIANT_GAS_HIGH {
			runGas = ENABLE_GAS_MEAS_H
		} else {
			runGas = ENABLE_GAS_MEAS_L
		}
	} else {
		hctrl = DISABLE_HEATER
		runGas = DISABLE_GAS_MEAS
	}

	ctrlGasData[0] = (ctrlGasData[0] & ^HCTRL_MSK) | ((hctrl << HCTRL_POS) & HCTRL_MSK)
	ctrlGasData[1] = (ctrlGasData[1] & ^NBCONV_MSK) | (nbConv & NBCONV_MSK)
	ctrlGasData[1] = (ctrlGasData[1] & ^RUN_GAS_MSK) | ((runGas << RUN_GAS_POS) & RUN_GAS_MSK)

	// write the new configuration
	if err := d.bus.Write(d.address, []uint8{REG_CTRL_GAS_0, REG_CTRL_GAS_1}, ctrlGasData[:]); err != nil {
		return err
	}

	return nil
}

// applyHeatrConfig sets the heater configurations.
func (d *Device) applyHeatrConfig() error {
	rhRegData := make([]uint8, 1)
	gwRegData := make([]uint8, 1)

	rhRegData[0] = d.calcResistanceHeat(d.config.HeatrTemp)
	gwRegData[0] = d.calcGasWait()

	// write the new configuration
	if err := d.bus.Write(d.address, []uint8{REG_RES_HEAT0}, rhRegData[:]); err != nil {
		return err
	}

	if err := d.bus.Write(d.address, []uint8{REG_GAS_WAIT0}, gwRegData[:]); err != nil {
		return err
	}

	return nil
}

// Read reads all sensor data and store it in the Device struct.
func (d *Device) Read() error {
	if d.measStart != 0 {
		return nil
	}

	if err := d.SetMode(ModeForced); err != nil {
		return fmt.Errorf("failed to set forced mode: %w", err)
	}

	// calculate delay period in microseconds
	delayusPeriod := d.calcMeasDuration() + (uint32(d.config.HeatrDur) * 1000)
	d.measStart = time.Now().UnixMilli()
	d.measPeriod = uint16(delayusPeriod) / 1000

	if d.measStart+int64(d.measPeriod) == 0 {
		return nil
	}

	remainingMillis := d.calRemainingReadingMillis()
	if remainingMillis > 0 {
		time.Sleep(time.Duration(remainingMillis*2) * time.Millisecond)
	}

	d.measStart = 0
	d.measPeriod = 0

	if err := d.readData(); err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	return nil
}

func (d *Device) readData() error {
	// try up to 5 times to read the data
	for i := uint8(0); i < 5; i++ {
		var data [17]byte
		if err := d.bus.Read(d.address, MEAS_STATUS_0+(i*17), data[:]); err != nil {
			return err
		}

		d.Status = data[0] & NEW_DATA_MSK
		d.GasIndex = data[0] & GAS_INDEX_MSK
		d.MeasIndex = data[1]

		// read the raw data from the sensor
		adcPres := uint32((uint32(data[2]) * 4096) | (uint32(data[3]) * 16) | (uint32(data[4]) / 16))
		adcTemp := uint32((uint32(data[5]) * 4096) | (uint32(data[6]) * 16) | (uint32(data[7]) / 16))
		adcHum := uint16((uint32(data[8]) * 256) | (uint32(data[9])))
		adcGasResLow := uint16(uint32(data[13])*4 | (uint32(data[14]) / 64))
		adcGasResHigh := uint16(uint32(data[15])*4 | (uint32(data[16]) / 64))
		gasRangeLow := data[14] & GAS_RANGE_MSK
		gasRangeHigh := data[16] & GAS_RANGE_MSK

		if d.VariantID == VARIANT_GAS_HIGH {
			d.Status |= data[16] & GASM_VALID_MSK
			d.Status |= data[16] & HEAT_STAB_MSK
		} else {
			d.Status |= data[14] & GASM_VALID_MSK
			d.Status |= data[14] & HEAT_STAB_MSK
		}

		// check if new data is available
		if d.Status&NEW_DATA_MSK != 0 {
			var resHeat [1]byte
			if err := d.bus.Read(d.address, REG_RES_HEAT0+d.GasIndex, resHeat[:]); err != nil {
				return err
			}
			d.ResHeat = resHeat[0]

			var idac [1]byte
			if err := d.bus.Read(d.address, REG_IDAC_HEAT0+d.GasIndex, idac[:]); err != nil {
				return err
			}
			d.Idac = idac[0]

			var gasWait [1]byte
			if err := d.bus.Read(d.address, REG_GAS_WAIT0+d.GasIndex, gasWait[:]); err != nil {
				return err
			}
			d.GasWait = gasWait[0]

			d.Temperature = d.calcTemperature(adcTemp)
			d.Pressure = d.calcPressure(adcPres)
			d.Humidity = d.calcHumidity(adcHum)

			// check if gas data is available
			if d.Status&(HEAT_STAB_MSK|GASM_VALID_MSK) != 0 {
				if d.VariantID == VARIANT_GAS_HIGH {
					d.GasResistance = d.calcGasResistanceHigh(adcGasResHigh, gasRangeHigh)
				} else {
					d.GasResistance = d.calcGasResistanceLow(adcGasResLow, gasRangeLow)
				}
			} else {
				d.GasResistance = 0
			}

			break
		}

		time.Sleep(time.Duration(d.config.PeriodPoll) * time.Microsecond)
	}

	return nil
}

func (d *Device) calcTemperature(adcTemp uint32) float32 {
	var1 := (((float32(adcTemp) / 16384) - (float32(d.calibrationCoefficients.t1) / 1024)) * float32(d.calibrationCoefficients.t2))
	var2 := ((((float32(adcTemp) / 131072) - (float32(d.calibrationCoefficients.t1) / 8192)) *
		((float32(adcTemp) / 131072) - (float32(d.calibrationCoefficients.t1) / 8192))) * (float32(d.calibrationCoefficients.t3) * 16))

	d.TemperatureFine = var1 + var2

	return d.TemperatureFine / 5120
}

func (d *Device) calcPressure(adcPres uint32) float32 {
	var1 := (float32(d.TemperatureFine)/2 - 64000)
	var2 := var1 * var1 * (float32(d.calibrationCoefficients.p6) / 131072)
	var2 += var1 * float32(d.calibrationCoefficients.p5) * 2
	var2 = (var2 / 4) + float32(d.calibrationCoefficients.p4)*65536
	var1 = (((float32(d.calibrationCoefficients.p3) * var1 * var1) / 16384) + (float32(d.calibrationCoefficients.p2) * var1)) / 524288
	var1 = (1.0 + (var1 / 32768)) * float32(d.calibrationCoefficients.p1)
	calcPres := 1048576 - float32(adcPres)

	// avoid division by zero
	if var1 == 0 {
		return 0
	}

	calcPres = ((calcPres - (var2 / 4096)) * 6250) / var1
	var1 = (float32(d.calibrationCoefficients.p9) * calcPres * calcPres) / 2147483648
	var2 = calcPres * (float32(d.calibrationCoefficients.p8) / 32768)
	var3 := ((calcPres / 256) * (calcPres / 256) * (calcPres / 256) * (float32(d.calibrationCoefficients.p10) / 131072))
	return calcPres + (var1+var2+var3+(float32(d.calibrationCoefficients.p7)*128))/16
}

func (d *Device) calcHumidity(adcHum uint16) float32 {
	tempComp := d.TemperatureFine / 5120.0
	var1 := float32(adcHum) - ((float32(d.calibrationCoefficients.h1) * 16.0) +
		((float32(d.calibrationCoefficients.h3) / 2.0) * tempComp))
	var2 := var1 * ((float32(d.calibrationCoefficients.h2) / 262144.0) *
		(1.0 + ((float32(d.calibrationCoefficients.h4) / 16384.0) * tempComp) +
			((float32(d.calibrationCoefficients.h5) / 1048576.0) * tempComp * tempComp)))
	var3 := float32(d.calibrationCoefficients.h6) / 16384.0
	var4 := float32(d.calibrationCoefficients.h7) / 2097152.0
	calcHum := var2 + ((var3 + (var4 * tempComp)) * var2 * var2)

	if calcHum > 100.0 {
		return 100.0
	}

	if calcHum < 0.0 {
		return 0.0
	}

	return calcHum
}

func (d *Device) calcGasResistanceLow(adcGasRes uint16, gasRange uint8) float32 {
	gasRangeF := float32(int(1) << gasRange)
	var1 := float32(1340.0 + (5.0 * float32(d.calibrationCoefficients.rangeSwErr)))
	var2 := var1 * (1.0 + lookupK1Range[gasRange]/100.0)
	var3 := 1.0 + (lookupK2Range[gasRange] / 100.0)

	return 1.0 / (var3 * (0.000000125) * gasRangeF * (((float32(adcGasRes) - 512.0) / var2) + 1.0))
}

func (d *Device) calcGasResistanceHigh(adcGasRes uint16, gasRange uint8) float32 {
	var1 := uint32(262144) >> gasRange
	var2 := int32(adcGasRes) - 512

	var2 *= 3
	var2 += 4096

	return 1000000.0 * float32(var1) / float32(var2)
}

// calcMeasDuration calculates the remaining duration that can be used for heating.
func (d *Device) calcMeasDuration() uint32 {
	measCycles := osToMeasCycles[d.config.Temperature]
	measCycles += osToMeasCycles[d.config.Pressure]
	measCycles += osToMeasCycles[d.config.Humidity]

	// TPHG measurement duration
	dur := uint32(measCycles) * MeasOffset
	dur += MeasDur
	dur += GasDur
	dur += WakeUpDur // wake up duration of 1ms

	return dur
}

func (d *Device) calRemainingReadingMillis() int64 {
	if d.measStart == 0 {
		return -1
	}

	remaingTime := int64(d.measPeriod) - (time.Now().UnixMilli() - d.measStart)

	if remaingTime < 0 {
		return 0
	}

	return remaingTime
}

// calcGasWait calculates the gas wait period. It takes the heater duration
// in ms and returns the calculated gas wait period.
func (d *Device) calcGasWait() uint8 {
	var factor uint8
	dur := d.config.HeatrDur

	if dur >= 0xFC0 {
		return MaxDuration
	}

	for dur > 0x3F {
		dur /= 4
		factor++
	}

	return uint8(uint8(dur) + (factor * 64))
}

// calcResistanceHeat calculates the heater resistance value. It takes the target
// temperature in degree Celsius and returns the calculated heater resistance
// value.
func (d *Device) calcResistanceHeat(target uint16) uint8 {
	// cap temperature to 400°C
	if target > 400 {
		target = 400
	}

	var1 := (float32(d.calibrationCoefficients.g1) / 16.0) + 49
	var2 := ((float32(d.calibrationCoefficients.g2) / 32768.0) * 0.0005) + 0.00235
	var3 := float32(d.calibrationCoefficients.g3) / 1024
	var4 := var1 * (1 + (var2 * float32(target)))
	var5 := var4 + (var3 * float32(d.config.AmbientTemperature))

	return uint8(3.4 *
		((var5 * (4 / (4 + float32(d.calibrationCoefficients.resHeatRange))) *
			(1 / (1 + (float32(d.calibrationCoefficients.resHeatVal) * 0.002)))) -
			25))
}

// CalcAltitude calculates the altitude in meters based on the sea level
// pressure and the current pressure. It uses the barometric formula to
// calculate the altitude. The sea level pressure is usually 1013.25 hPa.
func CalcAltitude(seaLevel, pressure float32) float64 {
	// Equation taken from BMP180 datasheet (page 16):
	// http://www.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf

	// Note that using the equation from wikipedia can give bad results
	// at high altitude. See this thread for more information:
	// http://forums.adafruit.com/viewtopic.php?f=22&t=58064

	atmospheric := pressure / 100

	return 44330.0 * (1.0 - math.Pow(float64(atmospheric/seaLevel), 0.1903))
}

// Config returns the current configuration of the sensor.
func (d *Device) Config() Config {
	return *d.config
}

// parseByte converts two bytes to T16.
func parseByte[T uint16 | int16](msb, lsb byte) T {
	return (T(msb) << 8) | T(lsb)
}

// String implements fmt.Stringer interface.
func (c Config) String() string {
	return fmt.Sprintf("pressure: %d, temperature: %d, humidity: %d, iir: %d, odr: %d, heatrTemp: %d°C, heatrDur: %dms,"+
		" heatrEnable: %t, ambientTemperature: %d, mode: %d",
		c.Pressure, c.Temperature, c.Humidity, c.IIR, c.ODR, c.HeatrTemp, c.HeatrDur, c.HeatrEnable, c.AmbientTemperature,
		c.mode,
	)
}

// String implements fmt.Stringer interface.
func (c calibrationCoefficients) String() string {
	return fmt.Sprintf(`temperature: t1: %d, t2: %d, t3: %d, pressure: p1: %d, p2: %d, p3: %d, p4: %d,
	p5: %d, p6: %d, p7: %d, p8: %d, p9: %d, p10: %d, humidity: h1: %d, h2: %d, h3: %d, h4: %d, h5: %d,
  	h6: %d, h7: %d, gas: g1: %d, g2: %d, g3: %d, res_heat_range: %d, res_heat_val: %d, range_sw_err: %d`,
		c.t1, c.t2, c.t3, c.p1, c.p2, c.p3, c.p4,
		c.p5, c.p6, c.p7, c.p8, c.p9, c.p10, c.h1, c.h2, c.h3, c.h4, c.h5,
		c.h6, c.h7, c.g1, c.g2, c.g3, c.resHeatRange, c.resHeatVal, c.rangeSwErr,
	)
}

// String implements fmt.Stringer interface.
func (d Device) String() string {
	return fmt.Sprintf("address: 0x%X, chip id: 0x%X, variant id: 0x%X, status: 0x%X,"+
		" temperature fine:%.2f, temperature: %.2f°C, pressure: %.2fPa, humidity: %.2f%%,"+
		" res gas: %.2fΩ, res heat: %dΩ, gas wait: %dms, idac: %d",
		d.address, d.chipID, d.VariantID, d.Status, d.TemperatureFine, d.Temperature,
		d.Pressure, d.Humidity, d.GasResistance, d.ResHeat, d.GasWait, d.Idac,
	)
}
