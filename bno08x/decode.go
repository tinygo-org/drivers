package bno08x

import "encoding/binary"

// decodeSensor decodes a sensor report payload into a SensorValue.
func decodeSensor(payload []byte, timestamp uint32) (SensorValue, bool) {
	if len(payload) < 4 {
		return SensorValue{}, false
	}

	value := SensorValue{
		id:        SensorID(payload[0]),
		sequence:  payload[1],
		status:    payload[2] & 0x03,
		delay:     payload[3],
		timestamp: uint64(timestamp),
	}

	data := payload[4:]

	switch value.id {
	case SensorRawAccelerometer:
		if len(data) >= 10 {
			value.rawAccelerometer = RawVector3{
				X:         int16(binary.LittleEndian.Uint16(data[0:])),
				Y:         int16(binary.LittleEndian.Uint16(data[2:])),
				Z:         int16(binary.LittleEndian.Uint16(data[4:])),
				Timestamp: binary.LittleEndian.Uint32(data[6:]),
			}
		}

	case SensorAccelerometer:
		if len(data) >= 6 {
			value.accelerometer = Vector3{
				X: qToFloat(data[0:], scaleAccel),
				Y: qToFloat(data[2:], scaleAccel),
				Z: qToFloat(data[4:], scaleAccel),
			}
		}

	case SensorLinearAcceleration:
		if len(data) >= 6 {
			value.linearAcceleration = Vector3{
				X: qToFloat(data[0:], scaleAccel),
				Y: qToFloat(data[2:], scaleAccel),
				Z: qToFloat(data[4:], scaleAccel),
			}
		}

	case SensorGravity:
		if len(data) >= 6 {
			value.gravity = Vector3{
				X: qToFloat(data[0:], scaleAccel),
				Y: qToFloat(data[2:], scaleAccel),
				Z: qToFloat(data[4:], scaleAccel),
			}
		}

	case SensorRawGyroscope:
		if len(data) >= 12 {
			value.rawGyroscope = RawGyroscope{
				X:           int16(binary.LittleEndian.Uint16(data[0:])),
				Y:           int16(binary.LittleEndian.Uint16(data[2:])),
				Z:           int16(binary.LittleEndian.Uint16(data[4:])),
				Temperature: int16(binary.LittleEndian.Uint16(data[6:])),
				Timestamp:   binary.LittleEndian.Uint32(data[8:]),
			}
		}

	case SensorGyroscope:
		if len(data) >= 6 {
			value.gyroscope = Vector3{
				X: qToFloat(data[0:], scaleGyro),
				Y: qToFloat(data[2:], scaleGyro),
				Z: qToFloat(data[4:], scaleGyro),
			}
		}

	case SensorGyroscopeUncalibrated:
		if len(data) >= 12 {
			value.gyroscopeUncal = GyroscopeUncalibrated{
				X:     qToFloat(data[0:], scaleGyro),
				Y:     qToFloat(data[2:], scaleGyro),
				Z:     qToFloat(data[4:], scaleGyro),
				BiasX: qToFloat(data[6:], scaleGyro),
				BiasY: qToFloat(data[8:], scaleGyro),
				BiasZ: qToFloat(data[10:], scaleGyro),
			}
		}

	case SensorRawMagnetometer:
		if len(data) >= 10 {
			value.rawMagnetometer = RawVector3{
				X:         int16(binary.LittleEndian.Uint16(data[0:])),
				Y:         int16(binary.LittleEndian.Uint16(data[2:])),
				Z:         int16(binary.LittleEndian.Uint16(data[4:])),
				Timestamp: binary.LittleEndian.Uint32(data[6:]),
			}
		}

	case SensorMagneticField:
		if len(data) >= 6 {
			value.magneticField = Vector3{
				X: qToFloat(data[0:], scaleMag),
				Y: qToFloat(data[2:], scaleMag),
				Z: qToFloat(data[4:], scaleMag),
			}
		}

	case SensorMagneticFieldUncalibrated:
		if len(data) >= 12 {
			value.magneticFieldUncal = MagneticFieldUncalibrated{
				X:     qToFloat(data[0:], scaleMag),
				Y:     qToFloat(data[2:], scaleMag),
				Z:     qToFloat(data[4:], scaleMag),
				BiasX: qToFloat(data[6:], scaleMag),
				BiasY: qToFloat(data[8:], scaleMag),
				BiasZ: qToFloat(data[10:], scaleMag),
			}
		}

	case SensorRotationVector:
		if len(data) >= 10 {
			value.quaternion = Quaternion{
				I:    qToFloat(data[0:], scaleQuat),
				J:    qToFloat(data[2:], scaleQuat),
				K:    qToFloat(data[4:], scaleQuat),
				Real: qToFloat(data[6:], scaleQuat),
			}
			value.quaternionAccuracy = qToFloat(data[8:], scaleAccuracy)
		}

	case SensorGameRotationVector:
		if len(data) >= 8 {
			value.quaternion = Quaternion{
				I:    qToFloat(data[0:], scaleQuat),
				J:    qToFloat(data[2:], scaleQuat),
				K:    qToFloat(data[4:], scaleQuat),
				Real: qToFloat(data[6:], scaleQuat),
			}
		}

	case SensorGeomagneticRotationVector:
		if len(data) >= 10 {
			value.quaternion = Quaternion{
				I:    qToFloat(data[0:], scaleQuat),
				J:    qToFloat(data[2:], scaleQuat),
				K:    qToFloat(data[4:], scaleQuat),
				Real: qToFloat(data[6:], scaleQuat),
			}
			value.quaternionAccuracy = qToFloat(data[8:], scaleAccuracy)
		}

	case SensorARVRStabilizedRV:
		if len(data) >= 10 {
			value.quaternion = Quaternion{
				I:    qToFloat(data[0:], scaleQuat),
				J:    qToFloat(data[2:], scaleQuat),
				K:    qToFloat(data[4:], scaleQuat),
				Real: qToFloat(data[6:], scaleQuat),
			}
			value.quaternionAccuracy = qToFloat(data[8:], scaleAccuracy)
		}

	case SensorARVRStabilizedGRV:
		if len(data) >= 8 {
			value.quaternion = Quaternion{
				I:    qToFloat(data[0:], scaleQuat),
				J:    qToFloat(data[2:], scaleQuat),
				K:    qToFloat(data[4:], scaleQuat),
				Real: qToFloat(data[6:], scaleQuat),
			}
		}

	case SensorGyroIntegratedRV:
		if len(data) >= 10 {
			value.quaternion = Quaternion{
				I:    qToFloat(data[0:], scaleQuat),
				J:    qToFloat(data[2:], scaleQuat),
				K:    qToFloat(data[4:], scaleQuat),
				Real: qToFloat(data[6:], scaleQuat),
			}
			// Angular velocity X at data[8:10]
		}

	case SensorPressure:
		if len(data) >= 4 {
			value.pressure = float32(int32(binary.LittleEndian.Uint32(data[0:]))) * scalePressure
		}

	case SensorAmbientLight:
		if len(data) >= 4 {
			value.ambientLight = float32(int32(binary.LittleEndian.Uint32(data[0:]))) * scaleLight
		}

	case SensorHumidity:
		if len(data) >= 2 {
			value.humidity = qToFloat(data[0:], scaleHumidity)
		}

	case SensorProximity:
		if len(data) >= 2 {
			value.proximity = qToFloat(data[0:], scaleProximity)
		}

	case SensorTemperature:
		if len(data) >= 2 {
			value.temperature = qToFloat(data[0:], scaleTemperature)
		}

	case SensorTapDetector:
		if len(data) >= 1 {
			value.tapDetector = TapDetector{
				Flags: data[0],
			}
		}

	case SensorStepDetector:
		if len(data) >= 4 {
			value.stepDetector = StepDetector{
				Latency: binary.LittleEndian.Uint32(data[0:]),
			}
		}

	case SensorStepCounter:
		if len(data) >= 8 {
			value.stepCounter = StepCounter{
				Count:   uint16(binary.LittleEndian.Uint32(data[4:8])),
				Latency: binary.LittleEndian.Uint32(data[0:4]),
			}
		}

	case SensorSignificantMotion:
		if len(data) >= 2 {
			value.significantMotion = SignificantMotion{
				Motion: binary.LittleEndian.Uint16(data[0:]),
			}
		}

	case SensorStabilityClassifier:
		if len(data) >= 1 {
			value.stabilityClassifier = StabilityClassifier{
				Classification: data[0],
			}
		}

	case SensorStabilityDetector:
		if len(data) >= 1 {
			value.stabilityDetector = data[0]
		}

	case SensorShakeDetector:
		if len(data) >= 2 {
			value.shakeDetector = ShakeDetector{
				Shake: binary.LittleEndian.Uint16(data[0:]),
			}
		}

	case SensorFlipDetector:
		if len(data) >= 2 {
			value.flipDetector = binary.LittleEndian.Uint16(data[0:2])
		}

	case SensorPickupDetector:
		if len(data) >= 2 {
			// Pickup detected at data[0:2]
		}

	case SensorPersonalActivityClassifier:
		if len(data) >= 16 {
			value.personalActivityClassifier = PersonalActivityClassifier{
				Page:            data[0],
				MostLikelyState: data[1],
				EndOfPage:       data[15],
			}
			for i := 0; i < 10 && i+2 < len(data); i++ {
				value.personalActivityClassifier.Confidence[i] = data[2+i]
			}
		}

	case SensorSleepDetector:
		if len(data) >= 1 {
			value.sleepDetector = data[0]
		}

	case SensorTiltDetector:
		if len(data) >= 1 {
			value.tiltDetector = data[0]
		}

	case SensorPocketDetector:
		if len(data) >= 1 {
			value.pocketDetector = data[0]
		}

	case SensorCircleDetector:
		if len(data) >= 1 {
			value.circleDetector = data[0]
		}

	case SensorHeartRateMonitor:
		if len(data) >= 2 {
			value.heartRateMonitor = binary.LittleEndian.Uint16(data[0:])
		}
	}

	return value, true
}

// qToFloat converts a Q-point fixed-point value to float32.
func qToFloat(data []byte, scale float32) float32 {
	if len(data) < 2 {
		return 0
	}
	return float32(int16(binary.LittleEndian.Uint16(data))) * scale
}
