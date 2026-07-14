// SH-2 specification found at https://www.ceva-ip.com/wp-content/uploads/SH-2-Reference-Manual.pdf

package bno08x

import (
	"encoding/binary"
	"time"
)

// getReportLen returns the length in bytes of a sensor report given its ID.
// Returns 0 for unknown report IDs.
func getReportLen(reportID byte) int {
	switch reportID {
	case 0xF1: // FLUSH_COMPLETED
		return 6
	case 0xFA: // TIMESTAMP_REBASE
		return 5
	case 0xFB: // BASE_TIMESTAMP_REF
		return 5
	case 0xFC: // GET_FEATURE_RESP
		return 17
	case 0x01: // Accelerometer (calibrated)
		return 10
	case 0x02: // Gyroscope (calibrated)
		return 10
	case 0x03: // Magnetic field (calibrated)
		return 10
	case 0x04: // Linear acceleration
		return 10
	case 0x05: // Rotation vector
		return 14
	case 0x06: // Gravity
		return 10
	case 0x07: // Gyroscope uncalibrated
		return 16
	case 0x08: // Game rotation vector
		return 12
	case 0x09: // Geomagnetic rotation vector
		return 14
	case 0x0A: // Pressure
		return 10
	case 0x0B: // Ambient light
		return 10
	case 0x0C: // Humidity
		return 10
	case 0x0D: // Proximity
		return 10
	case 0x0E: // Temperature
		return 10
	case 0x0F: // Magnetic field uncalibrated
		return 16
	case 0x10: // Tap detector
		return 5
	case 0x11: // Step counter
		return 12
	case 0x12: // Significant motion
		return 6
	case 0x13: // Stability classifier
		return 5
	case 0x14: // Raw accelerometer
		return 16
	case 0x15: // Raw gyroscope
		return 16
	case 0x16: // Raw magnetometer
		return 16
	case 0x18: // Step detector
		return 8
	case 0x19: // Shake detector
		return 6
	case 0x1A: // Flip detector
		return 6
	case 0x1B: // Pickup detector
		return 6
	case 0x1C: // Stability detector
		return 6
	case 0x1E: // Personal activity classifier
		return 16
	default:
		// For most sensor reports, they are typically 10-16 bytes
		// If we don't know the exact length, return a safe default
		// that covers most cases (the handler will bounds-check)
		if reportID < 0xF0 {
			return 10 // Most sensor reports are at least this long
		}
		return 0
	}
}

// sh2Protocol implements the Sensor Hub 2 (SH-2) application protocol.
type sh2Protocol struct {
	device               *Device
	transport            *shtp
	cmdSeq               uint8
	waiting              bool
	lastCmd              uint8
	pendingConfigRequest bool
	pendingConfigSensor  SensorID
	receivedConfig       SensorConfig
	configReady          bool
	configBuf            [17]byte                    // Reusable buffer for setSensorConfig
	commandBuf           [3 + commandParamCount]byte // Reusable buffer for sendCommand
}

func newSH2Protocol(device *Device) *sh2Protocol {
	proto := &sh2Protocol{
		device:    device,
		transport: device.shtp,
	}

	// Register handlers for each channel
	device.shtp.register(channelControl, proto.handleControl)
	device.shtp.register(channelSensorReport, proto.handleSensor)
	device.shtp.register(channelWakeReport, proto.handleSensor)
	device.shtp.register(channelGyroRV, proto.handleSensor)
	device.shtp.register(channelExecutable, proto.handleExecutable)

	return proto
}

// softReset sends a software reset command to the sensor.
func (s *sh2Protocol) softReset() error {
	payload := []byte{execDeviceCmdReset}
	return s.transport.send(channelExecutable, payload)
}

// initialize sends the initialize command to the sensor.
func (s *sh2Protocol) initialize() error {
	return s.sendCommand(cmdInitialize, []byte{initSystem})
}

// requestProductIDs requests product identification information.
func (s *sh2Protocol) requestProductIDs() error {
	payload := []byte{reportProdIDReq, 0x00}
	return s.transport.send(channelControl, payload)
}

