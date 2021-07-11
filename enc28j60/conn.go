package enc28j60

import (
	"io"
	"time"

	swtch "github.com/soypat/ether-swtch"
)

type Packet struct {
	ic     *Dev
	cursor uint16
	end    uint16
}

// NextPacket returns a Packet which reads data from the
// next packet in the FIFO queue. When one is done reading packet data, call
// Discard on said packet and NextPacket will return the next packet in the FIFO.
func (d *Dev) NextPacket(deadline time.Time) (swtch.Reader, error) {
	dbp("NextPacket")
	var err error
	for d.read(EPKTCNT) == 0 { // loop until a packet is received.
		if time.Since(deadline) > 0 {
			return nil, ErrRXDeadlineExceeded
		}
		time.Sleep(100 * time.Millisecond)
	}
	// p := &Packet{ic: d} // Weird bug when creating Packet before read loop. LLVM on AVR is buggy and ic will be nil afterwards.
	p := &d.rx
	// Set the read pointer to the start of the next packet
	d.write16(ERDPTL, d.nextPacketPtr)
	p.cursor = d.nextPacketPtr // Packet reader

	d.readBuffer(d.buf[:])
	d.nextPacketPtr = uint16(d.buf[0]) + uint16(d.buf[1])<<8
	// read the packet length (see datasheet page 43)
	plen := uint16(d.buf[2]) + uint16(d.buf[3])<<8 - 4 //remove the CRC count (minus 4)
	p.end = p.cursor + plen
	// read the receive status (see datasheet page 43)
	rxstat := uint16(d.buf[4]) + uint16(d.buf[5])<<8
	// check CRC and symbol errors (see datasheet page 44, table 7-3):
	// The ERXFCON.CRCEN is set by default. Normally we should not
	// need to check this.
	if (rxstat & 0x80) == 0 {
		err = ErrCRC
	}
	return p, err
}

// Discard drops the remaining packet data to be read. Any subsequent call
// to Read will return io.EOF error. This implements ether-swtch's Reader interface.
func (p *Packet) Discard() error {
	dbp("DiscardPacket")
	if p.cursor != p.end {
		p.cursor = p.end
		p.ic.writeOp(BIT_FIELD_SET, ECON2, ECON2_PKTDEC)
	}
	return nil
}

// Read reads packet data into buffer returning the amound
// of data read. io.EOF is returned when done with the packet.
func (p *Packet) Read(buff []byte) (n uint16, err error) {
	dbp("ReadPacket")
	// total remaining packet length
	plen := p.end - p.cursor
	if plen == 0 {
		return 0, io.EOF
	}
	if len(buff) == 0 {
		return 0, nil
	}
	// Limit retreive length if Total packet length is greater than buffer length
	if plen > uint16(len(buff)) {
		plen = uint16(len(buff))
	}
	// copy the packet from the receive buffer
	p.ic.readBuffer(buff[:plen])
	p.cursor += plen
	// Move the RX read pointer to where we ended reading
	p.ic.write16(ERXRDPTL, p.cursor)
	if p.cursor == p.end { // minus CRC length
		// decrement packet counter to indicate we are done with it.
		p.ic.writeOp(BIT_FIELD_SET, ECON2, ECON2_PKTDEC)
		err = io.EOF
	}
	return plen, err
}

// Write writes data into ENC28J60's TX buffer. Data must not exceed the buffer bounds or
// ErrBufferSize will be returned. Use Flush method to send data over network once
// finished writing.
func (d *Dev) Write(buff []byte) (uint16, error) {
	plen := uint16(len(buff))
	if plen+d.tcursor > MAX_FRAMELEN {
		d.tcursor = 0
		return 0, ErrBufferSize
	}

	d.write16(EWRPTL, TXSTART_INIT+d.tcursor)
	if d.tcursor == 0 {
		// write per-packet control byte (0x00 means use macon3 settings)
		d.writeOp(WRITE_BUF_MEM, 0, 0x00)
		d.tcursor++ // WBM spurious increment
	}
	// copy the packet into the transmit buffer
	d.writeBuffer(buff)
	d.tcursor += plen
	return plen, nil
}

// Flush sends the data written to the TX buffer with Write over the network
// as a packet. Aftyer flush ENC28J60 is ready to start writing a new packet.
func (d *Dev) Flush() error {
	dbp("send response")
	d.write16(ETXNDL, TXSTART_INIT+d.tcursor-1) // subtract WBM spurious increment
	// send the contents of the transmit buffer onto the network
	d.writeOp(BIT_FIELD_SET, ECON1, ECON1_TXRTS)
	// Reset the transmit logic problem. See Rev. B4 Silicon Errata point 12.
	if d.read(EIR)&EIR_TXERIF != 0 {
		d.writeOp(BIT_FIELD_CLR, ECON1, ECON1_TXRTS)
	}
	d.tcursor = 0
	return nil
}
