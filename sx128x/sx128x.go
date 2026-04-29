package sx128x

import (
	"errors"
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
		spiTxBuf: make([]byte, 255), // TODO: optimize buffer size
		spiRxBuf: make([]byte, 255),
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
			return errors.New("busy pin timeout")
		}
		runtime.Gosched()
	}
	return nil
}

// Get tranceiver status, returns circuit mode and command status
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, 0, err
	}
	d.nssPin.Set(false)
	status, err := d.spi.Transfer(CMD_GET_STATUS)
	d.nssPin.Set(true)

	if err != nil {
		return 0, 0, err
	}

	circuitMode := (status & CIRCUIT_MODE_MASK) >> 5
	commandStatus := (status & COMMAND_STATUS_MASK) >> 2
	return circuitMode, commandStatus, nil
}

func (d *Device) WriteRegister(addr uint16, data []byte) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_WRITE_REGISTER, uint8((addr>>8)&0xFF), uint8(addr&0xFF))
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_READ_REGISTER, uint8((addr&0xFF00)>>8), uint8(addr&0x00FF), 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:5]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return d.spiRxBuf[4], nil
}

func (d *Device) WriteBuffer(offset uint8, data []byte) error {
	if len(data) > 255 {
		return errors.New("length of data over max length of 255")
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_WRITE_BUFFER, offset)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_READ_BUFFER, offset, 0x00)
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
	if sleepConfig > 3 {
		return errors.New("sleep config must be 0 (no retention), 1 (ram retentation), 2 (buffer retention) or 3 (ram and buffer retention)")
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_SLEEP, sleepConfig)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Put device into standby mode, 0 (RC) or 1 (XOSC)
	if standbyConfig != STANDBY_RC && standbyConfig != STANDBY_XOSC {
		return errors.New("standby config must be 0 (RC) or 1 (XOSC)")
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_STANDBY, standbyConfig)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_FS)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

func checkPeriodBase(periodBase uint8) error {
	if periodBase != PERIOD_BASE_15_625_US && periodBase != PERIOD_BASE_62_5_US && periodBase != PERIOD_BASE_1_MS && periodBase != PERIOD_BASE_4_MS {
		return errors.New("period base must be 0, 1, 2 or 4")
	}
	return nil
}

// Sets the device in transmit mode, the IRQ status should be cleared before using this command
// timout is determined by periodBase * periodBaseCount
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_TX, periodBase, uint8((periodBaseCount>>8)&0xFF), uint8(periodBaseCount&0xFF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Sets the device in receive mode, the IRQ status should be cleared before using this command
// timeout is determined by periodBase * periodBaseCount
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_RX, periodBase, uint8((periodBaseCount>>8)&0xFF), uint8(periodBaseCount&0xFF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Sets the device in a continuous receive mode, it enters receive mode with a timeout of periodBase * rxPeriodBaseCount.
// If no packet is received it will enter sleep mode for periodBase * sleepPeriodBaseCount before re-entering receive mode.
// The loop is exited when a packet is received or the device is put into standby mode.
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_RX_DUTY_CYCLE, periodBase, uint8((rxPeriodBaseCount&0xFF00)>>8), uint8(rxPeriodBaseCount&0x00FF))
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_LONG_PREAMBLE)
	if enable {
		d.spiTxBuf = append(d.spiTxBuf, LONG_PREAMBLE_ENABLE)
	} else {
		d.spiTxBuf = append(d.spiTxBuf, LONG_PREAMBLE_DISABLE)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_CAD)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_TX_CONTINUOUS_WAVE)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_TX_CONTINUOUS_PREAMBLE)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// This command allows the transceiver to send a packet at a user programmable time after the end of a packet reception.
// This is useful for Bluetooth Low Energy (BLE) compatibility which requires the transceiver to be able to send back a response 150µs after a packet reception.
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_AUTO_TX, uint8((time&0xFF00)>>8), uint8(time&0x00FF))
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_AUTO_FS)
	if enable {
		d.spiTxBuf = append(d.spiTxBuf, AUTO_FS_ENABLE)
	} else {
		d.spiTxBuf = append(d.spiTxBuf, AUTO_FS_DISABLE)
	}
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Choose between GFSK, LoRa, Ranging, FLRC or BLE packet types, this will affect the available configuration parameters and the structure of the packet
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_PACKET_TYPE, packetType)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Get the currently configured packet type, this will be 0 (GFSK), 1 (LoRa), 2 (Ranging), 3 (FLRC) or 4 (BLE)
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_GET_PACKET_TYPE, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:3]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return d.spiRxBuf[2], nil
}

// Set the RF frequency in Hz, must be between 2.4 GHz and 2.5 GHz
		return errors.New("frequency must be greater than or equal to 2.4 GHz")
	}
	if frequency > 2500000000 {
		return errors.New("frequency must be less than or equal to 2.5 GHz")
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	freq := uint32((uint64(frequency) << 18) / 52000000)
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_RF_FREQUENCY, uint8((freq>>16)&0xFF), uint8((freq>>8)&0xFF), uint8(freq&0xFF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Set the output power in dBm, must be between -18 and 13 dBm, and the ramp time
	if powerdBm < -18 {
		return errors.New("power in dBm must be greater than or equal to -18")
	}
	if powerdBm > 13 {
		return errors.New("power in dBm must be less than or equal to 13")
	}

	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	adjustedPower := uint8(powerdBm + 18)
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_TX_PARAMS, adjustedPower, rampTime)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_CAD_PARAMS, cadSymbolNum)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_BUFFER_BASE_ADDRESS, txBase, rxBase)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_MODULATION_PARAMS, modParam1, modParam2, modParam3)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

