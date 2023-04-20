package tds

import (
	"errors"
	"machine"
	"time"
)

type device struct {
	adc            machine.ADC
	averageVoltage float32
	aref           float32 // ADC Reference voltage in volts
	resolution     float32
}

type Device interface {
	// GetTds collects readings and returns total dissolved solids in ppm.
	// Temperature must be in celsius, used to for temperature compensation.
	//
	//
	// Note: 1.8 is the closest for this manufacture. If no temperature can be provided, passing a constant 20.0 °C will result in 1.8 being calculated.
	GetTds(temp float32) (float32, error)
	// GetElectricalConductance collects readings and returns the electrical conductance compensated for temperature
	GetElectricalConductance(temp float32) float32
	Configure()
}

const (
	readCycle              time.Duration = time.Millisecond * 40
	sampleCount                          = 30
	referenceTemperature   float32       = 25.0
	temperatureCoefficient               = 0.02
	tdsFactor              float32       = 0.5 // electrical conductivity / 2
)

// New returns a new total dissolve solids sensor driver given an ADC pin.
func New(p machine.Pin, aref, resolution float32) Device {
	return &device{
		adc:        machine.ADC{Pin: p},
		aref:       aref,
		resolution: resolution,
	}
}

func (d *device) Configure() {
	d.adc.Configure(machine.ADCConfig{})
}

func (d *device) GetElectricalConductance(temp float32) float32 {
	m := d.findMedian(d.collectSamples())
	return d.calcElectricalConductance(d.adc2voltage(m), temp)
}

func (d *device) GetTds(temp float32) (float32, error) {
	m := d.findMedian(d.collectSamples())
	ec := d.calcElectricalConductance(d.adc2voltage(m), temp)
	return d.calcTds(ec)
}

func (d *device) collectSamples() []uint16 {
	readBuffer := make([]uint16, sampleCount)
	for i := 0; i < len(readBuffer); i++ {
		readBuffer[i] = d.adc.Get()
		time.Sleep(readCycle)
	}
	return readBuffer
}

func (d *device) findMedian(nums []uint16) uint16 {
	d.bubbleSort(nums)

	var median uint16
	l := len(nums)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (nums[l/2-1] + nums[l/2]) / 2
	} else {
		median = nums[l/2]
	}
	return median
}

func (d *device) adc2voltage(n uint16) float32 {
	return (float32(n) * d.aref) / d.resolution
}

// calcElectricalConductance returns the electrical conductance compensated for temperature
func (d *device) calcElectricalConductance(v float32, temp float32) float32 {
	// calculate temperature compensated voltage
	compV := v / d.temperatureCompensation(temp)

	// formulas were converted from source wiki Arduino example: http://www.cqrobot.wiki/index.php/TDS_(Total_Dissolved_Solids)_Meter_Sensor_SKU:_CQRSENTDS01#Arduino_Application
	// The TDS value is half of the electrical conductivity value: electrical conductivity / 2
	return 133.42*compV*compV*compV - 255.86*compV*compV + 857.39*compV
}

func (d *device) calcTds(ec float32) (float32, error) {
	tds := ec * tdsFactor
	if tds < 0.0 || tds > 1000.0 {
		return tds, errors.New("total dissolved solids reading is invalid:")
	}
	return tds, nil
}

// temperatureCompensation uses linear approximation to calculate the temperature coefficient of resistivity
// See for temperature dependance information: https://en.wikipedia.org/wiki/Electrical_resistivity_and_conductivity#Temperature_dependence
func (d *device) temperatureCompensation(temp float32) float32 {
	return 1.0 + temperatureCoefficient*(temp-referenceTemperature)
}

func (d *device) bubbleSort(nums []uint16) {
	for i := 0; i < len(nums)-1; i++ {
		for j := 0; j < len(nums)-i-1; j++ {
			if nums[j] > nums[j+1] {
				nums[j+1], nums[j] = nums[j], nums[j+1]
				break
			}
		}
	}
}
