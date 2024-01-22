// Package mcp2515 implements a driver for the MCP2515 CAN Controller.
//
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/MCP2515-Stand-Alone-CAN-Controller-with-SPI-20001801J.pdf
//
// Reference: https://github.com/coryjfowler/MCP_CAN_lib
package mcp2515 // import "tinygo.org/x/drivers/mcp2515"

import (
	"errors"
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps MCP2515 SPI CAN Module.
type Device struct {
	spi     SPI
	cs      machine.Pin
	msg     *CANMsg
	mcpMode byte
}

// CANMsg stores CAN message fields.
type CANMsg struct {
	ID   uint32
	Dlc  uint8
	Data []byte
	Ext  bool
	Rtr  bool
}

const (
	bufferSize int = 64
)

// New returns a new MCP2515 driver. Pass in a fully configured SPI bus.
func New(b drivers.SPI, csPin machine.Pin) *Device {
	d := &Device{
		spi: SPI{
			bus: b,
			tx:  make([]byte, 0, bufferSize),
			rx:  make([]byte, 0, bufferSize),
		},
		cs:  csPin,
		msg: &CANMsg{},
	}

	return d
}

// Configure sets up the device for communication.
func (d *Device) Configure() {
	d.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

const beginTimeoutValue int = 10

// Begin starts the CAN controller.
func (d *Device) Begin(speed byte, clock byte) error {
	timeOutCount := 0
	for {
		err := d.init(speed, clock)
		if err == nil {
			break
		}
		timeOutCount++
		if timeOutCount >= beginTimeoutValue {
			return err
		}
	}
	return nil
}

// Received returns true if CAN message is received.
func (d *Device) Received() bool {
	res, err := d.readStatus()
	if err != nil {
		panic(err)
	}
	// if RX STATUS INSTRUCTION  result is not 0x00 (= No RX message)
	// TODO: reconsider this logic
	return (res & mcpStatRxifMask) != 0x00
}

// Rx returns received CAN message.
func (d *Device) Rx() (*CANMsg, error) {
	err := d.readMsg()
	return d.msg, err
}

// Tx transmits CAN Message.
func (d *Device) Tx(canid uint32, dlc uint8, data []byte) error {
	// TODO: add ext, rtrBit, waitSent
	timeoutCount := 0

	var bufNum, res uint8
	var err error
	res = mcpAlltxbusy
	for res == mcpAlltxbusy && (timeoutCount < timeoutvalue) {
		if timeoutCount > 0 {
			time.Sleep(time.Microsecond * 10)
		}
		bufNum, res, err = d.getNextFreeTxBuf()
		if err != nil {
			return err
		}
		timeoutCount++
	}
	if timeoutCount == timeoutvalue {
		return fmt.Errorf("Tx: Tx timeout")
	}
	err = d.writeCANMsg(bufNum, canid, 0, 0, dlc, data)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) init(speed, clock byte) error {
	err := d.Reset()
	if err != nil {
		return err
	}

	if err := d.setCANCTRLMode(modeConfig); err != nil {
		return fmt.Errorf("setCANCTRLMode %s: ", err)
	}
	time.Sleep(time.Millisecond * 10)

	// set baudrate
	if err := d.configRate(speed, clock); err != nil {
		return fmt.Errorf("configRate %s: ", err)
	}
	time.Sleep(time.Millisecond * 10)

	if err := d.initCANBuffers(); err != nil {
		return fmt.Errorf("initCANBuffers: %s ", err)
	}
	if err := d.setRegister(mcpCANINTE, mcpRX0IF|mcpRX1IF); err != nil {
		return fmt.Errorf("setRegister: %s ", err)
	}
	if err := d.modifyRegister(mcpRXB0CTRL, mcpRxbRxMask|mcpRxbBuktMask, mcpRxbRxStdExt|mcpRxbBuktMask); err != nil {
		return fmt.Errorf("modifyRegister: %s ", err)
	}
	if err := d.modifyRegister(mcpRXB1CTRL, mcpRxbRxMask, mcpRxbRxStdExt); err != nil {
		return fmt.Errorf("modifyRegister: %s ", err)
	}

	if err := d.setMode(modeNormal); err != nil {
		return fmt.Errorf("setMode %s: ", err)
	}
	time.Sleep(time.Millisecond * 10)

	return nil
}

// Reset resets mcp2515.
func (d *Device) Reset() error {
	d.cs.Low()
	_, err := d.spi.readWrite(mcpReset)
	d.cs.High()
	// time.Sleep(time.Microsecond * 4)
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 10)

	return nil
}

func (d *Device) setCANCTRLMode(newMode byte) error {
	// If the chip is asleep and we want to change mode then a manual wake needs to be done
	// This is done by setting the wake up interrupt flag
	// This undocumented trick was found at https://github.com/mkleemann/can/blob/master/can_sleep_mcp2515.c
	m, err := d.getMode()
	if err != nil {
		return err
	}
	if m == modeSleep && newMode != modeSleep {
		r, err := d.readRegister(mcpCANINTE)
		if err != nil {
			return err
		}
		wakeIntEnabled := (r & mcpWAKIF) == 0x00
		if !wakeIntEnabled {
			d.modifyRegister(mcpCANINTE, mcpWAKIF, mcpWAKIF)
		}
		// Set wake flag (this does the actual waking up)
		d.modifyRegister(mcpCANINTF, mcpWAKIF, mcpWAKIF)

		// Wait for the chip to exit SLEEP and enter LISTENONLY mode.

		// If the chip is not connected to a CAN bus (or the bus has no other powered nodes) it will sometimes trigger the wake interrupt as soon
		// as it's put to sleep, but it will stay in SLEEP mode instead of automatically switching to LISTENONLY mode.
		// In this situation the mode needs to be manually set to LISTENONLY.

		if err := d.requestNewMode(modeListenOnly); err != nil {
			return err
		}

		// Turn wake interrupt back off if it was originally off
		if !wakeIntEnabled {
			d.modifyRegister(mcpCANINTE, mcpWAKIF, 0)
		}
	}

	// Clear wake flag
	d.modifyRegister(mcpCANINTF, mcpWAKIF, 0)

	return d.requestNewMode(newMode)
}

func (d *Device) setMode(opMode byte) error {
	if opMode != modeSleep {
		d.mcpMode = opMode
	}

	err := d.setCANCTRLMode(opMode)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) getMode() (byte, error) {
	r, err := d.readRegister(mcpCANSTAT)
	if err != nil {
		return 0, err
	}
	return r & modeMask, nil
}

func (d *Device) configRate(speed, clock byte) error {
	var cfg1, cfg2, cfg3 byte
	set := true
	switch clock {
	case Clock16MHz:
		switch speed {
		case CAN5kBps:
			cfg1 = mcp16mHz5kBpsCfg1
			cfg2 = mcp16mHz5kBpsCfg2
			cfg3 = mcp16mHz5kBpsCfg3
		case CAN10kBps:
			cfg1 = mcp16mHz10kBpsCfg1
			cfg2 = mcp16mHz10kBpsCfg2
			cfg3 = mcp16mHz10kBpsCfg3
		case CAN20kBps:
			cfg1 = mcp16mHz20kBpsCfg1
			cfg2 = mcp16mHz20kBpsCfg2
			cfg3 = mcp16mHz20kBpsCfg3
		case CAN25kBps:
			cfg1 = mcp16mHz25kBpsCfg1
			cfg2 = mcp16mHz25kBpsCfg2
			cfg3 = mcp16mHz25kBpsCfg3
		case CAN31k25Bps:
			cfg1 = mcp16mHz31k25BpsCfg1
			cfg2 = mcp16mHz31k25BpsCfg2
			cfg3 = mcp16mHz31k25BpsCfg3
		case CAN33kBps:
			cfg1 = mcp16mHz33kBpsCfg1
			cfg2 = mcp16mHz33kBpsCfg2
			cfg3 = mcp16mHz33kBpsCfg3
		case CAN40kBps:
			cfg1 = mcp16mHz40kBpsCfg1
			cfg2 = mcp16mHz40kBpsCfg2
			cfg3 = mcp16mHz40kBpsCfg3
		case CAN47kBps:
			cfg1 = mcp16mHz47kBpsCfg1
			cfg2 = mcp16mHz47kBpsCfg2
			cfg3 = mcp16mHz47kBpsCfg3
		case CAN50kBps:
			cfg1 = mcp16mHz50kBpsCfg1
			cfg2 = mcp16mHz50kBpsCfg2
			cfg3 = mcp16mHz50kBpsCfg3
		case CAN80kBps:
			cfg1 = mcp16mHz80kBpsCfg1
			cfg2 = mcp16mHz80kBpsCfg2
			cfg3 = mcp16mHz80kBpsCfg3
		case CAN83k3Bps:
			cfg1 = mcp16mHz83k3BpsCfg1
			cfg2 = mcp16mHz83k3BpsCfg2
			cfg3 = mcp16mHz83k3BpsCfg3
		case CAN95kBps:
			cfg1 = mcp16mHz95kBpsCfg1
			cfg2 = mcp16mHz95kBpsCfg2
			cfg3 = mcp16mHz95kBpsCfg3
		case CAN100kBps:
			cfg1 = mcp16mHz100kBpsCfg1
			cfg2 = mcp16mHz100kBpsCfg2
			cfg3 = mcp16mHz100kBpsCfg3
		case CAN125kBps:
			cfg1 = mcp16mHz125kBpsCfg1
			cfg2 = mcp16mHz125kBpsCfg2
			cfg3 = mcp16mHz125kBpsCfg3
		case CAN200kBps:
			cfg1 = mcp16mHz200kBpsCfg1
			cfg2 = mcp16mHz200kBpsCfg2
			cfg3 = mcp16mHz200kBpsCfg3
		case CAN250kBps:
			cfg1 = mcp16mHz250kBpsCfg1
			cfg2 = mcp16mHz250kBpsCfg2
			cfg3 = mcp16mHz250kBpsCfg3
		case CAN500kBps:
			cfg1 = mcp16mHz500kBpsCfg1
			cfg2 = mcp16mHz500kBpsCfg2
			cfg3 = mcp16mHz500kBpsCfg3
		case CAN666kBps:
			cfg1 = mcp16mHz666kBpsCfg1
			cfg2 = mcp16mHz666kBpsCfg2
			cfg3 = mcp16mHz666kBpsCfg3
		case CAN1000kBps:
			cfg1 = mcp16mHz1000kBpsCfg1
			cfg2 = mcp16mHz1000kBpsCfg2
			cfg3 = mcp16mHz1000kBpsCfg3
		default:
			set = false
		}
	case Clock8MHz:
		switch speed {
		case CAN5kBps:
			cfg1 = mcp8mHz5kBpsCfg1
			cfg2 = mcp8mHz5kBpsCfg2
			cfg3 = mcp8mHz5kBpsCfg3
		case CAN10kBps:
			cfg1 = mcp8mHz10kBpsCfg1
			cfg2 = mcp8mHz10kBpsCfg2
			cfg3 = mcp8mHz10kBpsCfg3
		case CAN20kBps:
			cfg1 = mcp8mHz20kBpsCfg1
			cfg2 = mcp8mHz20kBpsCfg2
			cfg3 = mcp8mHz20kBpsCfg3
		case CAN31k25Bps:
			cfg1 = mcp8mHz31k25BpsCfg1
			cfg2 = mcp8mHz31k25BpsCfg2
			cfg3 = mcp8mHz31k25BpsCfg3
		case CAN40kBps:
			cfg1 = mcp8mHz40kBpsCfg1
			cfg2 = mcp8mHz40kBpsCfg2
			cfg3 = mcp8mHz40kBpsCfg3
		case CAN50kBps:
			cfg1 = mcp8mHz50kBpsCfg1
			cfg2 = mcp8mHz50kBpsCfg2
			cfg3 = mcp8mHz50kBpsCfg3
		case CAN80kBps:
			cfg1 = mcp8mHz80kBpsCfg1
			cfg2 = mcp8mHz80kBpsCfg2
			cfg3 = mcp8mHz80kBpsCfg3
		case CAN100kBps:
			cfg1 = mcp8mHz100kBpsCfg1
			cfg2 = mcp8mHz100kBpsCfg2
			cfg3 = mcp8mHz100kBpsCfg3
		case CAN125kBps:
			cfg1 = mcp8mHz125kBpsCfg1
			cfg2 = mcp8mHz125kBpsCfg2
			cfg3 = mcp8mHz125kBpsCfg3
		case CAN200kBps:
			cfg1 = mcp8mHz200kBpsCfg1
			cfg2 = mcp8mHz200kBpsCfg2
			cfg3 = mcp8mHz200kBpsCfg3
		case CAN250kBps:
			cfg1 = mcp8mHz250kBpsCfg1
			cfg2 = mcp8mHz250kBpsCfg2
			cfg3 = mcp8mHz250kBpsCfg3
		case CAN500kBps:
			cfg1 = mcp8mHz500kBpsCfg1
			cfg2 = mcp8mHz500kBpsCfg2
			cfg3 = mcp8mHz500kBpsCfg3
		case CAN1000kBps:
			cfg1 = mcp8mHz1000kBpsCfg1
			cfg2 = mcp8mHz1000kBpsCfg2
			cfg3 = mcp8mHz1000kBpsCfg3
		default:
			set = false
		}
	default:
		set = false
	}
	if !set {
		return errors.New("invalid parameter")
	}
	if err := d.setRegister(mcpCNF1, cfg1); err != nil {
		return err
	}
	if err := d.setRegister(mcpCNF2, cfg2); err != nil {
		return err
	}
	if err := d.setRegister(mcpCNF3, cfg3); err != nil {
		return err
	}

	return nil
}

func (d *Device) initCANBuffers() error {
	a1 := byte(mcpTXB0CTRL)
	a2 := byte(mcpTXB1CTRL)
	a3 := byte(mcpTXB2CTRL)
	for i := 0; i < 14; i++ {
		if err := d.setRegister(a1, 0); err != nil {
			return err
		}
		if err := d.setRegister(a2, 0); err != nil {
			return err
		}
		if err := d.setRegister(a3, 0); err != nil {
			return err
		}
		a1++
		a2++
		a3++
	}

	if err := d.setRegister(mcpRXB0CTRL, 0); err != nil {
		return err
	}
	if err := d.setRegister(mcpRXB1CTRL, 0); err != nil {
		return err
	}

	return nil
}

func (d *Device) readMsg() error {
	status, err := d.readRxTxStatus()
	if err != nil {
		return err
	}
	if (status & mcpRX0IF) == 0x01 {
		err := d.readRxBuffer(mcpReadRx0)
		if err != nil {
			return err
		}
	} else if (status & mcpRX1IF) == 0x02 {
		err := d.readRxBuffer(mcpReadRx1)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("readMsg: nothing is received")
	}

	return nil
}

func (d *Device) readRxBuffer(loadAddr uint8) error {
	msg := d.msg
	d.cs.Low()
	defer d.cs.High()
	_, err := d.spi.readWrite(loadAddr)
	if err != nil {
		return err
	}
	err = d.spi.read(4)
	if err != nil {
		return err
	}
	buf := d.spi.rx
	msg.ID = uint32((uint32(buf[0]) << 3) + (uint32(buf[1]) >> 5))
	msg.Ext = false
	if (buf[1] & mcpTxbExideM) == mcpTxbExideM {
		// extended id
		msg.ID = uint32(uint32(msg.ID<<2) + uint32(buf[1]&0x03))
		msg.ID = uint32(uint32(msg.ID<<8) + uint32(buf[2]))
		msg.ID = uint32(uint32(msg.ID<<8) + uint32(buf[3]))
		msg.Ext = true
	}
	err = d.spi.read(1)
	if err != nil {
		return err
	}
	msgSize := d.spi.rx[0]
	msg.Dlc = uint8(msgSize & mcpDlcMask)
	msg.Rtr = false
	if (msgSize & mcpRtrMask) == 0x40 {
		msg.Rtr = true
	}
	readLen := uint8(canMaxCharInMessage)
	if msg.Dlc < canMaxCharInMessage {
		readLen = msg.Dlc
	}
	err = d.spi.read(int(readLen))
	if err != nil {
		return err
	}
	msg.Data = d.spi.rx

	return err
}

func (d *Device) getNextFreeTxBuf() (uint8, uint8, error) {
	status, err := d.readStatus()
	if err != nil {
		return 0, mcpAlltxbusy, err
	}
	status &= mcpStatTxPendingMask

	bufNum := uint8(0x00)

	if status == mcpStatTxPendingMask {
		return 0, mcpAlltxbusy, nil
	}

	for i := 0; i < int(mcpNTxbuffers-nReservedTx(0)); i++ {
		if (status & txStatusPendingFlag(uint8(i))) == 0 {
			bufNum = txCtrlReg(uint8(i)) + 1
			d.modifyRegister(mcpCANINTF, txIfFlag(uint8(i)), 0)
			return bufNum, mcp2515Ok, nil
		}
	}

	return 0, mcpAlltxbusy, nil
}

func (d *Device) writeCANMsg(bufNum uint8, canid uint32, ext, rtrBit, dlc uint8, data []byte) error {
	d.cs.Low()
	defer d.cs.High()
	_, err := d.spi.readWrite(txSidhToLoad(bufNum))
	if err != nil {
		return err
	}
	err = d.spi.clearBuffer(tx)
	if err != nil {
		return err
	}
	err = d.spi.setTxBufData(canid, ext, rtrBit, dlc, data)
	if err != nil {
		return err
	}
	err = d.spi.write()
	if err != nil {
		return err
	}
	// Since cs.Low and cs.High are executed in d.startTransmission,
	// it is necessary to set cs.High once to separate the instruction of mcp2515.
	d.cs.High()

	err = d.startTransmission(bufNum)
	if err != nil {
		return err
	}

	return nil
}

func (s *SPI) setTxBufData(canid uint32, ext, rtrBit, dlc uint8, data []byte) error {
	canid = canid & 0x0FFFF
	if ext == 1 {
		// TODO: add Extended ID
		err := s.setTxData(0)
		if err != nil {
			return err
		}
		err = s.setTxData(0)
		if err != nil {
			return err
		}
		err = s.setTxData(0)
		if err != nil {
			return err
		}
		err = s.setTxData(0)
		if err != nil {
			return err
		}
	} else {
		err := s.setTxData(byte(canid >> 3))
		if err != nil {
			return err
		}
		err = s.setTxData(byte((canid & 0x07) << 5))
		if err != nil {
			return err
		}
		err = s.setTxData(0)
		if err != nil {
			return err
		}
		err = s.setTxData(0)
		if err != nil {
			return err
		}
	}
	if rtrBit == 1 {
		dlc |= mcpRtrMask
	} else {
		dlc |= (0)
	}
	err := s.setTxData(dlc)
	if err != nil {
		return err
	}
	for _, d := range data {
		err := s.setTxData(d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) startTransmission(bufNum uint8) error {
	d.cs.Low()
	_, err := d.spi.readWrite(txSidhToRTS(bufNum))
	d.cs.High()
	if err != nil {
		return err
	}

	return nil
}

func nReservedTx(number uint8) uint8 {
	if number < mcpNTxbuffers {
		return number
	}
	return mcpNTxbuffers - 1
}

func txStatusPendingFlag(i uint8) uint8 {
	ret := uint8(0)
	switch i {
	case 0:
		ret = mcpStatTx0Pending
	case 1:
		ret = mcpStatTx1Pending
	case 2:
		ret = mcpStatTx2Pending
	}
	return ret
}

func txCtrlReg(status uint8) uint8 {
	ret := uint8(0)
	switch status {
	case 0:
		ret = mcpTXB0CTRL
	case 1:
		ret = mcpTXB1CTRL
	case 2:
		ret = mcpTXB2CTRL
	}
	return ret
}

func txIfFlag(i uint8) uint8 {
	ret := uint8(0)
	switch i {
	case 0:
		ret = mcpTX0IF
	case 1:
		ret = mcpTX1IF
	case 2:
		ret = mcpTX2IF
	}
	return ret
}

func txSidhToSidh(i uint8) uint8 {
	ret := uint8(0)
	switch i {
	case mcpTX0IF:
		ret = mcpTXB0SIDH
	case mcpTX1IF:
		ret = mcpTXB1SIDH
	case mcpTX2IF:
		ret = mcpTXB2SIDH
	}
	return ret
}

func txSidhToRTS(i uint8) uint8 {
	ret := uint8(0)
	switch i {
	case mcpTXB0SIDH:
		ret = mcpRtsTx0
	case mcpTXB1SIDH:
		ret = mcpRtsTx1
	case mcpTXB2SIDH:
		ret = mcpRtsTx2
	}
	return ret
}

func txSidhToLoad(i uint8) uint8 {
	ret := uint8(0)
	switch i {
	case mcpTXB0SIDH:
		ret = mcpLoadTx0
	case mcpTXB1SIDH:
		ret = mcpLoadTx1
	case mcpTXB2SIDH:
		ret = mcpLoadTx2
	}
	return ret
}

func (d *Device) setRegister(addr, value byte) error {
	d.cs.Low()
	defer d.cs.High()
	_, err := d.spi.readWrite(mcpWrite)
	if err != nil {
		return err
	}
	_, err = d.spi.readWrite(addr)
	if err != nil {
		return err
	}
	_, err = d.spi.readWrite(value)
	if err != nil {
		return err
	}
	// time.Sleep(time.Microsecond * 4)

	return nil
}

func (d *Device) readRegister(addr byte) (byte, error) {
	d.cs.Low()
	defer d.cs.High()
	_, err := d.spi.readWrite(mcpRead)
	if err != nil {
		return 0, err
	}
	_, err = d.spi.readWrite(addr)
	if err != nil {
		return 0, err
	}
	err = d.spi.read(1)
	if err != nil {
		return 0, err
	}
	// time.Sleep(time.Microsecond * 4)
	return d.spi.rx[0], nil
}

func (d *Device) modifyRegister(addr, mask, data byte) error {
	d.cs.Low()
	defer d.cs.High()
	_, err := d.spi.readWrite(mcpBitMod)
	if err != nil {
		return err
	}
	_, err = d.spi.readWrite(addr)
	if err != nil {
		return err
	}
	_, err = d.spi.readWrite(mask)
	if err != nil {
		return err
	}
	_, err = d.spi.readWrite(data)
	if err != nil {
		return err
	}
	// time.Sleep(time.Microsecond * 4)

	return nil
}

func (d *Device) requestNewMode(newMode byte) error {
	s := time.Now()
	for {
		err := d.modifyRegister(mcpCANCTRL, modeMask, newMode)
		if err != nil {
			return err
		}
		r, err := d.readRegister(mcpCANSTAT)
		if err != nil {
			return err
		}
		if r&modeMask == newMode {
			return nil
		} else if e := time.Now(); e.Sub(s) > 200*time.Millisecond {
			return errors.New("requestNewMode max time expired")
		}
	}
}

func (d *Device) readStatus() (byte, error) {
	d.cs.Low()
	defer d.cs.High()
	_, err := d.spi.readWrite(mcpReadStatus)
	if err != nil {
		return 0, err
	}
	err = d.spi.read(1)
	if err != nil {
		return 0, err
	}

	return d.spi.rx[0], nil
}

func (d *Device) readRxTxStatus() (byte, error) {
	status, err := d.readStatus()
	if err != nil {
		return 0, err
	}
	ret := status & (mcpStatTxifMask | mcpStatRxifMask)
	if (status & mcpStatTx0if) == 0x08 {
		ret |= mcpTX0IF
	}
	if (status & mcpStatTx1if) == 0x20 {
		ret |= mcpTX1IF
	}
	if (status & mcpStatTx2if) == 0x80 {
		ret |= mcpTX2IF
	}
	ret |= ret & mcpStatRxifMask

	return ret, nil
}

type SPI struct {
	bus drivers.SPI
	tx  []byte
	rx  []byte
}

const (
	tx = iota
	rx
)

func (s *SPI) readWrite(w byte) (byte, error) {
	return s.bus.Transfer(w)
}

func (s *SPI) read(readLength int) error {
	err := s.clearBuffer(rx)
	if err != nil {
		return err
	}
	err = s.setBufferLength(readLength, rx)
	if err != nil {
		return err
	}
	return s.bus.Tx(nil, s.rx)
}

func (s *SPI) write() error {
	return s.bus.Tx(s.tx, nil)
}

func (s *SPI) clearBuffer(dir int) error { return s.setBufferLength(0, dir) }

func (s *SPI) setBufferLength(length int, dir int) error {
	if dir == tx {
		if length > cap(s.tx) {
			return fmt.Errorf("length is longer than capacity")
		}
		s.tx = s.tx[:length]
	} else if dir == rx {
		if length > cap(s.rx) {
			return fmt.Errorf("length is longer than capacity")
		}
		s.rx = s.rx[:length]
	} else {
		return fmt.Errorf("invalid direction")
	}
	return nil
}

func (s *SPI) setTxData(data byte) error {
	if len(s.tx) >= bufferSize {
		return fmt.Errorf("cannot expand buffer (to avoid memory allocation)")
	}
	s.tx = append(s.tx, data)

	return nil
}

func (d *Device) dumpMode() error {
	m, err := d.getMode()
	if err != nil {
		return err
	}
	fmt.Printf("Mode: %02X\r\n", m)

	return nil
}

func (d *Device) dumpRegister(addr byte) error {
	r, err := d.readRegister(addr)
	if err != nil {
		return err
	}
	fmt.Printf("Register: %02X = %02X\r\n", addr, r)

	return nil
}
