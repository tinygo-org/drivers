// Package mcp2515 implements a driver for the MCP2515 CAN Controller.
//
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/MCP2515-Stand-Alone-CAN-Controller-with-SPI-20001801J.pdf
package mcp2515 // import "tinygo.org/x/drivers/mcp2515"

const DebugEn = 0

const (
	// begin mt

	timeoutvalue = 50
	mcpSidh      = 0
	mcpSidl      = 1
	mcpEid8      = 2
	mcpEid0      = 3

	mcpTxbExideM = 0x08 // in txbnsidl
	mcpDlcMask   = 0x0f //= 4 lsbits
	mcpRtrMask   = 0x40 // =(1<=<6) bit= 6

	mcpRxbRxAny    = 0x60
	mcpRxbRxExt    = 0x40
	mcpRxbRxStd    = 0x20
	mcpRxbRxStdExt = 0x00
	mcpRxbRxMask   = 0x60
	mcpRxbBuktMask = 1 << 2

	// bits in the txbnctrl registers.

	mcpTxbTxbufeM = 0x80
	mcpTxbAbtfM   = 0x40
	mcpTxbMloaM   = 0x20
	mcpTxbTxerrM  = 0x10
	mcpTxbTxreqM  = 0x08
	mcpTxbTxieM   = 0x04
	mcpTxbTxp10M  = 0x03

	mcpTxbRtrM = 0x40 // in txbndlc
	mcpRxbIdeM = 0x08 // in rxbnsidl
	mcpRxbRtrM = 0x40 // in rxbndlc

	mcpStatTxPendingMask = 0x54
	mcpStatTx0Pending    = 0x04
	mcpStatTx1Pending    = 0x10
	mcpStatTx2Pending    = 0x40
	mcpStatTxifMask      = 0xa8
	mcpStatTx0if         = 0x08
	mcpStatTx1if         = 0x20
	mcpStatTx2if         = 0x80
	mcpStatRxifMask      = 0x03
	mcpStatRx0if         = 1 << 0
	mcpStatRx1if         = 1 << 1

	mcpEflgRx1ovr    = 1 << 7
	mcpEflgRx0ovr    = 1 << 6
	mcpEflgTxbo      = 1 << 5
	mcpEflgTxep      = 1 << 4
	mcpEflgRxep      = 1 << 3
	mcpEflgTxwar     = 1 << 2
	mcpEflgRxwar     = 1 << 1
	mcpEflgEwarn     = 1 << 0
	mcpEflgErrormask = 0xf8 //= 5 ms-bits

	// define mcp2515 register addresses

	mcpRXF0SIDH  = 0x00
	mcpRXF0SIDL  = 0x01
	mcpRXF0EID8  = 0x02
	mcpRXF0EID0  = 0x03
	mcpRXF1SIDH  = 0x04
	mcpRXF1SIDL  = 0x05
	mcpRXF1EID8  = 0x06
	mcpRXF1EID0  = 0x07
	mcpRXF2SIDH  = 0x08
	mcpRXF2SIDL  = 0x09
	mcpRXF2EID8  = 0x0a
	mcpRXF2EID0  = 0x0b
	mcpBFPCTRL   = 0x0c
	mcpTXRTSCTRl = 0x0d
	mcpCANSTAT   = 0x0e
	mcpCANCTRL   = 0x0f
	mcpRXF3SIDH  = 0x10
	mcpRXF3SIDL  = 0x11
	mcpRXF3EID8  = 0x12
	mcpRXF3EID0  = 0x13
	mcpRXF4SIDH  = 0x14
	mcpRXF4SIDL  = 0x15
	mcpRXF4EID8  = 0x16
	mcpRXF4EID0  = 0x17
	mcpRXF5SIDH  = 0x18
	mcpRXF5SIDL  = 0x19
	mcpRXF5EID8  = 0x1a
	mcpRXF5EID0  = 0x1b
	mcpTEC       = 0x1c
	mcpREC       = 0x1d
	mcpRXM0SIDH  = 0x20
	mcpRXM0SIDL  = 0x21
	mcpRXM0EID8  = 0x22
	mcpRXM0EID0  = 0x23
	mcpRXM1SIDH  = 0x24
	mcpRXM1SIDL  = 0x25
	mcpRXM1EID8  = 0x26
	mcpRXM1EID0  = 0x27
	mcpCNF3      = 0x28
	mcpCNF2      = 0x29
	mcpCNF1      = 0x2a
	mcpCANINTE   = 0x2b
	mcpCANINTF   = 0x2c
	mcpEFLG      = 0x2d
	mcpTXB0CTRL  = 0x30
	mcpTXB0SIDH  = 0x31
	mcpTXB1CTRL  = 0x40
	mcpTXB1SIDH  = 0x41
	mcpTXB2CTRL  = 0x50
	mcpTXB2SIDH  = 0x51
	mcpRXB0CTRL  = 0x60
	mcpRXB0SIDH  = 0x61
	mcpRXB1CTRL  = 0x70
	mcpRXB1SIDH  = 0x71

	mcpTxInt   = 0x1c // enable all transmit interrup ts
	mcpTx01Int = 0x0c // enable txb0 and txb1 interru pts
	mcpRxInt   = 0x03 // enable receive interrupts
	mcpNoInt   = 0x00 // disable all interrupts

	mcpTx01Mask = 0x14
	mcpTxMask   = 0x54

	// define spi instruction set
	mcpWrite   = 0x02
	mcpRead    = 0x03
	mcpBitMod  = 0x05
	mcpLoadTx0 = 0x40
	mcpLoadTx1 = 0x42
	mcpLoadTx2 = 0x44

	mcpRtsTx0     = 0x81
	mcpRtsTx1     = 0x82
	mcpRtsTx2     = 0x84
	mcpRtsAll     = 0x87
	mcpReadRx0    = 0x90
	mcpReadRx1    = 0x94
	mcpReadStatus = 0xa0
	mcpRxStatus   = 0xb0
	mcpReset      = 0xc0

	// canctrl register values

	modeNormal     = 0x00
	modeSleep      = 0x20
	modeLoopBack   = 0x40
	modeListenOnly = 0x60
	modeConfig     = 0x80
	modePowerUp    = 0xe0
	modeMask       = 0xe0
	abortTx        = 0x10
	modeOneShot    = 0x08
	clkoutEnable   = 0x04
	clkoutDisable  = 0x00
	clkoutPs1      = 0x00
	clkoutPs2      = 0x01
	clkoutPs4      = 0x02
	clkoutPs8      = 0x03

	// cnf1 register values

	sjw1 = 0x00
	sjw2 = 0x40
	sjw3 = 0x80
	sjw4 = 0xc0

	//  cnf2 register values

	btlmode  = 0x80
	sample1x = 0x00
	sample3x = 0x40

	// cnf3 register values

	sofEnable     = 0x80
	sofDisable    = 0x00
	wakfilEnable  = 0x40
	wakfilDisable = 0x00

	// canintf register bits

	mcpRX0IF = 0x01
	mcpRX1IF = 0x02
	mcpTX0IF = 0x04
	mcpTX1IF = 0x08
	mcpTX2IF = 0x10
	mcpERRIF = 0x20
	mcpWAKIF = 0x40
	mcpMERRF = 0x80

	// bfpctrl register bits

	b1bfs = 0x20
	b0bfs = 0x10
	b1bfe = 0x08
	b0bfe = 0x04
	b1bfm = 0x02
	b0bfm = 0x01

	// txrtctrl register bits

	b2rts  = 0x20
	b1rts  = 0x10
	b0rts  = 0x08
	b2rtsm = 0x04
	b1rtsm = 0x02
	b0rtsm = 0x01

	// clock

	Clock16MHz = 1
	Clock8MHz  = 2

	// speed= 16m

	mcp16mHz1000kBpsCfg1 = 0x00
	mcp16mHz1000kBpsCfg2 = 0xd0
	mcp16mHz1000kBpsCfg3 = 0x82

	mcp16mHz500kBpsCfg1 = 0x00
	mcp16mHz500kBpsCfg2 = 0xf0
	mcp16mHz500kBpsCfg3 = 0x86

	mcp16mHz250kBpsCfg1 = 0x41
	mcp16mHz250kBpsCfg2 = 0xf1
	mcp16mHz250kBpsCfg3 = 0x85

	mcp16mHz200kBpsCfg1 = 0x01
	mcp16mHz200kBpsCfg2 = 0xfa
	mcp16mHz200kBpsCfg3 = 0x87

	mcp16mHz125kBpsCfg1 = 0x03
	mcp16mHz125kBpsCfg2 = 0xf0
	mcp16mHz125kBpsCfg3 = 0x86

	mcp16mHz100kBpsCfg1 = 0x03
	mcp16mHz100kBpsCfg2 = 0xfa
	mcp16mHz100kBpsCfg3 = 0x87

	mcp16mHz95kBpsCfg1 = 0x03
	mcp16mHz95kBpsCfg2 = 0xad
	mcp16mHz95kBpsCfg3 = 0x07

	mcp16mHz83k3BpsCfg1 = 0x03
	mcp16mHz83k3BpsCfg2 = 0xbe
	mcp16mHz83k3BpsCfg3 = 0x07

	mcp16mHz80kBpsCfg1 = 0x03
	mcp16mHz80kBpsCfg2 = 0xff
	mcp16mHz80kBpsCfg3 = 0x87

	mcp16mHz50kBpsCfg1 = 0x07
	mcp16mHz50kBpsCfg2 = 0xfa
	mcp16mHz50kBpsCfg3 = 0x87

	mcp16mHz40kBpsCfg1 = 0x07
	mcp16mHz40kBpsCfg2 = 0xff
	mcp16mHz40kBpsCfg3 = 0x87

	mcp16mHz33kBpsCfg1 = 0x09
	mcp16mHz33kBpsCfg2 = 0xbe
	mcp16mHz33kBpsCfg3 = 0x07

	mcp16mHz31k25BpsCfg1 = 0x0f
	mcp16mHz31k25BpsCfg2 = 0xf1
	mcp16mHz31k25BpsCfg3 = 0x85

	mcp16mHz25kBpsCfg1 = 0x0f
	mcp16mHz25kBpsCfg2 = 0xba
	mcp16mHz25kBpsCfg3 = 0x07

	mcp16mHz20kBpsCfg1 = 0x0f
	mcp16mHz20kBpsCfg2 = 0xff
	mcp16mHz20kBpsCfg3 = 0x87

	mcp16mHz10kBpsCfg1 = 0x1f
	mcp16mHz10kBpsCfg2 = 0xff
	mcp16mHz10kBpsCfg3 = 0x87

	mcp16mHz5kBpsCfg1 = 0x3f
	mcp16mHz5kBpsCfg2 = 0xff
	mcp16mHz5kBpsCfg3 = 0x87

	mcp16mHz666kBpsCfg1 = 0x00
	mcp16mHz666kBpsCfg2 = 0xa0
	mcp16mHz666kBpsCfg3 = 0x04

	// speed= 8m

	mcp8mHz1000kBpsCfg1 = 0x00
	mcp8mHz1000kBpsCfg2 = 0x80
	mcp8mHz1000kBpsCfg3 = 0x00

	mcp8mHz500kBpsCfg1 = 0x00
	mcp8mHz500kBpsCfg2 = 0x90
	mcp8mHz500kBpsCfg3 = 0x02

	mcp8mHz250kBpsCfg1 = 0x00
	mcp8mHz250kBpsCfg2 = 0xb1
	mcp8mHz250kBpsCfg3 = 0x05

	mcp8mHz200kBpsCfg1 = 0x00
	mcp8mHz200kBpsCfg2 = 0xb4
	mcp8mHz200kBpsCfg3 = 0x06

	mcp8mHz125kBpsCfg1 = 0x01
	mcp8mHz125kBpsCfg2 = 0xb1
	mcp8mHz125kBpsCfg3 = 0x05

	mcp8mHz100kBpsCfg1 = 0x01
	mcp8mHz100kBpsCfg2 = 0xb4
	mcp8mHz100kBpsCfg3 = 0x06

	mcp8mHz80kBpsCfg1 = 0x01
	mcp8mHz80kBpsCfg2 = 0xbf
	mcp8mHz80kBpsCfg3 = 0x07

	mcp8mHz50kBpsCfg1 = 0x03
	mcp8mHz50kBpsCfg2 = 0xb4
	mcp8mHz50kBpsCfg3 = 0x06

	mcp8mHz40kBpsCfg1 = 0x03
	mcp8mHz40kBpsCfg2 = 0xbf
	mcp8mHz40kBpsCfg3 = 0x07

	mcp8mHz31k25BpsCfg1 = 0x07
	mcp8mHz31k25BpsCfg2 = 0xa4
	mcp8mHz31k25BpsCfg3 = 0x04

	mcp8mHz20kBpsCfg1 = 0x07
	mcp8mHz20kBpsCfg2 = 0xbf
	mcp8mHz20kBpsCfg3 = 0x07

	mcp8mHz10kBpsCfg1 = 0x0f
	mcp8mHz10kBpsCfg2 = 0xbf
	mcp8mHz10kBpsCfg3 = 0x07

	mcp8mHz5kBpsCfg1 = 0x1f
	mcp8mHz5kBpsCfg2 = 0xbf
	mcp8mHz5kBpsCfg3 = 0x07

	mcp16mHz47kBpsCfg1 = 0x06
	mcp16mHz47kBpsCfg2 = 0xbe
	mcp16mHz47kBpsCfg3 = 0x07

	mcpdebug      = 0
	mcpdebugTxbuf = 0
	mcpNTxbuffers = 3

	mcpRxbuf0 = 0x61
	mcpRxbuf1 = 0x71

	mcp2515Ok    = 0
	mcp2515Fail  = 1
	mcpAlltxbusy = 2

	candebug = 1

	canuseloop = 0

	cansendtimeout = 200 // milliseconds

	mcpPinHiz = 0
	mcpPinInt = 1
	mcpPinOut = 2
	mcpPinIn  = 3

	mcpRx0bf  = 0
	mcpRx1bf  = 1
	mcpTx0rts = 2
	mcpTx1rts = 3
	mcpTx2rts = 4

	// initial value of gcanautoprocess

	canautoprocess     = 1
	canautoon          = 1
	canautooff         = 0
	canStdid           = 0
	canExtid           = 1
	candefaultident    = 0x55cc
	candefaultidentext = 1

	CAN5kBps    = 1
	CAN10kBps   = 2
	CAN20kBps   = 3
	CAN25kBps   = 4
	CAN31k25Bps = 5
	CAN33kBps   = 6
	CAN40kBps   = 7
	CAN50kBps   = 8
	CAN80kBps   = 9
	CAN83k3Bps  = 10
	CAN95kBps   = 11
	CAN100kBps  = 12
	CAN125kBps  = 13
	CAN200kBps  = 14
	CAN250kBps  = 15
	CAN500kBps  = 16
	CAN666kBps  = 17
	CAN1000kBps = 18
	CAN47kBps   = 19

	canOk             = 0
	canFailinit       = 1
	canFailtx         = 2
	canMsgavail       = 3
	canNomsg          = 4
	canCtrlerror      = 5
	canGettxbftimeout = 6
	canSendmsgtimeout = 7
	canFail           = 0xff

	canMaxCharInMessage = 8
)
