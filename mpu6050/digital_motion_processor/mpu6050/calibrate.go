package mpu6050

import (
	"math"
	"strconv"
	"time"
)

/* Arduino version
func mapValue(x, in_min, in_max, out_min, out_max int64) int64 {
	val := (x-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
	println("mapValue: ", val)
	return val
}
*/

// My verison
func mapValue(value, fromLow, fromHi, toLow, toHi int64) int64 {
	/*
		if value < fromLow {
			value = fromLow
		}
		if value > toHi {
			value = toHi
		}
	*/
	inRatio := (value - fromLow) / (fromHi - fromLow)
	outValue := toLow + inRatio*(toHi-toLow)
	return outValue
}

func (d Device) CalibrateGyro(Loops uint8) {
	kP := float64(0.3)
	kI := float64(90)
	var x float64
	x = (100 - float64(mapValue(int64(Loops), 1, 5, 20, 0))) * .01
	kP *= x
	kI *= x

	d.PID(0x43, kP, kI, Loops)
}

func (d Device) CalibrateAccel(Loops uint8) {

	kP := float64(0.3)
	kI := float64(20)
	var x float64
	x = (100 - float64(mapValue(int64(Loops), 1, 5, 20, 0))) * .01
	kP *= x
	kI *= x
	d.PID(0x3B, kP, kI, Loops)
}

func (d Device) PID(readAddress uint8, kP, kI float64, loops uint8) {
	var saveAddress uint8
	gravity := uint16(8192)
	shift := uint8(2)
	id, _ := d.getDeviceID()
	if readAddress == 0x3B {
		ar, _ := d.getFullScaleAccelRange()
		gravity = 16384 >> ar
		if id < 0x38 {
			saveAddress = 0x06
		} else {
			saveAddress = 0x77
			shift = uint8(3)
		}
	} else {
		saveAddress = 0x13
	}

	debug_print("id: 0x" + strconv.FormatUint(uint64(id), 16))
	debug_print("readAddress: 0x" + strconv.FormatUint(uint64(readAddress), 16))
	debug_print("saveAddress: 0x" + strconv.FormatUint(uint64(saveAddress), 16))
	debug_print("shift: 0x" + strconv.FormatUint(uint64(shift), 16))

	data := make([]int16, 1)
	var reading float64
	bitZero := make([]int16, 3)
	var pid_error, pterm float64
	iterm := make([]float64, 3)
	var esample int16
	var esum uint32

	print(">")
	for i := uint8(0); i < 3; i++ {
		err := d.ReadWords(saveAddress+i*shift, data)
		if err != nil {
			println("Unable to read words. Aborting")
			for {
				//failure
			}
		}
		reading = float64(data[0])
		if saveAddress != 0x13 {
			bitZero[i] = data[0] & int16(1)
			iterm[i] = reading * 8
		} else {
			iterm[i] = reading * 4
		}
	}

	for l := 0; l < int(loops); l++ {
		esample = 0
		for c := 0; c < 100; c++ {
			esum = 0
			for i := uint8(0); i < 3; i++ {
				err := d.ReadWords(readAddress+i*2, data)
				if err != nil {
					println("Unable to read words. Aborting")
					for {
						//failure
					}
				}
				reading = float64(data[0])
				if readAddress == 0x3B && i == 2 {
					reading -= float64(gravity)
				}
				pid_error = -reading
				esum += uint32(math.Abs(reading))
				pterm = kP * pid_error
				iterm[i] += pid_error * 0.001 * kI // Integral term 1000 calculations
				if saveAddress != 0x13 {
					data[0] = int16(math.Round((pterm + iterm[i]) / 8))  // compute pid output
					data[0] = int16(uint16(data[0])&0xFFFE) | bitZero[i] // insert Bit0 saved at beginning
				} else {
					data[0] = int16(math.Round((pterm + iterm[i]) / 4)) // compute PID output
				}
				err = d.WriteWords(saveAddress+i*shift, data)
				if err != nil {
					println("Unable to write words. Aborting")
					for {
						//failure
					}
				}
			}
			if (c == 99) && (esum > 1000) {
				c = 0
				print("*")
			}
			var factor float64
			if readAddress == 0x3B {
				factor = 0.05
			} else {
				factor = 1.0
			}
			if (float64(esum) * factor) < 5 {
				esample++
			}
			if (esum < 100) && (c > 10) && (esample >= 10) {
				break
			}
			time.Sleep(time.Millisecond)
		}
		print(".")
		kP *= 0.75
		kI *= 0.75
		for i := uint8(0); i < 3; i++ {
			if saveAddress != 0x13 {
				data[0] = int16(math.Round(iterm[i] / 8))
				data[0] = int16(uint16(data[0])&0xFFFE) | bitZero[i]
			} else {
				data[0] = int16(math.Round(iterm[i] / 4))
			}
			err := d.WriteWords(saveAddress+i*shift, data)
			if err != nil {
				println("Error writing words to MPU6050. Aborting ...")
				for {
					//failure
				}
			}
		}

	}
	d.resetFIFO()
	d.resetDMP()
}

func (d Device) getFullScaleAccelRange() (uint8, error) {
	return d.ReadBits(ACCEL_CONFIG, ACCEL_CONFIG_AFS_SEL_BIT, ACCEL_CONFIG_AFS_SEL_LENGTH)
}

func (d Device) getDeviceID() (uint8, error) {
	return d.ReadBits(WHO_AM_I, WHO_AM_I_BIT, WHO_AM_I_LENGTH)
}
