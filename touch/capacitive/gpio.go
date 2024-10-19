package capacitive

import (
	"machine"
	"runtime/interrupt"
	"time"
)

const (
	// How often to measure.
	// The Update function will wait until this amount of time has passed.
	measurementFrequency       = 200
	minTimeBetweenMeasurements = time.Second / measurementFrequency

	// How much to multiply values before averaging. A value higher than 1 will
	// help to avoid integer rounding errors and may improve accuracy slightly.
	oversampling = 8

	// How many samples to use for the moving average.
	movingAverageWindow = 16

	// After how many samples should the touch sensor be recalibrated?
	// This should be a power of two (for efficient division) and be a multiple
	// of movingAverageWindow. Ideally it should cause a recalibration every 5s
	// or so.
	recalibrationSamples = 1024
)

type Array struct {
	// Time when the last update finished. This is used to make sure we call
	// Update() the expected number of times per second.
	lastUpdate time.Time

	// List of pins to measure each time.
	pins []machine.Pin

	// Raw values (non-smoothed) from the last read.
	values []uint16

	hasFirstMeasurement bool

	// Static threshold. Zero if using a dynamic threshold.
	staticThreshold uint16

	// How long to measure.
	measureCycles uint16

	// Sensitivity (in promille) for the dynamic threshold.
	sensitivity uint16

	// Capacitance trackers for dynamic capacitance measurement.
	trackers []capacitanceTracker
}

// Create a new array of pins to be used as touch sensors.
// The pins do not need to be initialized. The array is immediately ready to
// use.
//
// By default, NewArray configures a static threshold that is not very
// sensitive. If you want the touch inputs to be more sensitive, use
// SetDynamicThreshold.
func NewArray(pins []machine.Pin) *Array {
	for _, pin := range pins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.High()
	}
	array := &Array{
		pins:          pins,
		values:        make([]uint16, len(pins)),
		measureCycles: uint16(machine.CPUFrequency() / 125000), // 1000 on the RP2040 (which is 125MHz)
		lastUpdate:    time.Now(),
	}

	// A threshold of 500 works well on the RP2040. Scale this number to
	// something similar on other chips.
	array.SetStaticThreshold(int(machine.CPUFrequency() / 250000))

	return array
}

// Use a static threshold. This works well on simple touch surfaces where you'll
// directly touch the metal.
func (a *Array) SetStaticThreshold(threshold int) {
	if threshold > 0xffff {
		threshold = 0xffff
	}
	a.staticThreshold = uint16(threshold)
	a.trackers = nil
}

// Use a dynamic threshold (as promille), that will calibrate automatically.
// This is needed when you want to be able to detect touches through a
// non-conducting surface for example. Something like 100‰ (10%) will probably
// work in many cases, though you may need to try different value to reliably
// detect touches.
func (a *Array) SetDynamicThreshold(sensitivity int) {
	a.sensitivity = uint16(sensitivity)
	a.staticThreshold = 0
	a.trackers = make([]capacitanceTracker, len(a.pins))
}

// Measure all GPIO pins. This function must be called very often, ideally about
// 100-200 times per second (it will delay a bit when called more than 200 times
// per second).
func (a *Array) Update() {
	// Wait until enough time has passed to charge all pins.
	now := time.Now()
	timeSinceLastUpdate := now.Sub(a.lastUpdate)
	sleepTime := minTimeBetweenMeasurements - timeSinceLastUpdate
	time.Sleep(sleepTime)
	a.lastUpdate = now.Add(sleepTime) // should be ~equivalent to time.Now()

	// Measure each pin in turn.
	for i, pin := range a.pins {
		// Interrupts must be disabled during measuring for accurate results.
		mask := interrupt.Disable()

		// Switch to input. This will stop the charging, and let it discharge
		// through the resistor.
		pin.Configure(machine.PinConfig{Mode: machine.PinInput})

		// Wait for the pin to go low again.
		// A longer duration means more capacitance, which means something is
		// touching it (finger, banana, etc).
		count := uint32(i)
		for i := 0; i < int(a.measureCycles); i++ {
			if !pin.Get() {
				break
			}
			count++
		}

		interrupt.Restore(mask)

		a.values[i] = uint16(count)

		// Set the pin to high, to charge it for the next measurement.
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.High()
	}

	// The first measurement tends to be slightly off (too low value) so ignore
	// that one.
	if !a.hasFirstMeasurement {
		a.hasFirstMeasurement = true
		return
	}

	for i := 0; i < len(a.trackers); i++ {
		a.trackers[i].addValue(int(a.values[i]), int(a.sensitivity))
	}
}

// Return the raw value of the given pin index of the most recent call to
// Update. This value is not smoothed in any way.
func (a *Array) RawValue(index int) int {
	return int(a.values[index])
}

// Return the value from the moving average. This value is only available when a
// dynamic threshold has been set, it will panic otherwise.
func (a *Array) SmoothedValue(index int) int {
	return int(a.trackers[index].avg) / oversampling
}

