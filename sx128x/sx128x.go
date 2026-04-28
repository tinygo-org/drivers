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

func (d *Device) WaitWhileBusy() error {
	// largest busy period is on boot with around ~400ish this should be more than enough
	retries := 1000
	for retries > 0 && d.busyPin.Get() {
		runtime.Gosched()
		retries--
	}
	if retries == 0 {
		return errors.New("busy pin timeout")
	}
	return nil
}

func (d *Device) GetStatus() (uint8, uint8, error) {
	err := d.WaitWhileBusy()
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
	err := d.WaitWhileBusy()
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
	err := d.WaitWhileBusy()
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
	err := d.WaitWhileBusy()
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

func (d *Device) ReadBuffer(offset uint8, length uint8) ([]byte, error) {
	err := d.WaitWhileBusy()
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

func (d *Device) SetSleep(sleepConfig uint8) error {
	if sleepConfig > 3 {
		return errors.New("sleep config must be 0 (no retention), 1 (ram retentation), 2 (buffer retention) or 3 (ram and buffer retention)")
	}
	err := d.WaitWhileBusy()
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

func (d *Device) SetStandby(standbyConfig uint8) error {
	if standbyConfig != STANDBY_RC && standbyConfig != STANDBY_XOSC {
		return errors.New("standby config must be 0 (RC) or 1 (XOSC)")
	}
	err := d.WaitWhileBusy()
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

func (d *Device) SetFs() error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetTx(periodBase uint8, periodBaseCount uint16) error {
	err := checkPeriodBase(periodBase)
	if err != nil {
		return err
	}
	err = d.WaitWhileBusy()
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

func (d *Device) SetRx(periodBase uint8, periodBaseCount uint16) error {
	err := checkPeriodBase(periodBase)
	if err != nil {
		return err
	}
	err = d.WaitWhileBusy()
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

func (d *Device) SetRxDutyCycle(periodBase uint8, rxPeriodBaseCount uint16, sleepPeriodBaseCount uint16) error {
	err := checkPeriodBase(periodBase)
	if err != nil {
		return err
	}
	err = d.WaitWhileBusy()
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

func (d *Device) SetLongPreamble(enable bool) error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetCAD() error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetTxContinuousWave() error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetTxContinuousPreamble() error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetAutoTx(time uint16) error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetAutoFs(enable bool) error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetPacketType(packetType uint8) error {
	err := d.WaitWhileBusy()
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

func (d *Device) GetPacketType() (uint8, error) {
	err := d.WaitWhileBusy()
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

func (d *Device) SetRfFrequency(frequency uint32) error {
	if frequency < 2400000000 {
		return errors.New("frequency must be greater than or equal to 2.4 GHz")
	}
	if frequency > 2500000000 {
		return errors.New("frequency must be less than or equal to 2.5 GHz")
	}
	err := d.WaitWhileBusy()
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

func (d *Device) SetTxParams(powerdBm int8, rampTime uint8) error {
	if powerdBm < -18 {
		return errors.New("power in dBm must be greater than or equal to -18")
	}
	if powerdBm > 13 {
		return errors.New("power in dBm must be less than or equal to 13")
	}

	err := d.WaitWhileBusy()
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

func (d *Device) SetCadParams(cadSymbolNum uint8) error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetBufferBaseAddress(txBase uint8, rxBase uint8) error {
	err := d.WaitWhileBusy()
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
	err := d.WaitWhileBusy()
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
	err := d.WaitWhileBusy()
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

// RxBufferStatus: payloadLength, bufferStartPointer
func (d *Device) GetRxBufferStatus() (uint8, uint8, error) {
	err := d.WaitWhileBusy()
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
	err := d.WaitWhileBusy()
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

func (d *Device) GetRssiInst() (int8, error) {
	err := d.WaitWhileBusy()
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

func (d *Device) SetDioIrqParams(irqMask uint16, dio1Mask uint16, dio2Mask uint16, dio3Mask uint16) error {
	err := d.WaitWhileBusy()
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

func (d *Device) GetIrqStatus() (uint16, error) {
	err := d.WaitWhileBusy()
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

func (d *Device) ClearIrqStatus(irqMask uint16) error {
	err := d.WaitWhileBusy()
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

func (d *Device) SetRegulatorMode(mode uint8) error {
	if mode != REGULATOR_LDO && mode != REGULATOR_DC_DC {
		return errors.New("regulator mode must be 0 (LDO) or 1 (DC-DC)")
	}
	err := d.WaitWhileBusy()
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

func (d *Device) SetSaveContext() error {
	err := d.WaitWhileBusy()
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