// enableReport enables a sensor report at the specified interval.
func (s *sh2Protocol) enableReport(id SensorID, intervalUs uint32) error {
	config := SensorConfig{
		ReportInterval: intervalUs,
	}
	return s.setSensorConfig(id, config)
}

// getSensorConfig retrieves the configuration for a sensor.
// This method sends a GET_FEATURE request and waits for the response
// by polling the device. It will timeout after approximately 1 second.
func (s *sh2Protocol) getSensorConfig(id SensorID) (SensorConfig, error) {
	// Mark that we're waiting for a config response
	s.pendingConfigRequest = true
	s.pendingConfigSensor = id
	s.configReady = false

	payload := []byte{reportGetFeature, byte(id)}
	err := s.transport.send(channelControl, payload)
	if err != nil {
		s.pendingConfigRequest = false
		return SensorConfig{}, err
	}

	// Poll for response with timeout
	maxAttempts := 100 // ~1 second with 10ms delays
	for i := 0; i < maxAttempts; i++ {
		// Service the device to process incoming messages
		s.device.shtp.poll()

		if s.configReady {
			s.pendingConfigRequest = false
			s.configReady = false
			return s.receivedConfig, nil
		}

		// Small delay between polls
		time.Sleep(10 * time.Millisecond)
	}

	s.pendingConfigRequest = false
	return SensorConfig{}, errTimeout
}

// setSensorConfig configures a sensor.
func (s *sh2Protocol) setSensorConfig(id SensorID, config SensorConfig) error {
	// Use pre-allocated buffer to avoid allocations
	payload := s.configBuf[:]
	payload[0] = reportSetFeature
	payload[1] = byte(id)

	// Build feature flags
	var flags uint8
	if config.ChangeSensitivityEnabled {
		flags |= featChangeSensitivityEnabled
	}
	if config.ChangeSensitivityRelative {
		flags |= featChangeSensitivityRelative
	}
	if config.WakeupEnabled {
		flags |= featWakeEnabled
	}
	if config.AlwaysOnEnabled {
		flags |= featAlwaysOnEnabled
	}
	payload[2] = flags

	binary.LittleEndian.PutUint16(payload[3:5], config.ChangeSensitivity)
	binary.LittleEndian.PutUint32(payload[5:9], config.ReportInterval)
	binary.LittleEndian.PutUint32(payload[9:13], config.BatchInterval)
	binary.LittleEndian.PutUint32(payload[13:17], config.SensorSpecific)

	return s.transport.send(channelControl, payload)
}

// sendCommand sends a command with parameters to the sensor.
func (s *sh2Protocol) sendCommand(command byte, params []byte) error {
	// Use pre-allocated buffer to avoid allocations
	payload := s.commandBuf[:]
	payload[0] = reportCommandReq
	payload[1] = s.cmdSeq
	payload[2] = command
	s.cmdSeq++
	s.lastCmd = command
	s.waiting = true

	for i := 0; i < commandParamCount && i < len(params); i++ {
		payload[3+i] = params[i]
	}

	return s.transport.send(channelControl, payload[:3+commandParamCount])
}

// handleControl processes control channel messages.
func (s *sh2Protocol) handleControl(payload []byte, timestamp uint32) {
	if len(payload) == 0 {
		return
	}

	reportID := payload[0]

	switch reportID {
	case reportProdIDResp:
		s.handleProdID(payload, timestamp)
	case reportCommandResp:
		s.handleCommandResp(payload, timestamp)
	case reportGetFeatureResp:
		s.handleGetFeatureResp(payload, timestamp)
	case reportFRSReadResp:
		// FRS (Flash Record System) read response
		// Not implemented in basic version
	}
}