// Return whether the given pin index is currently being touched.
func (a *Array) Touching(index int) bool {
	if a.staticThreshold != 0 {
		// Using a static threshold.
		return a.values[index] > a.staticThreshold
	}

	return a.trackers[index].touching
}

// Separate object to store calibration data and track capacitance over time.
type capacitanceTracker struct {
	recentValues [movingAverageWindow]uint16
	sum          uint32
	avg          uint16

	baseline   uint16
	noise      uint16
	valueCount uint8
	touching   bool

	recalibrationCount    uint8
	recalibrationPrevAvg  uint16
	recalibrationNoiseSum int32
	recalibrationSum      uint32
}

func (ct *capacitanceTracker) addValue(value int, sensitivity int) {
	// Maybe increase the resolution slightly by oversampling. This should
	// increase the resolution a little bit after averaging and should reduce
	// rounding errors.
	// Typical input values on the RP2040 are 100-200 (or up to 1000 or so when
	// touching the metal) so multiplying by 4-8 should be fine. Other chips
	// generally have much lower values.
	value *= oversampling
	if value > 0xffff {
		value = 0xffff // unlikely, but make sure we don't overflow
	}

	// This does a number of things at the same time:
	//  * Add the new value to the recentValues array.
	//  * Calculate the moving sum (and average) of recentValues using a
	//    recursive moving average algorithm:
	//    https://www.dspguide.com/ch15/5.htm
	ptr := &ct.recentValues[ct.valueCount%movingAverageWindow]
	ct.sum -= uint32(*ptr)
	ct.sum += uint32(value)
	ct.avg = uint16(ct.sum / movingAverageWindow)
	*ptr = uint16(value)
	ct.valueCount++

	// Do an initial calibration once the first values have been read.
	if ct.baseline == 0 && ct.valueCount == movingAverageWindow {
		ct.baseline = ct.avg

		// Calculate initial noise as an average absolute deviation:
		// https://en.wikipedia.org/wiki/Average_absolute_deviation
		// This is a quick and imprecise way to find the noise, better noise
		// detection happens during recalibration.
		var diffSum uint32
		for _, sample := range ct.recentValues {
			diff := int(ct.avg) - int(sample)
			if diff < 0 {
				diff = -diff
			}
			diffSum += uint32(diff)
		}
		ct.noise = uint16(diffSum / (movingAverageWindow / 2))
	}

	// Now determine whether the touch pad is being touched.

	if ct.baseline == 0 {
		// Not yet calibrated.
		ct.touching = false
		return
	}

	// Calculate the threshold.
	// Divide by 65536 (instead of 65500) to avoid a potentially expensive
	// division while still being close enough.
	threshold := (uint32(ct.baseline) * uint32(sensitivity+1000) * 65) / 65536

	// Add noise to the threshold, to avoid toggling quickly. This mainly
	// filters out mains noise.
	threshold += uint32(ct.noise)

	// Implement some hysteresis: if the touch pad was previously touched, lower
	// the threshold a little to avoid bouncing effects.
	// TODO: let this hysteresis depend on the amount of noise.
	if ct.touching {
		threshold = (threshold*3 + uint32(ct.baseline)) / 4 // lower the threshold by 25%
	}

	// Is the pad being touched?
	ct.touching = uint32(ct.avg) > threshold

	// Do a recalibration after the sensor hasn't been touched for ~5s, to
	// account for drift over time (humidity etc).
	if ct.touching {
		// Reset calibration (start from zero).
		ct.recalibrationCount = 0
		ct.recalibrationSum = 0
		ct.recalibrationNoiseSum = 0
	} else {
		// Add the last batch of samples to the sum.
		if ct.valueCount%movingAverageWindow == 0 {
			ct.recalibrationCount++

			// Wait a few cycles before starting data collection for
			// calibration.
			cycle := int(ct.recalibrationCount) - 3

			if cycle < 0 {
				// Store the previous average, to calculate the noise value.
				ct.recalibrationPrevAvg = ct.avg

			} else if cycle >= 0 {
				// Collect data for recalibration.
				ct.recalibrationSum += ct.sum

				// Add difference between two (averaged) samples as a measure of
				// the noise.
				diff := int32(ct.recalibrationPrevAvg) - int32(ct.avg)
				if diff < 0 {
					diff = -diff
				}
				ct.recalibrationNoiseSum += diff
				ct.recalibrationPrevAvg = ct.avg

			}

			// Do the recalibration after enough samples have been collected.
			// Note: the noise is basically the average of absolute differences
			// between two averaging windows. I don't know whether this
			// algorithm has a name, but it seems to work here to detect the
			// amount of noise.
			const totalRecalibrationCount = recalibrationSamples / movingAverageWindow
			if cycle == totalRecalibrationCount {
				ct.baseline = uint16(ct.recalibrationSum / recalibrationSamples)
				ct.noise = uint16(ct.recalibrationNoiseSum / (totalRecalibrationCount / 2))
				ct.recalibrationCount = 0
				ct.recalibrationSum = 0
				ct.recalibrationNoiseSum = 0
			}
		}
	}
}
