// Driver for the P1AM-100 base controller.
//
// This is an embedded device on the P1AM-100 board.
// Based on v1.0.1 of the Arduino library: https://github.com/facts-engineering/P1AM/tree/1.0.1

package p1am

import (
	"encoding/binary"
	"errors"
	"fmt"
	"machine"
	"time"
)

type P1AM struct {
	bus                                        machine.SPI
	slaveSelectPin, slaveAckPin, baseEnablePin machine.Pin

	// SkipAutoConfig will skip loading a default configuration into each module.
	SkipAutoConfig bool

	Slots int
	// Access slots via Slot()
	slots []Slot
}

var Controller = P1AM{
	bus:            machine.SPI0,
	slaveSelectPin: machine.BASE_SLAVE_SELECT_PIN,
	slaveAckPin:    machine.BASE_SLAVE_ACK_PIN,
	baseEnablePin:  machine.BASE_ENABLE_PIN,
}

type baseSlotConstants struct {
	DI, DO, AI, AO, Status, Config, DataSize byte
}

func (p *P1AM) Initialize() error {
	p.slaveSelectPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	p.slaveAckPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	p.baseEnablePin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	if err := p.bus.Configure(machine.SPIConfig{
		Frequency: 1000000,
		Mode:      2,
		LSBFirst:  false,
	}); err != nil {
		return err
	}

	p.SetEnabled(true)
	time.Sleep(100 * time.Millisecond)

	if err := p.waitAck(5 * time.Second); err != nil {
		return errors.New("no base controller activity; check external supply connection")
	}

	for i := 0; i < 5; i++ {
		if err := p.handleHDR(MOD_HDR); err == nil {
			time.Sleep(5 * time.Millisecond)
			slots, err := p.spiSendRecvByte(0xFF)
			if err == nil && slots > 0 && slots <= 15 {
				p.Slots = int(slots)
				break
			}
		}
		if i > 2 {
			// Try restarting the base controller
			p.SetEnabled(false)
			time.Sleep(10 * time.Millisecond)
			p.SetEnabled(true)
			time.Sleep(10 * time.Millisecond)
		}
	}
	if p.Slots <= 0 || p.Slots > 15 {
		return errors.New("zero modules in the base")
	}

	moduleIDs := make([]uint32, p.Slots)

	p.waitAck(200 * time.Millisecond)
	if err := binary.Read(p, binary.LittleEndian, &moduleIDs); err != nil {
		return err
	}

	baseConstants := make([]baseSlotConstants, p.Slots)
	p.slots = make([]Slot, p.Slots)

	for i := 1; i <= p.Slots; i++ {
		slot := p.Slot(i)
		slot.p = p
		slot.slot = byte(i)
		slot.ID = moduleIDs[i-1]
		// What if 0xFFFFFFFF isn't at position -2?
		slot.Props = &modules[len(modules)-2]
		for j := 0; j < len(modules); j++ {
			if modules[j].ModuleID == slot.ID {
				slot.Props = &modules[j]
			}
			bc := &baseConstants[i-1]
			bc.DI = slot.Props.DI
			bc.DO = slot.Props.DO
			bc.AI = slot.Props.AI
			bc.AO = slot.Props.AO
			bc.Status = slot.Props.Status
			bc.Config = slot.Props.Config
			bc.DataSize = slot.Props.DataSize
		}
	}

	p.waitAck(200 * time.Millisecond)
	if err := binary.Write(p, binary.LittleEndian, &baseConstants); err != nil {
		return err
	}

	if !p.SkipAutoConfig {
		for i := 1; i <= p.Slots; i++ {
			s := p.Slot(i)
			if s.Props.Config > 0 {
				cfg := defaultConfig[s.ID]
				if cfg != nil {
					s.Configure(cfg)
				}
			}
		}
	}

	return nil
}

func (p *P1AM) Version() ([3]byte, error) {
	if err := p.handleHDR(VERSION_HDR); err != nil {
		return [3]byte{}, err
	}
	var buf [4]byte
	if err := p.spiSendRecvBuf(nil, buf[:]); err != nil {
		return [3]byte{}, err
	}
	return [3]byte{
		byte(buf[1] >> 4),
		byte(buf[1] & 0xF),
		byte(buf[0]),
	}, p.dataSync()
}

