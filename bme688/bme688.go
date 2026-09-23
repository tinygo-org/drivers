// Package bme688 provides a driver for the BME688 digital environmental sensor
// by Bosch, which measures temperature, pressure, humidity and gas (VOC).
//
// # It is available in the Pimoroni Enviro indoor board
//
// Datasheet: https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bme688-ds000.pdf
// Pimoroni Driver
//
//	https://github.com/pimoroni/pimoroni-pico/tree/main/examples/breakout_bme688
package bme688

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// calibrationCoefficients holds the device calibration data read at startup.
type calibrationCoefficients struct {
	// Temperature
	t1 uint16
	t2 int16
	t3 int8
	// Pressure
	p1  uint16
	p2  int16
	p3  int8
	p4  int16
	p5  int16
	p6  int8
	p7  int8
	p8  int16
	p9  int16
	p10 uint8
	// Humidity
	h1 uint16
	h2 uint16
	h3 int8
	h4 int8
	h5 int8
	h6 uint8
	h7 int8
	// Gas heater
	gh1          int8
	gh2          int16
	gh3          int8
	resHeatRange uint8
	resHeatVal   int8
	rangeSWErr   int8
	// Running fine temperature value shared by T/P/H calculations
	tFine int32
}

// Config holds sensor configuration settings.
type Config struct {
	Temperature Oversampling
	Pressure    Oversampling
	Humidity    Oversampling
	IIR         Filter
	// Gas heater settings
	GasEnabled     bool
	HeaterTemp     uint16 // target heater temperature in °C (max 400)
	HeaterDuration uint16 // heater-on duration in ms
	AmbientTemp    int8   // ambient temperature for heater resistance calculation
}

// Measurement holds a complete set of sensor readings.
type Measurement struct {
	Temperature   int32  // milli-degrees Celsius (e.g. 25000 = 25.000 °C)
	Pressure      uint32 // Pascals
	Humidity      uint32 // relative humidity × 1000 (e.g. 50000 = 50.000 %)
	GasResistance uint32 // Ohms
	GasValid      bool   // true when gas reading is valid
	HeaterStable  bool   // true when heater reached target temperature
}

// Device wraps an I2C connection to a BME688 (or BME680) device.
type Device struct {
	bus       drivers.I2C
	Address   uint16
	variantID byte
	calib     calibrationCoefficients
	Config    Config
	// Cached values from the last Update call.
	lastTemp          int32
	lastPressure      uint32
	lastHumidity      uint32
	lastGasResistance uint32
	lastGasValid      bool
	lastHeaterStable  bool
}

// New creates a new BME688 driver. The I2C bus must already be configured.
// This only creates the Device object; it does not communicate with the sensor.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure sets up the sensor with default settings and reads calibration data.
// Defaults: 2× oversampling for T/P/H, no IIR filter, gas enabled at 320 °C / 150 ms.
func (d *Device) Configure() {
	d.ConfigureWithSettings(Config{})
}

// ConfigureWithSettings sets up the sensor with the provided configuration.
// Zero-value fields revert to the defaults described in Configure.
func (d *Device) ConfigureWithSettings(config Config) {
	d.Config = config
	if d.Config == (Config{}) {
		d.Config = Config{
			Temperature:    Sampling2X,
			Pressure:       Sampling2X,
			Humidity:       Sampling2X,
			IIR:            FilterOff,
			GasEnabled:     true,
			HeaterTemp:     320,
			HeaterDuration: 150,
			AmbientTemp:    25,
		}
	}
	if d.Config.AmbientTemp == 0 {
		d.Config.AmbientTemp = 25
	}

	// Read variant ID to handle BME680 vs BME688 run_gas bit difference.
	var vbuf [1]byte
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_VARIANT_ID, vbuf[:])
	d.variantID = vbuf[0]

	d.Reset()
	time.Sleep(10 * time.Millisecond)

	d.readCalibration()

	// Apply IIR filter setting.
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_CONFIG,
		[]byte{byte(d.Config.IIR) << 2})

	// Apply humidity oversampling.
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL_HUM,
		[]byte{byte(d.Config.Humidity)})

	if d.Config.GasEnabled {
		// Enable heater (clear heat_off bit).
		legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL_GAS_0, []byte{0x00})

		// Write heater resistance and wait duration for profile 0.
		resHeat := d.calcResHeat(d.Config.HeaterTemp)
		gasWait := calcGasWait(d.Config.HeaterDuration)
		legacy.WriteRegister(d.bus, uint8(d.Address), REG_RES_HEAT_0, []byte{resHeat})
		legacy.WriteRegister(d.bus, uint8(d.Address), REG_GAS_WAIT_0, []byte{gasWait})

		// Enable gas measurement on profile 0.
		var runGasBit byte = RUN_GAS_688
		if d.variantID == VARIANT_BME680 {
			runGasBit = RUN_GAS_680
		}
		legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL_GAS_1,
			[]byte{runGasBit | 0x00}) // nb_conv = 0
	} else {
		// Disable heater.
		legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL_GAS_0,
			[]byte{HEAT_OFF_MSK})
	}
}