// handleProdID processes product ID responses.
func (s *sh2Protocol) handleProdID(payload []byte, timestamp uint32) {
	if len(payload) < 16 {
		return
	}

	entry := ProductID{
		ResetCause:   payload[1],
		VersionMajor: payload[2],
		VersionMinor: payload[3],
		PartNumber:   binary.LittleEndian.Uint32(payload[4:8]),
		BuildNumber:  binary.LittleEndian.Uint32(payload[8:12]),
		VersionPatch: binary.LittleEndian.Uint16(payload[12:14]),
		Reserved0:    payload[14],
		Reserved1:    payload[15],
	}

	// Store in first slot
	s.device.productIDs.Entries[0] = entry
	s.device.productIDs.NumEntries = 1
}

// handleCommandResp processes command responses.
func (s *sh2Protocol) handleCommandResp(payload []byte, timestamp uint32) {
	if len(payload) < 16 {
		return
	}

	// seq := payload[1]
	command := payload[2]
	// commandSeq := payload[3]
	// respSeq := payload[4]

	// Check if this response is for our command
	if s.waiting && command == s.lastCmd {
		s.waiting = false
		// Status is in payload[6]
		// For now, we just acknowledge receipt
	}
}

// handleGetFeatureResp processes get feature responses.
func (s *sh2Protocol) handleGetFeatureResp(payload []byte, timestamp uint32) {
	if len(payload) < 17 {
		return
	}

	// Parse the response
	sensorID := SensorID(payload[1])
	flags := payload[2]
	changeSensitivity := binary.LittleEndian.Uint16(payload[3:5])
	reportInterval := binary.LittleEndian.Uint32(payload[5:9])
	batchInterval := binary.LittleEndian.Uint32(payload[9:13])
	sensorSpecific := binary.LittleEndian.Uint32(payload[13:17])

	// If we're waiting for this sensor's config, store it
	if s.pendingConfigRequest && s.pendingConfigSensor == sensorID {
		s.receivedConfig = SensorConfig{
			ChangeSensitivityEnabled:  flags&featChangeSensitivityEnabled != 0,
			ChangeSensitivityRelative: flags&featChangeSensitivityRelative != 0,
			WakeupEnabled:             flags&featWakeEnabled != 0,
			AlwaysOnEnabled:           flags&featAlwaysOnEnabled != 0,
			ChangeSensitivity:         changeSensitivity,
			ReportInterval:            reportInterval,
			BatchInterval:             batchInterval,
			SensorSpecific:            sensorSpecific,
		}
		s.configReady = true
	}
}

// handleSensor processes sensor report messages.
// The payload can contain multiple sensor reports batched together.
func (s *sh2Protocol) handleSensor(payload []byte, timestamp uint32) {
	cursor := 0
	var referenceDelta uint32

	for cursor < len(payload) {
		if cursor >= len(payload) {
			break
		}

		reportID := payload[cursor]
		reportLen := getReportLen(reportID)

		if reportLen == 0 {
			// Unknown report ID
			break
		}

		if cursor+reportLen > len(payload) {
			// Not enough data for this report
			break
		}

		// Handle special report types
		switch reportID {
		case 0xFB: // SENSORHUB_BASE_TIMESTAMP_REF
			if reportLen >= 5 {
				// Extract timebase (little-endian uint32)
				timebase := binary.LittleEndian.Uint32(payload[cursor+1 : cursor+5])
				referenceDelta = -timebase // Store negative for delta calculation
			}

		case 0xFA: // SENSORHUB_TIMESTAMP_REBASE
			if reportLen >= 5 {
				timebase := binary.LittleEndian.Uint32(payload[cursor+1 : cursor+5])
				referenceDelta += timebase
			}

		case 0xF1: // SENSORHUB_FLUSH_COMPLETED
			// Route to control handler
			s.handleControl(payload[cursor:cursor+reportLen], timestamp)

		default:
			// Regular sensor report
			value, ok := decodeSensor(payload[cursor:cursor+reportLen], timestamp)
			if ok {
				s.device.enqueue(value)
			}
		}

		cursor += reportLen
	}
} // handleExecutable processes executable channel messages.
func (s *sh2Protocol) handleExecutable(payload []byte, timestamp uint32) {
	if len(payload) == 0 {
		return
	}

	reportID := payload[0]

	switch reportID {
	case execDeviceRespResetComplete:
		s.device.lastReset = true
	}
}