func (p *P1AM) Active() (bool, error) {
	if _, err := p.spiSendRecvByte(ACTIVE_HDR); err != nil {
		return false, err
	}
	if err := p.waitAck(200 * time.Millisecond); err != nil {
		return false, err
	}
	buf, err := p.spiSendRecvByte(DUMMY)
	defer p.dataSync()
	return buf != 0, err
}

const wdToggleTime = 100 * time.Millisecond

func (p *P1AM) ConfigureWatchdog(interval time.Duration, reset bool) error {
	ms := interval / time.Millisecond
	toggleMs := wdToggleTime / time.Millisecond
	resetB := byte(0)
	if reset {
		resetB = 1
	}
	buf := [6]byte{
		CONFIGWD_HDR,
		byte(ms),
		byte(ms >> 8),
		byte(toggleMs),
		byte(toggleMs >> 8),
		resetB,
	}
	if err := p.spiSendRecvBuf(buf[:], nil); err != nil {
		return err
	}
	return p.dataSync()
}

func (p *P1AM) sendWatchdog(hdr byte) error {
	if _, err := p.spiSendRecvByte(hdr); err != nil {
		return err
	}
	if err := p.waitAck(200 * time.Millisecond); err != nil {
		return err
	}
	if _, err := p.spiSendRecvByte(DUMMY); err != nil {
		return err
	}
	return p.dataSync()
}

func (p *P1AM) StartWatchdog() error {
	return p.sendWatchdog(STARTWD_HDR)
}

func (p *P1AM) StopWatchdog() error {
	return p.sendWatchdog(STOPWD_HDR)
}

func (p *P1AM) PetWatchdog() error {
	return p.sendWatchdog(PETWD_HDR)
}

func (p *P1AM) Slot(i int) *Slot {
	if i < 1 || i > p.Slots {
		return nil
	}
	return &p.slots[i-1]
}

type Slot struct {
	p    *P1AM
	slot byte
	ID   uint32
	// TODO: Embed this?
	Props *ModuleProps
}

func (s *Slot) Configure(data []byte) error {
	if s == nil {
		return errors.New("invalid slot")
	}
	if len(data) != int(s.Props.Config) {
		return fmt.Errorf("expected %d config bytes, got %d", s.Props.Config, len(data))
	}

	if len(data) == 0 {
		return errors.New("no config bytes")
	}

	out := make([]byte, len(data)+2)
	out[0] = CFG_HDR
	out[1] = s.slot
	copy(out[2:], data)

	if err := s.p.spiSendRecvBuf(out, nil); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	s.p.dataSync()
	s.p.dataSync()
	return nil
}

func (s *Slot) ReadDiscrete() (uint32, error) {
	if s == nil {
		return 0, errors.New("invalid slot")
	}
	bytes := s.Props.DI
	out := [2]byte{
		READ_DISCRETE_HDR,
		s.slot,
	}
	if err := s.p.spiSendRecvBuf(out[:], nil); err != nil {
		return 0, err
	}
	if err := s.p.waitAck(200 * time.Millisecond); err != nil {
		return 0, err
	}
	var data [4]byte
	if err := s.p.spiSendRecvBuf(nil, data[:bytes]); err != nil {
		return 0, err
	}
	err := s.p.dataSync()
	return binary.LittleEndian.Uint32(data[:]), err
}

func (s *Slot) WriteDiscrete(value uint32) error {
	return s.writeDiscrete(0, value)
}

func (s *Slot) writeDiscrete(channel byte, value uint32) error {
	if s == nil {
		return errors.New("invalid slot")
	}
	bytes := s.Props.DO
	buf := [7]byte{
		WRITE_DISCRETE_HDR,
		s.slot,
		channel,
	}
	binary.LittleEndian.PutUint32(buf[3:], value)
	out := buf[:3+bytes]
	if channel != 0 {
		out = buf[:4]
		out[3] &= 1
	}
	if err := s.p.spiSendRecvBuf(out, nil); err != nil {
		return err
	}
	return s.p.dataSync()
}

