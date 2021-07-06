package enc28j60

// ENC28J60 Control Registers
// Control register definitions are a combination of address,
// bank number, and Ethernet/MAC/PHY indicator bits.

// - Register address        (bits 0-4)
// - Bank number        (bits 5-6)
// - MAC/PHY indicator        (bit 7)
const (
	ADDR_MASK = 0x1F
	BANK_MASK = 0x60
	SPRD_MASK = 0x80
)

// Bank masks
const (
	bank0 = 0 << 5
	bank1 = 1 << 5
	bank2 = 2 << 5
	bank3 = 3 << 5
)

// All-bank registers
const (
	EIE   = 0x1B
	EIR   = 0x1C
	ESTAT = 0x1D
	ECON2 = 0x1E
	ECON1 = 0x1F
)

// Bank 0 registers. ADDR = (uint8 addr | bank mask)
const (
	ERDPTL = (0x00 | bank0)
	ERDPTH = (0x01 | bank0)
	EWRPTL = (0x02 | bank0)
	EWRPTH = (0x03 | bank0)
	ETXSTL = (0x04 | bank0)
	ETXSTH = (0x05 | bank0)
	ETXNDL = (0x06 | bank0)
	ETXNDH = (0x07 | bank0)
	ERXSTL = (0x08 | bank0)
	ERXSTH = (0x09 | bank0)
	ERXNDL = (0x0A | bank0)
	ERXNDH = (0x0B | bank0)
	// Bank0. The ERXRDPT registers define a location within the
	// FIFO where the receive hardware is forbidden to write
	// to. In normal operation, the receive hardware will write
	// data up to, but not including, the memory pointed to by
	// ERXRDPT.
	ERXRDPTL, ERXRDPTH = (0x0C | bank0), (0x0D | bank0)
	ERXWRPTL           = (0x0E | bank0)
	ERXWRPTH           = (0x0F | bank0)
	EDMASTL            = (0x10 | bank0)
	EDMASTH            = (0x11 | bank0)
	EDMANDL            = (0x12 | bank0)
	EDMANDH            = (0x13 | bank0)
	EDMADSTL           = (0x14 | bank0)
	EDMADSTH           = (0x15 | bank0)
	EDMACSL            = (0x16 | bank0)
	EDMACSH            = (0x17 | bank0)
)

// Bank 1 registers ADDR = (uint8 addr | bank mask)
const (
	EHT0    = (0x00 | bank1)
	EHT1    = (0x01 | bank1)
	EHT2    = (0x02 | bank1)
	EHT3    = (0x03 | bank1)
	EHT4    = (0x04 | bank1)
	EHT5    = (0x05 | bank1)
	EHT6    = (0x06 | bank1)
	EHT7    = (0x07 | bank1)
	EPMM0   = (0x08 | bank1)
	EPMM1   = (0x09 | bank1)
	EPMM2   = (0x0A | bank1)
	EPMM3   = (0x0B | bank1)
	EPMM4   = (0x0C | bank1)
	EPMM5   = (0x0D | bank1)
	EPMM6   = (0x0E | bank1)
	EPMM7   = (0x0F | bank1)
	EPMCSL  = (0x10 | bank1)
	EPMCSH  = (0x11 | bank1)
	EPMOL   = (0x14 | bank1)
	EPMOH   = (0x15 | bank1)
	EWOLIE  = (0x16 | bank1)
	EWOLIR  = (0x17 | bank1)
	ERXFCON = (0x18 | bank1)
	EPKTCNT = (0x19 | bank1)
)

// Bank 2 registers ADDR = (uint8 addr | bank mask | SPRD mask)
const (
	MACON1 = (0x00 | bank2 | 0x80)
	MACON2 = (0x01 | bank2 | 0x80)
	MACON3 = (0x02 | bank2 | 0x80)
	MACON4 = (0x03 | bank2 | 0x80)
	// When FULDPX (MACON3<0>) = 0 : Nibble  time  offset  delay  between  the  end  of  one  transmission  and  the  beginning  of  the  next  in  aback-to-back sequence. The register value should be programmed to the desired period in nibble timesminus 6. The recommended setting is 12h which represents the minimum IEEE specified Inter-PacketGap (IPG) of 9.6us
	MABBIPG  = (0x04 | bank2 | 0x80)
	MAIPGL   = (0x06 | bank2 | 0x80)
	MAIPGH   = (0x07 | bank2 | 0x80)
	MACLCON1 = (0x08 | bank2 | 0x80)
	MACLCON2 = (0x09 | bank2 | 0x80)
	MAMXFLL  = (0x0A | bank2 | 0x80)
	MAMXFLH  = (0x0B | bank2 | 0x80)
	MAPHSUP  = (0x0D | bank2 | 0x80)
	MICON    = (0x11 | bank2 | 0x80)
	MICMD    = (0x12 | bank2 | 0x80)
	MIREGADR = (0x14 | bank2 | 0x80)
	MIWRL    = (0x16 | bank2 | 0x80)
	MIWRH    = (0x17 | bank2 | 0x80)
	MIRDL    = (0x18 | bank2 | 0x80)
	MIRDH    = (0x19 | bank2 | 0x80)
)

