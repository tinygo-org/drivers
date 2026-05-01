package sx128x

import (
	"runtime"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/pin"
)

type Device struct {
	spi      drivers.SPI
	nssPin   pin.Output
	resetPin pin.Output
	busyPin  pin.Input
	spiTxBuf []byte
	spiRxBuf []byte
}

func New(spi drivers.SPI, nssPin pin.Output, resetPin pin.Output, busyPin pin.Input) *Device {
	return &Device{
		spi:      spi,
		nssPin:   nssPin,
		resetPin: resetPin,
		busyPin:  busyPin,
		spiTxBuf: make([]byte, 256), // TODO: optimize buffer size
		spiRxBuf: make([]byte, 256),
	}
}

func (d *Device) Reset() {
	d.resetPin.Set(false)
	time.Sleep(10 * time.Millisecond)
	d.resetPin.Set(true)
	time.Sleep(10 * time.Millisecond)
}

func (d *Device) WaitWhileBusy(timeout time.Duration) error {
	// largest busy period is on boot with around ~400ish this should be more than enough
	now := time.Now()
	for d.busyPin.Get() {
		if time.Since(now) > timeout {
			return ErrBusyPinTimeout
		}
		runtime.Gosched()
	}
	return nil
}

// Get tranceiver status, returns circuit mode and command status
func (d *Device) GetStatus() (CircuitMode, CommandStatus, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, 0, err
	}
	d.nssPin.Set(false)
	status, err := d.spi.Transfer(cmdGetStatus)
	d.nssPin.Set(true)

	if err != nil {
		return 0, 0, err
	}

	circuitMode := (status & circuitModeMask) >> 5
	commandStatus := (status & commandStatusMask) >> 2
	return circuitMode, commandStatus, nil
}

func (d *Device) WriteRegister(addr uint16, data []byte) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdWriteRegister, uint8((addr>>8)&0xFF), uint8(addr&0xFF))
	d.spiTxBuf = append(d.spiTxBuf, data...)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

func (d *Device) ReadRegister(addr uint16) (uint8, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdReadRegister, uint8((addr&0xFF00)>>8), uint8(addr&0x00FF), 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:5]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return d.spiRxBuf[4], nil
}

func (d *Device) WriteBuffer(offset uint8, data []byte) error {
	if len(data) > 256 {
		return errDataTooLong
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdWriteBuffer, offset)
	d.spiTxBuf = append(d.spiTxBuf, data...)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Read data from the payload buffer starting at the given offset with the given length
func (d *Device) ReadBuffer(offset uint8, length uint8) ([]byte, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return nil, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdReadBuffer, offset, 0x00)
	for i := uint8(0); i < length; i++ {
		d.spiTxBuf = append(d.spiTxBuf, 0x00)
	}
	d.spiRxBuf = d.spiRxBuf[:len(d.spiTxBuf)]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return nil, err
	}
	return d.spiRxBuf[3:], nil
}

