package tmc5160

import (
	"github.com/orsinium-labs/tinymath"
	"golang.org/x/exp/constraints"
)

// VelocityToVMAX calculates the VMAX register value from the current stepper velocity which is  in microsteps per tRef (i.e 1/clock speed)
func (stepper *Stepper) CurrentVelocityToVMAX() uint32 {
	tref := float32(16777216) / (float32(stepper.Fclk) * 1000000)
	r := stepper.VelocitySPS * stepper.GearRatio * float32(tref)
	return constrain(uint32(r), 0, maxVMAX) // VMAX register value cannot exceed maxVMAX
}
func (stepper *Stepper) DesiredVelocityToVMAX(v float32) uint32 {
	tref := 16777216 / (float32(stepper.Fclk) * 1000000)
	r := tinymath.Round(v * stepper.GearRatio * tref)
	return constrain(uint32(r), 0, maxVMAX) // VMAX register value cannot exceed maxVMAX
}

func (stepper *Stepper) DesiredAccelToAMAX(dacc float32, dVel float32) uint32 {
	dVelToVMAX := stepper.DesiredVelocityToVMAX(dVel)
	_a := uint64(dVelToVMAX) * 131072
	_b := float32(_a) / dacc
	_c := _b / float32(uint32(stepper.Fclk)*1000000)
	return uint32(_c)

}

// Convert threshold speed (Hz) to internal TSTEP value
func (stepper *Stepper) DesiredSpeedToTSTEP(thrsSpeed uint32) uint32 {
	if thrsSpeed < 0 {
		return 0
	}
	_a := stepper.DesiredVelocityToVMAX(float32(thrsSpeed))
	_b := float32(16777216 / _a)
	_c := float32(stepper.MSteps) / float32(256)
	_d := uint32(_b * _c)
	return constrain(_d, 0, 1048575)
}

func (stepper *Stepper) VMAXToTSTEP(vmax uint32) uint32 {
	_b := float32(16777216 / vmax)
	_c := float32(stepper.MSteps) / float32(256)
	_d := tinymath.Round(_b * _c)
	return constrain(uint32(_d), 0, 1048575)
}

// Constrain function to limit values to a specific range (supports multiple types).
func constrain[T constraints.Ordered](value, min, max T) T {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}