// Connected returns whether the sensor responds with the expected chip ID.
func (d *Device) Connected() bool {
	var buf [1]byte
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_CHIP_ID, buf[:])
	return buf[0] == CHIP_ID
}

// Reset performs a soft reset of the device.
func (d *Device) Reset() {
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_RESET, []byte{RESET_CMD})
}

// Read triggers a forced-mode measurement and returns all sensor values.
func (d *Device) Read() (Measurement, error) {
	// Set ctrl_meas: oversampling settings + forced mode trigger.
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL_MEAS, []byte{
		byte(d.Config.Temperature)<<5 |
			byte(d.Config.Pressure)<<2 |
			byte(ModeForced),
	})

	// Wait for the measurement to complete.
	time.Sleep(d.measurementDelay())

	// Poll for new data (up to ~50 extra ms in case of slight timing variance).
	var field [LEN_FIELD]byte
	for i := 0; i < 10; i++ {
		legacy.ReadRegister(d.bus, uint8(d.Address), REG_FIELD_0, field[:])
		if field[0]&NEW_DATA_MSK != 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	return d.parseField(field), nil
}

// ReadTemperature returns temperature in milli-degrees Celsius.
func (d *Device) ReadTemperature() (int32, error) {
	if err := d.Update(drivers.Temperature); err != nil {
		return 0, err
	}
	return d.Temperature(), nil
}

// ReadPressure returns pressure in Pascals.
func (d *Device) ReadPressure() (uint32, error) {
	if err := d.Update(drivers.Pressure); err != nil {
		return 0, err
	}
	return d.Pressure(), nil
}

// ReadHumidity returns relative humidity × 1000 (e.g. 50000 = 50.000 %).
func (d *Device) ReadHumidity() (uint32, error) {
	if err := d.Update(drivers.Humidity); err != nil {
		return 0, err
	}
	return d.Humidity(), nil
}

// ReadGasResistance returns gas resistance in Ohms.
// GasValid and HeaterStable should be checked before trusting this value.
func (d *Device) ReadGasResistance() (uint32, error) {
	if err := d.Update(drivers.Concentration); err != nil {
		return 0, err
	}
	return d.GasResistance(), nil
}

// Update performs a forced-mode measurement and caches the results.
// It satisfies the drivers.Sensor interface.
// The BME688 always measures temperature, pressure, humidity, and gas together,
// so a full read is performed regardless of which measurements are requested.
func (d *Device) Update(which drivers.Measurement) error {
	m, err := d.Read()
	if err != nil {
		return err
	}
	if which&drivers.Temperature != 0 {
		d.lastTemp = m.Temperature
	}
	if which&drivers.Pressure != 0 {
		d.lastPressure = m.Pressure
	}
	if which&drivers.Humidity != 0 {
		d.lastHumidity = m.Humidity
	}
	if which&drivers.Concentration != 0 {
		d.lastGasResistance = m.GasResistance
		d.lastGasValid = m.GasValid
		d.lastHeaterStable = m.HeaterStable
	}
	return nil
}

// Temperature returns the last measured temperature in milli-degrees Celsius.
func (d *Device) Temperature() int32 { return d.lastTemp }

// Pressure returns the last measured pressure in Pascals.
func (d *Device) Pressure() uint32 { return d.lastPressure }

// Humidity returns the last measured relative humidity × 1000 (e.g. 50000 = 50.000 %).
func (d *Device) Humidity() uint32 { return d.lastHumidity }

// GasResistance returns the last measured gas resistance in Ohms.
func (d *Device) GasResistance() uint32 { return d.lastGasResistance }

// GasValid reports whether the last gas reading was valid.
func (d *Device) GasValid() bool { return d.lastGasValid }

// HeaterStable reports whether the heater reached its target temperature during the last reading.
func (d *Device) HeaterStable() bool { return d.lastHeaterStable }

// ------------------------------------------------------------------ internal