func (d *Device) SetModulationParamsBLE(bitrateBandwidth uint8, modulationIndex uint8, modulationShaping uint8) error {
	return d.SetModulationParams(bitrateBandwidth, modulationIndex, modulationShaping)
}

func (d *Device) SetModulationParamsGFSK(bitrateBandwidth uint8, modulationIndex uint8, modulationShaping uint8) error {
	return d.SetModulationParams(bitrateBandwidth, modulationIndex, modulationShaping)
}

func (d *Device) SetModulationParamsFLRC(bitrateBandwidth uint8, codingRate uint8, modulationShaping uint8) error {
	return d.SetModulationParams(bitrateBandwidth, codingRate, modulationShaping)
}

func (d *Device) SetModulationParamsLoRa(spreadingFactor uint8, bandwidth uint8, codingRate uint8) error {
	return d.SetModulationParams(spreadingFactor, bandwidth, codingRate)
}

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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_PACKET_PARAMS, param1, param2, param3, param4, param5, param6, param7)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

func (d *Device) SetPacketParamsGFSK(preambleLength uint8, syncWordLength uint8, syncWordMatch uint8, headerType uint8, payloadLength uint8, crcLength uint8, whitening uint8) error {
	return d.SetPacketParams(preambleLength, syncWordLength, syncWordMatch, headerType, payloadLength, crcLength, whitening)
}

func (d *Device) SetPacketParamsFLRC(preambleLength uint8, syncWordLength uint8, syncWordMatch uint8, headerType uint8, payloadLength uint8, crcLength uint8, whitening uint8) error {
	return d.SetPacketParams(preambleLength, syncWordLength, syncWordMatch, headerType, payloadLength, crcLength, whitening)
}

func (d *Device) SetPacketParamsBLE(connectionState uint8, crcLength uint8, bleTestPayload uint8, whitening uint8) error {
	return d.SetPacketParams(connectionState, crcLength, bleTestPayload, whitening, 0, 0, 0)
}

func (d *Device) SetPacketParamsLoRa(preambleLength uint32, headerType uint8, payloadLength uint8, crcType uint8, iqType uint8) error {
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_GET_RX_BUFFER_STATUS, 0x00, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:4]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, 0, err
	}
	return d.spiRxBuf[2], d.spiRxBuf[3], nil
}

func (d *Device) GetPacketStatus() (uint8, uint8, uint8, uint8, uint8, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_GET_PACKET_STATUS, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:7]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	return d.spiRxBuf[2], d.spiRxBuf[3], d.spiRxBuf[4], d.spiRxBuf[5], d.spiRxBuf[6], nil
}

// Get the instantaneous RSSI value during reception of the packet
func (d *Device) GetRssiInst() (int8, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_GET_RSSI_INST, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:3]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return int8(d.spiRxBuf[2]/2) * -1, nil
}

// Configure the overall IRQ mask and the mapping of individual IRQs to the DIO1, DIO2 and DIO3 pins
func (d *Device) SetDioIrqParams(irqMask uint16, dio1Mask uint16, dio2Mask uint16, dio3Mask uint16) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_DIO_IRQ_PARAMS, uint8((irqMask&0xFF00)>>8), uint8(irqMask&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, uint8((dio1Mask&0xFF00)>>8), uint8(dio1Mask&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, uint8((dio2Mask&0xFF00)>>8), uint8(dio2Mask&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, uint8((dio3Mask&0xFF00)>>8), uint8(dio3Mask&0x00FF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Get the current IRQ status.
func (d *Device) GetIrqStatus() (uint16, error) {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return 0, err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_GET_IRQ_STATUS, 0x00, 0x00, 0x00)
	d.spiRxBuf = d.spiRxBuf[:4]
	err = d.spi.Tx(d.spiTxBuf, d.spiRxBuf)
	d.nssPin.Set(true)
	if err != nil {
		return 0, err
	}
	return uint16(d.spiRxBuf[2])<<8 | uint16(d.spiRxBuf[3]), err
}

// Clear the IRQ bits specified in the irqMask.
func (d *Device) ClearIrqStatus(irqMask uint16) error {
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_CLEAR_IRQ_STATUS, uint8((irqMask&0xFF00)>>8), uint8(irqMask&0x00FF))
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}

// Switch between the low-dropout regulator (LDO) and the DC-DC converter for internal power regulation.
	if mode != REGULATOR_LDO && mode != REGULATOR_DC_DC {
		return errors.New("regulator mode must be 0 (LDO) or 1 (DC-DC)")
	}
	err := d.WaitWhileBusy(time.Second)
	if err != nil {
		return err
	}
	d.nssPin.Set(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_REGULATOR_MODE, mode)
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
	d.spiTxBuf = append(d.spiTxBuf, CMD_SET_SAVE_CONTEXT)
	err = d.spi.Tx(d.spiTxBuf, nil)
	d.nssPin.Set(true)
	return err
}