// Set the device into sleep mode with the given configuration: 0 (no retention), 1 (ram retentation), 2 (buffer retention) or 3 (ram and buffer retention)
func (d *Device) SetSleep(sleepConfig SleepConfig) error {
	if sleepConfig > (SLEEP_DATA_BUFFER_RETAIN | SLEEP_DATA_RAM_RETAIN) {
		return errInvalidSleepConfig
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetSleep, sleepConfig)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Put device into standby mode, 0 (RC) or 1 (XOSC)
func (d *Device) SetStandby(standbyConfig StandbyConfig) error {
	if standbyConfig > STANDBY_XOSC { // XOSC is the highest standby config anything higher is invalid
		return errInvalidStandbyConfig
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetStandby, standbyConfig)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Set the device into Frequency Synthesizer mode
func (d *Device) SetFs() error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetFS)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

func checkPeriodBase(periodBase PeriodBase) error {
	if periodBase > PERIOD_BASE_4_MS { // 4ms is the highest period base anything higher is invalid
		return errInvalidPeriodBase
	}
	return nil
}

// Sets the device in transmit mode, the IRQ status should be cleared before using this command
// timout is determined by periodBase * periodBaseCount
func (d *Device) SetTx(periodBase PeriodBase, periodBaseCount uint16) error {
	err := checkPeriodBase(periodBase)
	if err != nil {
		return err
	}
	err = d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetTx, periodBase, uint8((periodBaseCount>>8)&0xFF), uint8(periodBaseCount&0xFF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Sets the device in receive mode, the IRQ status should be cleared before using this command
// timeout is determined by periodBase * periodBaseCount
func (d *Device) SetRx(periodBase PeriodBase, periodBaseCount uint16) error {
	err := checkPeriodBase(periodBase)
	if err != nil {
		return err
	}
	err = d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetRx, periodBase, uint8((periodBaseCount>>8)&0xFF), uint8(periodBaseCount&0xFF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Sets the device in a continuous receive mode, it enters receive mode with a timeout of periodBase * rxPeriodBaseCount.
// If no packet is received it will enter sleep mode for periodBase * sleepPeriodBaseCount before re-entering receive mode.
// The loop is exited when a packet is received or the device is put into standby mode.
func (d *Device) SetRxDutyCycle(periodBase PeriodBase, rxPeriodBaseCount uint16, sleepPeriodBaseCount uint16) error {
	err := checkPeriodBase(periodBase)
	if err != nil {
		return err
	}
	err = d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetRxDutyCycle, periodBase, uint8((rxPeriodBaseCount&0xFF00)>>8), uint8(rxPeriodBaseCount&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, uint8((sleepPeriodBaseCount&0xFF00)>>8), uint8(sleepPeriodBaseCount&0x00FF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Sets the transceiver into Long Preamble mode, and can only be used with either the LoRa mode and GFSK mode
func (d *Device) SetLongPreamble(enable bool) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetLongPreamble)
	if enable {
		d.spiTxBuf = append(d.spiTxBuf, 1)
	} else {
		d.spiTxBuf = append(d.spiTxBuf, 0)
	}
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Channel activity detection (CAD) is a LoRa specific mode of operation where the device searches for a LoRa signal.
// After search has completed, the device returns to STDBY_RC mode. The length of the search is configured via the SetCadParams() command.
func (d *Device) SetCAD() error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetCAD)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Test command to generate a Continuous Wave (RF tone) at a selected frequency and output power
// The device remains in Tx Continuous Wave until the host sends a mode configuration command.
func (d *Device) SetTxContinuousWave() error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetTxContinuousWave)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Test command to generate an infinite sequence of alternating ‘0’s and ‘1’s in
// GFSK modulation and symbol 0 in LoRa. The device remains in transmit until the host sends a mode configuration command.
func (d *Device) SetTxContinuousPreamble() error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetContinuousPreamble)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// This command allows the transceiver to send a packet at a user programmable time after the end of a packet reception.
// This is useful for Bluetooth Low Energy (BLE) compatibility which requires the transceiver to be able to send back a response 150µs after a packet reception.
func (d *Device) SetAutoTx(timeUs uint16) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetAutoTx, uint8((timeUs&0xFF00)>>8), uint8(timeUs&0x00FF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Modifies the chip behavior so that the state following a Rx or Tx operation is FS and not standby.
// This allows for faster transitions between Rx and/or Tx.
func (d *Device) SetAutoFs(enable bool) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetAutoFS)
	if enable {
		d.spiTxBuf = append(d.spiTxBuf, 1)
	} else {
		d.spiTxBuf = append(d.spiTxBuf, 0)
	}
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Choose between GFSK, LoRa, Ranging, FLRC or BLE packet types, this will affect the available configuration parameters and the structure of the packet
func (d *Device) SetPacketType(packetType PacketType) error {
	if packetType > PACKET_TYPE_BLE { // BLE is the highest packet type anything higher is invalid.
		return errInvalidPacketType
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetPacketType, packetType)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Get the currently configured packet type, this will be 0 (GFSK), 1 (LoRa), 2 (Ranging), 3 (FLRC) or 4 (BLE)
func (d *Device) GetPacketType() (PacketType, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdGetPacketType, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:3]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return d.spiRxBuf[2], nil
}

// Set the RF frequency in Hz, must be between 2.4 GHz and 2.5 GHz
func (d *Device) SetRfFrequency(frequencyHz uint32) error {
	if frequencyHz < 2400000000 {
		return errFrequencyTooLow
	}
	if frequencyHz > 2500000000 {
		return errFrequencyTooHigh
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	rfFrequency := uint32((uint64(frequencyHz) << 18) / 52000000)
	d.spiTxBuf = append(d.spiTxBuf, cmdSetRFFrequency, uint8((rfFrequency>>16)&0xFF), uint8((rfFrequency>>8)&0xFF), uint8(rfFrequency&0xFF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Set the output power in dBm, must be between -18 and 13 dBm, and the ramp time
func (d *Device) SetTxParams(powerdBm int8, rampTime RadioRampTime) error {
	if powerdBm < -18 {
		return errPowerTooLow
	}
	if powerdBm > 13 {
		return errPowerTooHigh
	}

	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	adjustedPower := uint8(powerdBm + 18)
	d.spiTxBuf = append(d.spiTxBuf, cmdSetTxParams, adjustedPower, rampTime)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Set the number of symbols used for channel activity detection which determines the sensitivity of the detection.
// This is only applicable in LoRa mode.
func (d *Device) SetCadParams(cadSymbolNum uint8) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetCADParams, cadSymbolNum)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Set the base address for the internal buffer for Tx and Rx operations.
// When transmitting or receiving data is read from or written to the buffer starting at the given offset.
func (d *Device) SetBufferBaseAddress(txBase uint8, rxBase uint8) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetBufferBaseAddress, txBase, rxBase)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// The arguments to this function depend on the packet type. It is recommended to use the mode specific functions for a better experience.
// BLE & GFSK: BitrateBandwidth, ModulationIndex, ModulationShaping
// FLRC: BitrateBandwidth, CodingRate, ModulationShaping
// LoRa & Ranging: SpreadingFactor, Bandwidth, CodingRate
func (d *Device) SetModulationParams(modParam1, modParam2, modParam3 uint8) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetModulationParams, modParam1, modParam2, modParam3)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

func (d *Device) SetModulationParamsBLE(bitrateBandwidth GFSKBLEBitrateBandwidth, modulationIndex ModulationIndex, modulationShaping ModulationShaping) error {
	return d.SetModulationParams(bitrateBandwidth, modulationIndex, modulationShaping)
}

func (d *Device) SetModulationParamsGFSK(bitrateBandwidth GFSKBLEBitrateBandwidth, modulationIndex ModulationIndex, modulationShaping ModulationShaping) error {
	return d.SetModulationParams(bitrateBandwidth, modulationIndex, modulationShaping)
}

func (d *Device) SetModulationParamsFLRC(bitrateBandwidth FLRCBitrateBandwidth, codingRate FLRCCodingRate, modulationShaping ModulationShaping) error {
	return d.SetModulationParams(bitrateBandwidth, codingRate, modulationShaping)
}

func (d *Device) SetModulationParamsLoRa(spreadingFactor LoRaSpreadingFactor, bandwidth LoRaBandwidth, codingRate LoRaCodingRate) error {
	return d.SetModulationParams(spreadingFactor, bandwidth, codingRate)
}

// The arguments to this function depend on the packet type. It is recommended to use the mode specific functions for a better experience.
// GFSK & FLRC: PreambleLength, SyncWordLength, SyncWordMatch, HeaderType, PayloadLength, CrcLength, Whitening
// BLE: ConnectionState, CrcLength, BleTestPayload, Whitening
// LoRa & Ranging: PreambleLength, HeaderType, PayloadLength, CRC, InvertIQ/chirp invert
func (d *Device) SetPacketParams(param1, param2, param3, param4, param5, param6, param7 uint8) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetPacketParams, param1, param2, param3, param4, param5, param6, param7)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Set GFSK related packet parameters, this assumes the packet type is already set to GFSK.
// - payloadLength:  range of 0-255
func (d *Device) SetPacketParamsGFSK(preambleLength GFSKPreambleLength, syncWordLength GFSKSyncWordLength, syncWordMatch GFSKSyncWordMatch, headerType GFSKHeaderType, payloadLength uint8, crcLength GFSKCrcType, whitening bool) error {
	var whiteningVal uint8
	if whitening {
		whiteningVal = whiteningEnable
	} else {
		whiteningVal = whiteningDisable
	}
	return d.SetPacketParams(preambleLength, syncWordLength, syncWordMatch, headerType, payloadLength, crcLength, whiteningVal)
}

// Set FLRC related packet parameters, this assumes the packet type is already set to FLRC.
// - payloadLength: range of 6-127
func (d *Device) SetPacketParamsFLRC(preambleLength FLRCPreambleLength, syncWordLength FLRCSyncWordLength, syncWordMatch FLRCSyncWordMatch, headerType FLRCHeaderType, payloadLength uint8, crcLength FLRCCrcType) error {
	if payloadLength < 6 {
		return errPayloadLengthTooShort
	}
	if payloadLength > 127 {
		return errPayloadLengthTooLong
	}
	return d.SetPacketParams(preambleLength, syncWordLength, syncWordMatch, headerType, payloadLength, crcLength, whiteningDisable)
}

// Set BLE related packet parameters, this assumes the packet type is already set to BLE.
func (d *Device) SetPacketParamsBLE(connectionState BLEConnectionState, crcLength BLECrcType, bleTestPayload BLETestPayload, whitening bool) error {
	var whiteningVal uint8
	if whitening {
		whiteningVal = whiteningEnable
	} else {
		whiteningVal = whiteningDisable
	}
	return d.SetPacketParams(connectionState, crcLength, bleTestPayload, whiteningVal, 0, 0, 0)
}

// Set LoRa related packet parameters, this assumes the packet type is already set to LoRa.
// - payloadLength: range of 1-255
func (d *Device) SetPacketParamsLoRa(preambleLength uint32, headerType LoRaHeaderType, payloadLength uint8, crcType LoRaCrcType, iqType LoRaIqType) error {
	if payloadLength == 0 {
		return errPayloadLengthTooShort
	}
	exponent, mantissa := getExponentAndMantissa(preambleLength)
	return d.SetPacketParams(uint8(exponent<<4)|mantissa, headerType, payloadLength, crcType, iqType, 0, 0)
}

func getExponentAndMantissa(value uint32) (uint8, uint8) {
	// pulled from RadioLib https://github.com/jgromes/RadioLib/blob/master/src/modules/SX128x/SX128x.cpp
	e := uint8(1)
	m := uint8(1)
	len := uint32(0)
	for e = uint8(1); e <= 15; e++ {
		for m = uint8(1); m <= 15; m++ {
			len = uint32(m) * (uint32(1 << e))
			if len >= value {
				break
			}
		}
		if len >= value {
			break
		}
	}

	return e, m
}

// Get information about the most recent packet received.
// Return the payload length, the offset in the buffer where the payload starts.
func (d *Device) GetRxBufferStatus() (uint8, uint8, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdGetRxBufferStatus, 0x00, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:4]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, 0, err
	}
	return d.spiRxBuf[2], d.spiRxBuf[3], nil
}

// The return type of this function depends on the packet type. Use mode specific function for typed returns.
// BLE, GFSK & FLRC: unused, rssiSync, errors, status, sync
// LoRa & Ranging: rssiSync, SNR
func (d *Device) GetPacketStatus() (uint8, uint8, uint8, uint8, uint8, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdGetPacketStatus, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:7]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	return d.spiRxBuf[2], d.spiRxBuf[3], d.spiRxBuf[4], d.spiRxBuf[5], d.spiRxBuf[6], nil
}

// Get information about the most recent GFSK packet received or transmitted:
// - RSSI of last received packet
// - packet information (each bit represents a different error or status flag)
// - whether the last packet transmission has ended
// - the sync word that was used for the last packet reception (0-3)
func (d *Device) GetPacketStatusGFSK() (float32, GFSKPacketInfo, bool, uint8, error) {
	_, rssiSync, packetInfo, status, sync, err := d.GetPacketStatus()
	if err != nil {
		return 0, 0, false, 0, err
	}
	return float32(int8(rssiSync)) / 2 * -1, GFSKPacketInfo(packetInfo), status != 0, sync, nil
}

// Get information about the most recent BLE packet received or transmitted:
// - RSSI of last received packet
// - packet information (each bit represents a different error or status flag)
// - whether the last packet transmission has ended
// - the sync word that was used for the last packet reception (0-1)
func (d *Device) GetPacketStatusBLE() (float32, BLEPacketInfo, bool, uint8, error) {
	_, rssiSync, packetInfo, status, sync, err := d.GetPacketStatus()
	if err != nil {
		return 0, 0, false, 0, err
	}
	return float32(int8(rssiSync)) / 2 * -1, BLEPacketInfo(packetInfo), status != 0, sync, nil
}

// Get information about the most recent BLE packet received or transmitted:
// - RSSI of last received packet
// - packet information (each bit represents a different error or status flag)
// - PID field of the received packet
// - NO_ACK field of the received packet
// - PID check status of the current packet
// - whether the last packet transmission has ended
// - the sync word that was used for the last packet reception (0-1)
func (d *Device) GetPacketStatusFLRC() (float32, FLRCPacketInfo, uint8, bool, bool, bool, uint8, error) {
	_, rawRSSI, packetInfo, rxTxInfo, sync, err := d.GetPacketStatus()

	rxPid := (rxTxInfo & 0b11000000) >> 6
	noAck := (rxTxInfo & 0b00100000) != 0
	pidCheck := (rxTxInfo & 0b00010000) != 0
	txDone := (rxTxInfo & 0b00000001) != 0

	if err != nil {
		return 0, 0, 0, false, false, false, 0, err
	}
	return float32(int8(rawRSSI)) / 2 * -1, FLRCPacketInfo(packetInfo), rxPid, noAck, pidCheck, txDone, sync, nil
}

// Get information about the most recent LoRa packet received:
// - RSSI of last received packet
// - signal-to-noise ratio (SNR) of last received packet
func (d *Device) GetPacketStatusLoRa() (float32, float32, error) {
	rawRSSI, rawSnr, _, _, _, err := d.GetPacketStatus()
	if err != nil {
		return 0, 0, err
	}
	return float32(int8(rawRSSI)) / 2 * -1, float32(int8(rawSnr)) / 4, nil
}

// Get the instantaneous RSSI value during reception of the packet
func (d *Device) GetRssiInst() (float32, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdGetRSSIInst, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:3]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return float32(int8(d.spiRxBuf[2])) / 2 * -1, nil
}

// Configure the overall IRQ mask and the mapping of individual IRQs to the DIO1, DIO2 and DIO3 pins
func (d *Device) SetDioIrqParams(irqMask IRQMask, dio1Mask IRQMask, dio2Mask IRQMask, dio3Mask IRQMask) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetDIOIRQParams, uint8((irqMask&0xFF00)>>8), uint8(irqMask&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, uint8((dio1Mask&0xFF00)>>8), uint8(dio1Mask&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, uint8((dio2Mask&0xFF00)>>8), uint8(dio2Mask&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, uint8((dio3Mask&0xFF00)>>8), uint8(dio3Mask&0x00FF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Get the current IRQ status.
func (d *Device) GetIrqStatus() (IRQMask, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdGetIRQStatus, 0x00, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:4]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return uint16(d.spiRxBuf[2])<<8 | uint16(d.spiRxBuf[3]), err
}

// Clear the IRQ bits specified in the irqMask.
func (d *Device) ClearIrqStatus(irqMask IRQMask) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdClearIRQStatus, uint8((irqMask&0xFF00)>>8), uint8(irqMask&0x00FF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Switch between the low-dropout regulator (LDO) and the DC-DC converter for internal power regulation.
func (d *Device) SetRegulatorMode(mode RegulatorMode) error {
	if mode > REGULATOR_DC_DC { // DC-DC is the highest regulator mode anything higher is invalid
		return errInvalidRegulatorMode
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetRegulatorMode, mode)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Stores the present context of the radio register values to the Data RAM which will be restored when the device wakes up from sleep mode.
func (d *Device) SetSaveContext() error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmdSetSaveContext)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}