// Bank 3 registers
const (
	MAADR1  = (0x00 | bank3 | 0x80)
	MAADR0  = (0x01 | bank3 | 0x80)
	MAADR3  = (0x02 | bank3 | 0x80)
	MAADR2  = (0x03 | bank3 | 0x80)
	MAADR5  = (0x04 | bank3 | 0x80)
	MAADR4  = (0x05 | bank3 | 0x80)
	EBSTSD  = (0x06 | bank3)
	EBSTCON = (0x07 | bank3)
	EBSTCSL = (0x08 | bank3)
	EBSTCSH = (0x09 | bank3)
	MISTAT  = (0x0A | bank3 | 0x80)
	EREVID  = (0x12 | bank3)
	ECOCON  = (0x15 | bank3)
	EFLOCON = (0x17 | bank3)
	EPAUSL  = (0x18 | bank3)
	EPAUSH  = (0x19 | bank3)
)

// PHY registers
const (
	PHCON1  = 0x00
	PHSTAT1 = 0x01
	PHHID1  = 0x02
	PHHID2  = 0x03
	PHCON2  = 0x10
	PHSTAT2 = 0x11
	PHIE    = 0x12
	PHIR    = 0x13
	PHLCON  = 0x14
)

// ENC28J60 ERXFCON Register Bit Definitions
const (
	ERXFCON_UCEN  = 0x80
	ERXFCON_ANDOR = 0x40
	ERXFCON_CRCEN = 0x20
	ERXFCON_PMEN  = 0x10
	ERXFCON_MPEN  = 0x08
	ERXFCON_HTEN  = 0x04
	ERXFCON_MCEN  = 0x02
	ERXFCON_BCEN  = 0x01
)

// ENC28J60 EIE Register Bit Definitions
const (
	EIE_INTIE  = 0x80
	EIE_PKTIE  = 0x40
	EIE_DMAIE  = 0x20
	EIE_LINKIE = 0x10
	EIE_TXIE   = 0x08
	EIE_WOLIE  = 0x04
	EIE_TXERIE = 0x02
	EIE_RXERIE = 0x01
)

// ENC28J60 EIR Register Bit Definitions
const (
	EIR_PKTIF  = 0x40
	EIR_DMAIF  = 0x20
	EIR_LINKIF = 0x10
	EIR_TXIF   = 0x08
	EIR_WOLIF  = 0x04
	EIR_TXERIF = 0x02
	EIR_RXERIF = 0x01
)

// ENC28J60 ESTAT Register Bit Definitions
const (
	ESTAT_INT     = 0x80
	ESTAT_LATECOL = 0x10
	ESTAT_RXBUSY  = 0x04
	ESTAT_TXABRT  = 0x02
	ESTAT_CLKRDY  = 0x01
)

// ENC28J60 ECON2 Register Bit Definitions
const (
	ECON2_AUTOINC = 0x80
	ECON2_PKTDEC  = 0x40
	ECON2_PWRSV   = 0x20
	ECON2_VRPS    = 0x08
)

// ENC28J60 ECON1 Register Bit Definitions
const (
	ECON1_TXRST  = 0x80
	ECON1_RXRST  = 0x40
	ECON1_DMAST  = 0x20
	ECON1_CSUMEN = 0x10
	ECON1_TXRTS  = 0x08
	ECON1_RXEN   = 0x04
	ECON1_BSEL1  = 0x02
	ECON1_BSEL0  = 0x01
)

// ENC28J60 MACON1 Register Bit Definitions
const (
	MACON1_LOOPBK  = 0x10
	MACON1_TXPAUS  = 0x08
	MACON1_RXPAUS  = 0x04
	MACON1_PASSALL = 0x02
	MACON1_MARXEN  = 0x01
)

