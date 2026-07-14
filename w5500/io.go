package w5500

import "time"

func (d *Device) irqPoll(sockn uint8, state uint8, deadline time.Time) uint8 {
	waitTime := 500 * time.Microsecond
	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			// If a deadline is set and it has passed, return 0.
			return sockIntUnknown
		}

		irq := d.readByte(sockInt, sockAddr(sockn)) & 0b00011111
		if got := irq & state; got != 0 {
			// Acknowledge the interrupt.
			d.writeByte(sockInt, sockAddr(sockn), got)

			return got
		}

		d.mu.Unlock()

		time.Sleep(waitTime)

		// Exponential backoff for polling.
		waitTime *= 2
		if waitTime > 10*time.Millisecond {
			waitTime = 10 * time.Millisecond
		}

		d.mu.Lock()
	}
}

func (d *Device) read(addr uint16, bsb uint8, p []byte) {
	d.cs(false)
	if len(p) == 0 {
		return
	}

	d.sendReadHeader(addr, bsb)
	_ = d.bus.Tx(nil, p)
	d.cs(true)
}

func (d *Device) readUint16(addr uint16, bsb uint8) uint16 {
	d.cs(false)
	d.sendReadHeader(addr, bsb)
	buf := d.cmdBuf
	_ = d.bus.Tx(nil, buf[:2])
	d.cs(true)
	return uint16(buf[1]) | uint16(buf[0])<<8
}

func (d *Device) readByte(addr uint16, bsb uint8) byte {
	d.cs(false)
	d.sendReadHeader(addr, bsb)
	r, _ := d.bus.Transfer(byte(0))
	d.cs(true)
	return r
}

func (d *Device) write(addr uint16, bsb uint8, p []byte) {
	d.cs(false)
	if len(p) == 0 {
		return
	}
	d.sendWriteHeader(addr, bsb)
	_ = d.bus.Tx(p, nil)
	d.cs(true)
}

func (d *Device) writeUint16(addr uint16, bsb uint8, v uint16) {
	d.cs(false)
	d.sendWriteHeader(addr, bsb)
	buf := d.cmdBuf
	buf[0] = byte(v >> 8)
	buf[1] = byte(v & 0xff)
	_ = d.bus.Tx(buf[:2], nil)
	d.cs(true)
}

func (d *Device) writeByte(addr uint16, bsb uint8, b byte) {
	d.cs(false)
	d.sendWriteHeader(addr, bsb)
	_, _ = d.bus.Transfer(b)
	d.cs(true)
}

func (d *Device) sendReadHeader(addr uint16, bsb uint8) {
	buf := d.cmdBuf
	buf[0] = byte(addr >> 8)
	buf[1] = byte(addr & 0xff)
	buf[2] = bsb << 3
	_ = d.bus.Tx(buf[:], nil)
}

func (d *Device) sendWriteHeader(addr uint16, bsb uint8) {
	buf := d.cmdBuf
	buf[0] = byte(addr >> 8)
	buf[1] = byte(addr & 0xff)
	buf[2] = bsb<<3 | 0b100
	_ = d.bus.Tx(buf[:], nil)
}
