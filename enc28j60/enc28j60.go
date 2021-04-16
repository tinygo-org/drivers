package enc28j60

import (
	"machine"
	"runtime/interrupt"

	"time"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
	"tinygo.org/x/drivers"
)

// ETHERCARD_STASH Enable access to IC memory
const ETHERCARD_STASH = false

// Device is the SPI interface to a ENC28J60
type Dev struct {
	// Chip select pin
	CSB           machine.Pin
	Bank          uint8
	NextPacketPtr uint16
	buffer        []byte
	// subnet mask
	mask net.IPMask
	// device IP address
	myip net.IP
	// myip but masked
	broadcastip net.IP
	// which IP is recieving requests or the router
	gatewayip net.IP
	dummy     [2]byte
	// mac address
	macaddr net.HardwareAddr
	// SPI bus (requires chip select to be usable).
	Bus drivers.SPI
	// *Stash
	// interrupt state
	is interrupt.State
}

// NewSPI returns a new device driver. The SPI is configured in this call
func New(csb machine.Pin, spi drivers.SPI) *Dev {
	return &Dev{
		CSB:  csb, // chip select
		Bus:  spi,
		Bank: 255, // bad bank so as to force bank set on first read
	}
}

// Init initializes device for use and configures the enc28j60's registries.
func (d *Dev) Init(buff []byte, macaddr []byte) error {
	if len(macaddr) != 6 {
		return errBadMac
	}
	if buff == nil || len(buff) < 64 || len(buff) > 1500 {
		return errBufferSize
	}
	d.buffer = buff
	if ETHERCARD_STASH {
		// d.Stash = &Stash{}
		// d.Stash.InitMap(SCRATCH_PAGE_NUM)
	}
	d.macaddr = macaddr
	dbp("cfg call w/mac:", macaddr)
	d.configure(macaddr)
	if d.GetRev() == 0 {
		return errBadRev
	}
	return nil
}

// read len(data) bytes from buffer
func (d *Dev) readBuffer(data []byte) {
	d.enableCS()
	cmd := [1]byte{READ_BUF_MEM}
	d.Bus.Tx(cmd[:], nil)
	d.Bus.Tx(nil, data)
	d.disableCS()
}

// write data to buffer
func (d *Dev) writeBuffer(data []byte) {
	d.enableCS()
	cmd := [1]byte{WRITE_BUF_MEM}
	d.Bus.Tx(append(cmd[:], data...), nil)
	d.disableCS()
}

func (d *Dev) Reset() {
	d.enableCS()
	d.Bank = 255
	d.Bus.Tx([]byte{SOFT_RESET}, nil)
	d.disableCS()
}

func (d *Dev) clkOut(clk uint8) {
	//setup clkout: 2 is 12.5MHz:
	d.write(ECOCON, clk&0x7)
}

// Init initializes communication and device.
//
// macaddr is of length 6.
func (d *Dev) configure(macaddr []byte) {
	d.Reset()
	time.Sleep(50 * time.Millisecond)

	// check CLKRDY bit to see if reset is complete
	// The CLKRDY does not work. See Rev. B4 Silicon Errata point. Just wait.
	// for d.readOp(READ_CTL_REG, ESTAT)&ESTAT_CLKRDY == 0 {
	// }

	// bank 0 stuff
	// initialize receive buffer
	// 16-bit transfers, must write low byte first
	// set receive buffer start address
	// NextPacketPtr = RXSTART_INIT
	// Rx start
	d.write(ERXSTL, RXSTART_INIT&0xFF)
	d.write(ERXSTH, RXSTART_INIT>>8)
	// set receive pointer address (should be programmed with same value, see 6.1)
	d.write(ERXRDPTL, RXSTART_INIT&0xFF)
	d.write(ERXRDPTH, RXSTART_INIT>>8)
	// RX end
	d.write(ERXNDL, RXSTOP_INIT&0xFF)
	d.write(ERXNDH, RXSTOP_INIT>>8)
	// TX start
	d.write(ETXSTL, TXSTART_INIT&0xFF)
	d.write(ETXSTH, TXSTART_INIT>>8)
	// TX end
	d.write(ETXNDL, TXSTOP_INIT&0xFF)
	d.write(ETXNDH, TXSTOP_INIT>>8)
	// do bank 1 stuff, packet filter:
	// For broadcast packets we allow only ARP packtets
	// All other packets should be unicast only for our mac (MAADR)
	//
	// The pattern to match on is therefore
	// Type     ETH.DST
	// ARP      BROADCAST
	// 06 08 -- ff ff ff ff ff ff -> ip checksum for theses bytes=f7f9
	// in binary these poitions are:11 0000 0011 1111
	// This is hex 303F->EPMM0=0x3f,EPMM1=0x30
	d.write(ERXFCON, ERXFCON_UCEN|ERXFCON_CRCEN|ERXFCON_PMEN)
	d.write(EPMM0, 0x3f)
	d.write(EPMM1, 0x30)
	d.write(EPMCSL, 0xf9)
	d.write(EPMCSH, 0xf7)
	//
	//
	// do bank 2 stuff
	// enable MAC receive frame (see 6.5 bullet 1)
	d.write(MACON1, MACON1_MARXEN|MACON1_TXPAUS|MACON1_RXPAUS)
	// bring MAC out of reset
	d.write(MACON2, 0x00)
	// enable automatic padding to 60bytes and CRC operations
	d.writeOp(BIT_FIELD_SET, MACON3, MACON3_PADCFG0|MACON3_TXCRCEN|MACON3_FRMLNEN)
	// set inter-frame gap (non-back-to-back)
	d.write(MAIPGL, 0x12)
	d.write(MAIPGH, 0x0C)
	// set inter-frame gap (back-to-back)
	d.write(MABBIPG, 0x12)
	// Set the maximum packet size which the controller will accept
	// Do not send packets longer than MAX_FRAMELEN:
	d.write(MAMXFLH, MAX_FRAMELEN>>8)
	// do bank 3 stuff
	// write MAC address
	// NOTE: MAC address in ENC28J60 is byte-backward
	d.write(MAADR5, macaddr[0])
	d.write(MAADR4, macaddr[1])
	d.write(MAADR3, macaddr[2])
	d.write(MAADR2, macaddr[3])
	d.write(MAADR1, macaddr[4])
	d.write(MAADR0, macaddr[5])
	// no loopback of transmitted frames
	d.phyWrite(PHCON2, PHCON2_HDLDIS)
	// switch to bank 0
	d.setBank(ECON1)
	// enable interrutps
	d.writeOp(BIT_FIELD_SET, EIE, EIE_INTIE|EIE_PKTIE)
	// enable packet reception
	d.writeOp(BIT_FIELD_SET, ECON1, ECON1_RXEN)
}