// readCalibration reads all three calibration coefficient blocks from the device.
func (d *Device) readCalibration() {
	var c1 [LEN_COEFF_1]byte
	var c2 [LEN_COEFF_2]byte
	var c3 [LEN_COEFF_3]byte

	legacy.ReadRegister(d.bus, uint8(d.Address), REG_COEFF_1, c1[:])
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_COEFF_2, c2[:])
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_COEFF_3, c3[:])

	// Temperature coefficients (coeff1 block, from 0x8A)
	d.calib.t2 = int16(uint16(c1[1])<<8 | uint16(c1[0]))
	d.calib.t3 = int8(c1[2])

	// Pressure coefficients
	d.calib.p1 = uint16(c1[5])<<8 | uint16(c1[4])
	d.calib.p2 = int16(uint16(c1[7])<<8 | uint16(c1[6]))
	d.calib.p3 = int8(c1[8])
	d.calib.p4 = int16(uint16(c1[11])<<8 | uint16(c1[10]))
	d.calib.p5 = int16(uint16(c1[13])<<8 | uint16(c1[12]))
	d.calib.p7 = int8(c1[14])
	d.calib.p6 = int8(c1[15])
	d.calib.p8 = int16(uint16(c1[19])<<8 | uint16(c1[18]))
	d.calib.p9 = int16(uint16(c1[21])<<8 | uint16(c1[20]))
	d.calib.p10 = c1[22]

	// Humidity coefficients (coeff2 block, from 0xE1)
	// 0xE2 byte is shared: [7:4]=H2_lsb, [3:0]=H1_lsb
	d.calib.h2 = uint16(c2[0])<<4 | uint16(c2[1]>>4)
	d.calib.h1 = uint16(c2[2])<<4 | uint16(c2[1]&0x0F)
	d.calib.h3 = int8(c2[3])
	d.calib.h4 = int8(c2[4])
	d.calib.h5 = int8(c2[5])
	d.calib.h6 = c2[6]
	d.calib.h7 = int8(c2[7])

	// Temperature coefficient t1 (also in coeff2 block)
	d.calib.t1 = uint16(c2[9])<<8 | uint16(c2[8])

	// Gas heater coefficients
	d.calib.gh2 = int16(uint16(c2[11])<<8 | uint16(c2[10]))
	d.calib.gh1 = int8(c2[12])
	d.calib.gh3 = int8(c2[13])

	// Gas calibration extras (coeff3 block, from 0x00)
	d.calib.resHeatVal = int8(c3[0])
	d.calib.resHeatRange = (c3[2] & 0x30) >> 4
	d.calib.rangeSWErr = int8((c3[4] & 0xF0) >> 4)
}

// parseField converts a raw 17-byte field into a Measurement.
func (d *Device) parseField(f [LEN_FIELD]byte) Measurement {
	adcP := uint32(f[2])<<12 | uint32(f[3])<<4 | uint32(f[4])>>4
	adcT := uint32(f[5])<<12 | uint32(f[6])<<4 | uint32(f[7])>>4
	adcH := uint16(f[8])<<8 | uint16(f[9])
	adcG := uint16(f[13])<<2 | uint16(f[14]>>6)
	gasRange := f[14] & GAS_RANGE_MSK

	temp, tFine := d.calcTemperature(adcT)
	d.calib.tFine = tFine

	return Measurement{
		Temperature:   temp,
		Pressure:      d.calcPressure(adcP),
		Humidity:      d.calcHumidity(adcH),
		GasResistance: d.calcGasResistance(adcG, gasRange),
		GasValid:      f[14]&GAS_VALID_MSK != 0,
		HeaterStable:  f[14]&HEAT_STAB_MSK != 0,
	}
}

// calcTemperature returns temperature in milli-°C and the tFine compensation value.
// Formula from Bosch BME68x API (integer variant).
func (d *Device) calcTemperature(adcT uint32) (int32, int32) {
	var1 := int32(adcT>>3) - int32(d.calib.t1)<<1
	var2 := (var1 * int32(d.calib.t2)) >> 11
	var3 := ((var1 >> 1) * (var1 >> 1)) >> 12
	var3 = (var3 * (int32(d.calib.t3) << 4)) >> 14
	tFine := var2 + var3
	// (tFine*5+128)>>8 gives temperature in 0.01 °C; ×10 → milli-°C
	return 10 * ((tFine*5 + 128) >> 8), tFine
}