type Channel struct {
	s       *Slot
	channel int
}

func (s *Slot) Channel(channel int) Channel {
	return Channel{
		s:       s,
		channel: channel,
	}
}

func (c Channel) ReadDiscrete() (bool, error) {
	if c.channel < 1 || c.channel > int(c.s.Props.DI)*8 {
		return false, errors.New("invalid channel")
	}
	data, err := c.s.ReadDiscrete()
	return (data>>(c.channel-1))&1 == 1, err
}

func (c Channel) WriteDiscrete(value bool) error {
	if c.channel < 1 || c.channel > int(c.s.Props.DO)*8 {
		return errors.New("invalid channel")
	}
	data := uint32(0)
	if value {
		data = 1
	}
	return c.s.writeDiscrete(byte(c.channel), data)
}

const ackTimeout = 200 * time.Millisecond

func awaitPin(pin machine.Pin, state bool, timeout time.Duration) bool {
	start := time.Now()
	for pin.Get() != state {
		time.Sleep(100 * time.Microsecond)
		if time.Since(start) > timeout {
			return false
		}
	}
	return true
	// TODO: Use channels when https://github.com/tinygo-org/tinygo/pull/1402 is merged.
	// edge := machine.PinRising
	// if state {
	// 	edge = machine.PinFalling
	// }
	// ch := make(chan struct{}, 1)
	// defer close(ch)
	// pin.SetInterrupt(edge, func(machine.Pin) {
	// 	ch <- struct{}{}
	// })
	// defer pin.SetInterrupt(0, nil)
	// select {
	// case <-ch:
	// 	return true
	// case <-time.After(timeout):
	// 	return false
	// }
}

var dataSyncErr = errors.New("base sync timeout")

func (p *P1AM) dataSync() error {
	if !awaitPin(p.slaveAckPin, true, ackTimeout) {
		return dataSyncErr
	}
	time.Sleep(time.Microsecond)
	if !awaitPin(p.slaveAckPin, false, ackTimeout) {
		return dataSyncErr
	}
	time.Sleep(time.Microsecond)
	if !awaitPin(p.slaveAckPin, true, ackTimeout) {
		return dataSyncErr
	}
	time.Sleep(time.Microsecond)
	return nil
}

func (p *P1AM) handleHDR(HDR byte) error {
	for !p.slaveAckPin.Get() {
	}
	if _, err := p.spiSendRecvByte(HDR); err != nil {
		return err
	}
	return p.spiTimeout(MAX_TIMEOUT*time.Millisecond, HDR, 2*time.Second)
}

func (p *P1AM) Read(data []byte) (int, error) {
	return len(data), p.spiSendRecvBuf(nil, data)
}

func (p *P1AM) Write(data []byte) (int, error) {
	return len(data), p.spiSendRecvBuf(data, nil)
}

func (p *P1AM) spiSendRecvBuf(w, r []byte) error {
	p.slaveSelectPin.Low()
	defer p.slaveSelectPin.High()
	return p.bus.Tx(w, r)
}

func (p *P1AM) spiSendRecvByte(data byte) (byte, error) {
	p.slaveSelectPin.Low()
	defer p.slaveSelectPin.High()
	return p.bus.Transfer(data)
}

func (p *P1AM) waitAck(timeout time.Duration) error {
	return p.spiTimeout(timeout, 0, 0)
}

var timeoutErr = errors.New("timeout")

func (p *P1AM) spiTimeout(timeout time.Duration, resendMsg byte, retryPeriod time.Duration) error {
	end := time.Now().Add(timeout)
	retry := time.Now().Add(retryPeriod)
	for time.Now().Before(end) {
		if p.slaveAckPin.Get() {
			time.Sleep(50 * time.Microsecond)
			return nil
		}
		if retryPeriod > 0 && time.Now().After(retry) {
			p.spiSendRecvByte(resendMsg)
			retry = retry.Add(retryPeriod)
		}
	}
	return timeoutErr
}

func (p *P1AM) SetEnabled(enabled bool) {
	p.baseEnablePin.Set(enabled)
}