func (d *Dev) GetRev() uint8 {
	return d.read(EREVID)
}

func (d *Dev) PacketSend(len uint16, packet []byte) {
	d.write(EWRPTL, TXSTART_INIT&0xFF)
	d.write(EWRPTH, TXSTART_INIT>>8)
	// Set the TXND pointer to correspond to the packet size given
	d.write(ETXNDL, uint8(TXSTART_INIT+len&0xFF))
	d.write(ETXNDH, uint8((TXSTART_INIT+len)>>8))
	// write per-packet control byte (0x00 means use macon3 settings)
	d.writeOp(WRITE_BUF_MEM, 0, 0x00)
	// copy the packet into the transmit buffer
	d.writeBuffer(packet[:len])
	// send the contents of the transmit buffer onto the network
	d.writeOp(BIT_FIELD_SET, ECON1, ECON1_TXRTS)
	// Reset the transmit logic problem. See Rev. B4 Silicon Errata point 12.
	if d.read(EIR)&EIR_TXERIF != 0 {
		d.writeOp(BIT_FIELD_CLR, ECON1, ECON1_TXRTS)
	}
}

// return packet length in buffer
func (d *Dev) PacketRecieve(maxlen uint16, packet []byte) (len uint16) {
	var rxstat uint16
	if d.read(EPKTCNT) == 0 {
		return 0
	}

	// Set the read pointer to the start of the received packet
	d.write16(ERDPTL, d.NextPacketPtr)
	var fromBuff [2 + 2 + 2]byte
	d.readBuffer(fromBuff[:])
	d.NextPacketPtr = uint16(fromBuff[0]) + uint16(fromBuff[1])<<8
	// read the packet length (see datasheet page 43)
	len = uint16(fromBuff[2]) + uint16(fromBuff[3])<<8 - 4 //remove the CRC count (minus 4)

	rxstat = uint16(fromBuff[4]) + uint16(fromBuff[5])<<8
	// read the next packet pointer
	d.NextPacketPtr = uint16(d.readOp(READ_BUF_MEM, 0))
	d.NextPacketPtr |= uint16(d.readOp(READ_BUF_MEM, 0)) << 8
	// read the packet length (see datasheet page 43)
	len = uint16(d.readOp(READ_BUF_MEM, 0))
	len |= uint16(d.readOp(READ_BUF_MEM, 0)) << 8
	len -= 4 //remove the CRC count
	// read the receive status (see datasheet page 43)
	rxstat = uint16(d.readOp(READ_BUF_MEM, 0))
	rxstat |= uint16(d.readOp(READ_BUF_MEM, 0)) << 8
	// limit retrieve length
	if len > maxlen-1 {
		len = maxlen - 1
	}
	// check CRC and symbol errors (see datasheet page 44, table 7-3):
	// The ERXFCON.CRCEN is set by default. Normally we should not
	// need to check this.
	if (rxstat & 0x80) == 0 {
		// invalid
		len = 0
	} else {
		// copy the packet from the receive buffer
		d.readBuffer(packet[:len])
	}
	// Move the RX read pointer to the start of the next received packet
	// This frees the memory we just read out
	d.write(ERXRDPTL, uint8(d.NextPacketPtr))
	d.write(ERXRDPTH, uint8(d.NextPacketPtr>>8))
	// decrement the packet counter indicate we are done with this packet
	d.writeOp(BIT_FIELD_SET, ECON2, ECON2_PKTDEC)
	return len
}