// calcPressure returns pressure in Pascals.
// Formula from Bosch BME68x API (integer variant).
func (d *Device) calcPressure(adcP uint32) uint32 {
	var1 := d.calib.tFine/2 - 64000
	var2 := ((var1/4)*(var1/4))>>11*int32(d.calib.p6)/2 +
		var1*int32(d.calib.p5)*2
	var2 = var2/4 + int32(d.calib.p4)*65536
	var1 = (((var1/4)*(var1/4))>>13*(int32(d.calib.p3)<<5)>>3 +
		int32(d.calib.p2)*var1/2) >> 18
	var1 = ((32768 + var1) * int32(d.calib.p1)) >> 15
	if var1 == 0 {
		return 0
	}
	comp := uint32(1048576 - int32(adcP))
	comp = uint32((int32(comp) - var2/4096) * 3125)
	if comp >= 0x80000000 {
		comp = (comp / uint32(var1)) * 2
	} else {
		comp = comp * 2 / uint32(var1)
	}
	var1 = int32(d.calib.p9) * int32(((comp>>3)*(comp>>3))>>13) >> 12
	var2 = int32(comp>>2) * int32(d.calib.p8) >> 13
	var3 := int32(comp>>8) * int32(comp>>8) * int32(comp>>8) * int32(d.calib.p10) >> 17
	comp = uint32(int32(comp) + (var1+var2+var3+int32(d.calib.p7)*128)/16)
	return comp
}

// calcHumidity returns relative humidity × 1000.
// Formula from Bosch BME68x API (integer variant).
func (d *Device) calcHumidity(adcH uint16) uint32 {
	tempScaled := (d.calib.tFine*5 + 128) >> 8
	var1 := int32(adcH) -
		int32(d.calib.h1)*16 -
		(tempScaled*int32(d.calib.h3)/100)>>1
	var2 := int32(d.calib.h2) *
		(tempScaled*int32(d.calib.h4)/100 +
			(((tempScaled * ((tempScaled * int32(d.calib.h5)) / 100)) >> 6) / 100) +
			(1 << 14)) >> 10
	var3 := var1 * var2
	var4 := int32(d.calib.h6) << 7
	var4 = (var4 + (tempScaled * int32(d.calib.h7) / 100)) >> 4
	var5 := ((var3 >> 14) * (var3 >> 14)) >> 10
	var6 := (var4 * var5) >> 1
	compHum := ((var3 + var6) >> 10) * 1000 >> 12
	if compHum < 0 {
		compHum = 0
	}
	if compHum > 100000 {
		compHum = 100000
	}
	return uint32(compHum)
}

// calcGasResistance returns gas resistance in Ohms.
// Formula from Bosch BME68x API using integer lookup tables.
func (d *Device) calcGasResistance(adcG uint16, gasRange byte) uint32 {
	var1 := int64(1340+5*int64(d.calib.rangeSWErr)) *
		int64(gasLookupTable1[gasRange]) >> 16
	var2 := int64(int64(adcG)<<15-16777216) + var1
	var3 := int64(gasLookupTable2[gasRange]) * var1 >> 9
	return uint32((var3 + var2>>1) / var2)
}

// calcResHeat converts a target heater temperature to the res_heat register value.
// Formula from Bosch BME68x API (float variant – called only at configure time).
func (d *Device) calcResHeat(targetTemp uint16) byte {
	if targetTemp > 400 {
		targetTemp = 400
	}
	ambT := float32(d.Config.AmbientTemp)
	tgt := float32(targetTemp)

	v1 := float32(d.calib.gh1)/16.0 + 49.0
	v2 := float32(d.calib.gh2)/32768.0*0.0005 + 0.00235
	v3 := float32(d.calib.gh3) / 1024.0
	v4 := v1 * (1.0 + v2*tgt)
	v5 := v4 + v3*ambT
	res := 3.4 * ((v5 * (4.0 / (4.0 + float32(d.calib.resHeatRange))) *
		(1.0 / (1.0 + float32(d.calib.resHeatVal)*0.002))) - 25.0)
	return byte(res)
}

// calcGasWait converts a duration in milliseconds to the gas_wait register value.
// Format: [7:6] = multiplier exponent (×4^n), [5:0] = base value.
func calcGasWait(dur uint16) byte {
	if dur >= 0xFC0 {
		return 0xFF
	}
	var factor byte
	for dur > 0x3F {
		dur >>= 2
		factor++
	}
	return byte(dur) + factor*64
}

// measurementDelay calculates how long to sleep after triggering forced mode.
func (d *Device) measurementDelay() time.Duration {
	const (
		overhead  = 1250 // µs
		perSample = 2300 // µs per oversampling step
	)
	sampleCounts := [6]int{0, 1, 2, 4, 8, 16}

	clamp := func(o Oversampling) int {
		if int(o) < len(sampleCounts) {
			return sampleCounts[o]
		}
		return 16
	}

	µs := overhead +
		clamp(d.Config.Temperature)*perSample +
		clamp(d.Config.Pressure)*perSample +
		clamp(d.Config.Humidity)*perSample

	dur := time.Duration(µs) * time.Microsecond
	if d.Config.GasEnabled {
		dur += time.Duration(d.Config.HeaterDuration)*time.Millisecond + 3*time.Millisecond
	}
	return dur
}
