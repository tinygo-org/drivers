package as560x // import tinygo.org/x/drivers/ams560x

import "math"

// convertFromNativeAngle converts and scales an angle from the device's native 12-bit range to the requested units
func convertFromNativeAngle(angle uint16, maxAngle uint16, units AngleUnit) (uint16, float32) {
	// MANG == 0 & MANG == NATIVE_ANGLE_RANGE (1 << 12) mean the same thing: use full circle range
	// but the latter makes the maths/code simpler
	if 0 == maxAngle {
		maxAngle = NATIVE_ANGLE_RANGE
	}
	switch units {
	case ANGLE_NATIVE:
		// For native angles, scaling has already been done by the device
		return angle, float32(angle)
	case ANGLE_DEGREES_INT:
		// Convert to degrees using integer arithmetic. Less accuracy but faster
		var deg int = 0
		if NATIVE_ANGLE_RANGE == maxAngle {
			// Simplify the conversion when using the full range
			deg = int(angle) * 360 >> 12
		} else {
			// Using an integer degrees scale with a narrower native range is pointless since we don't
			// benefit at all from the increase in native resolution, in fact we LOSE precision.
			// Alas, we have to return something
			// First get maxAngle on the degrees scale
			degMang, _ := convertFromNativeAngle(maxAngle, NATIVE_ANGLE_RANGE, units)
			// Now scale angle
			deg = int(angle) * int(degMang) / NATIVE_ANGLE_RANGE
		}
		return uint16(deg), float32(deg)
	case ANGLE_DEGREES_FLOAT:
		// Convert to degrees using floating point. More accuracy at expense of speed
		var degF float32 = 0.0
		if NATIVE_ANGLE_RANGE == maxAngle {
			// Simplify the conversion when using the full range
			degF = float32(angle) * 360.0 / NATIVE_ANGLE_RANGE
		} else {
			// Scale to degrees using a narrower native range
			// First get maxAngle on the degrees scale
			_, degMangF := convertFromNativeAngle(maxAngle, NATIVE_ANGLE_RANGE, units)
			// Now scale angle
			degF = float32(angle) * degMangF / NATIVE_ANGLE_RANGE
		}
		return uint16(degF), degF
	case ANGLE_RADIANS:
		// Convert to radians. Can only be done using floating point.
		var rad float32 = 0.0
		if NATIVE_ANGLE_RANGE == maxAngle {
			// Simplify the conversion when using the full range
			rad = float32(angle) * 2 * math.Pi / NATIVE_ANGLE_RANGE
		} else {
			// Scale to radians using a narrower native range
			// First get maxAngle on the radians scale
			_, radMang := convertFromNativeAngle(maxAngle, NATIVE_ANGLE_RANGE, units)
			// Now scale angle
			rad = float32(angle) * radMang / NATIVE_ANGLE_RANGE
		}
		return uint16(rad), rad
	default:
		panic("Unknown angle measurement unit")
	}
}

// convertToNativeAngle converts an angle from the requested units to the device's native 12-bit range.
func convertToNativeAngle(angle float32, units AngleUnit) uint16 {
	var pos uint16 = 0
	switch units {
	case ANGLE_NATIVE:
		pos = uint16(angle)
	case ANGLE_DEGREES_INT:
		fallthrough
	case ANGLE_DEGREES_FLOAT:
		// Convert from degrees
		angle = float32(math.Mod(float64(angle), 360.0))
		if angle < 0.0 {
			angle += 360.0
		}
		pos = uint16(math.Round(float64(angle) * NATIVE_ANGLE_RANGE / 360.0))
	case ANGLE_RADIANS:
		// Convert from radians
		const circRad = 2.0 * math.Pi
		angle = float32(math.Mod(float64(angle), circRad))
		if angle < 0.0 {
			angle += circRad
		}
		pos = uint16(math.Round(float64(angle) * NATIVE_ANGLE_RANGE / circRad))
	default:
		panic("Unknown angle measurement unit")
	}
	if pos > NATIVE_ANGLE_MAX {
		pos = NATIVE_ANGLE_MAX
	}
	return pos
}