// ENC28J60 MACON2 Register Bit Definitions
const (
	MACON2_MARST   = 0x80
	MACON2_RNDRST  = 0x40
	MACON2_MARXRST = 0x08
	MACON2_RFUNRST = 0x04
	MACON2_MATXRST = 0x02
	MACON2_TFUNRST = 0x01
)

// ENC28J60 MACON3 Register Bit Definitions
const (
	// All short frames will be zero-padded to 64 bytes and a valid CRC will then be appended
	MACON3_ZPADCRC = 0b111 << 5
	MACON3_PADCFG2 = 0x80
	MACON3_PADCFG1 = 0x40
	MACON3_PADCFG0 = 0x20
	// MAC will append a  valid CRC to all frames transmitted regardless of  PADCFG bits. TXCRCEN must be set if the PADCFG bits specify that a valid CRC will be appended.
	MACON3_TXCRCEN = 0x10
	MACON3_PHDRLEN = 0x08
	MACON3_HFRMLEN = 0x04
	//The type/length field of transmitted and received frames will be checked. If it represents a length, the frame size will be compared and mismatches will be reported in the transmit/receive status vector.
	MACON3_FRMLNEN = 0x02
	// MAC will operate in Full-Duplex mode. PDPXMD bit must also be set.
	MACON3_FULDPX = 0x01
)

// ENC28J60 MACON4 Register Bit Definitions
const (
	//When the medium is occupied, the MAC will wait indefinitely for it to become free when attempting to transmit (use this setting for IEEE 802.3™ compliance)
	MACON4_DEFER = 1 << 6
	//  After  incidentally  causing  a  collision  during  backpressure,  the  MAC  will  immediately  begin retransmitting
	MACON4_BPEN = 1 << 5
	//After any collision, the MAC will immediately begin retransmitting
	MACON4_NOBKOFF = 1 << 4
)

// ENC28J60 MICMD Register Bit Definitions
const (
	MICMD_MIISCAN = 0x02
	MICMD_MIIRD   = 0x01
)

// ENC28J60 MISTAT Register Bit Definitions
const (
	MISTAT_NVALID = 0x04
	MISTAT_SCAN   = 0x02
	MISTAT_BUSY   = 0x01
)

// ENC28J60 PHY PHCON1 Register Bit Definitions
const (
	PHCON1_PRST    = 0x8000
	PHCON1_PLOOPBK = 0x4000
	PHCON1_PPWRSV  = 0x0800
	PHCON1_PDPXMD  = 0x0100
)

// ENC28J60 PHY PHSTAT1 Register Bit Definitions
const (
	PHSTAT1_PFDPX  = 0x1000
	PHSTAT1_PHDPX  = 0x0800
	PHSTAT1_LLSTAT = 0x0004
	PHSTAT1_JBSTAT = 0x0002
)

// ENC28J60 PHY PHCON2 Register Bit Definitions
const (
	PHCON2_FRCLINK = 0x4000
	PHCON2_TXDIS   = 0x2000
	PHCON2_JABBER  = 0x0400
	PHCON2_HDLDIS  = 0x0100
)

// ENC28J60 Packet Control Byte Bit Definitions
const (
	PKTCTRL_PHUGEEN   = 0x08
	PKTCTRL_PPADEN    = 0x04
	PKTCTRL_PCRCEN    = 0x02
	PKTCTRL_POVERRIDE = 0x01
)

// SPI operation codes
const (
	READ_CTL_REG  = 0x00
	READ_BUF_MEM  = 0x3A
	WRITE_CTL_REG = 0x40
	WRITE_BUF_MEM = 0x7A
	BIT_FIELD_SET = 0x80
	BIT_FIELD_CLR = 0xA0
	SOFT_RESET    = 0xFF
)

// The RXSTART_INIT should be zero. See Rev. B4 Silicon Errata
// buffer boundaries applied to internal 8K ram
// the entire available packet buffer space is allocated
//
// start with recbuf at 0/
const RXSTART_INIT = 0x0

// receive buffer end
const RXSTOP_INIT = (0x1FFF - 0x0600 - 1)

// start TX buffer at 0x1FFF-0x0600, pace for one full ethernet frame (1536 bytes)
const TXSTART_INIT = (0x1FFF - 0x0600)

// stp TX buffer at end of mem leaving space for status vector
const TXSTOP_INIT = 0x1FFF

//
// max frame length which the conroller will accept:
const MAX_FRAMELEN = 1500 // (note: maximum ethernet frame length would be 1518)
// MAX_FRAMELEN     600
