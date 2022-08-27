package gc9a01

// Registers
const (
	NOP        = 0x00
	SWRESET    = 0x01
	RDDIDIF    = 0x04
	RDDST      = 0x09
	SLPIN      = 0x10
	SLPOUT     = 0x11
	PTLON      = 0x12
	NORON      = 0x13
	INVOFF     = 0x20
	INVON      = 0x21
	DISPOFF    = 0x28
	DISPON     = 0x29
	CASET      = 0x2A
	RASET      = 0x2B
	RAMWR      = 0x2C
	PTLAR      = 0x30
	VSCRDEF    = 0x33
	TEOFF      = 0x34
	TEON       = 0x35
	MADCTR     = 0x36
	VSCRSADD   = 0x37
	IDMOFF     = 0x38
	IDMON      = 0x39
	COLMOD     = 0x3A
	RMEMCON    = 0x3C
	STTRSCL    = 0x44
	GTSCL      = 0x45
	WRDISBV    = 0x51
	WRCTRLD    = 0x51
	MADCTL_MY  = 0x80
	MADCTL_MX  = 0x40
	MADCTL_MV  = 0x20
	MADCTL_ML  = 0x10
	MADCTL_RGB = 0x00
	MADCTL_BGR = 0x08
	MADCTL_MH  = 0x04
	RDID1      = 0xDA
	RDID2      = 0xDB
	RDID3      = 0xDC
	RGBICTR    = 0xB0
	BLPCHCTRL  = 0xB5
	DISFNCTL   = 0xB6
	TECTL      = 0xBA
	INTCTL     = 0xBA
	FRMCTL     = 0xE8
	SPICTL     = 0xE9
	PWCTR1     = 0xC1
	PWCTR2     = 0xC2
	PWCTR3     = 0xC3
	PWCTR4     = 0xC4
	PWCTR7     = 0xC7
	INTEN1     = 0xFE
	INTEN2     = 0xEF
	GMSET1     = 0xF0
	GMSET2     = 0xF1
	GMSET3     = 0xF2
	GMSET4     = 0xF3

	HORIZONTAL Orientation = 0
	VERTICAL   Orientation = 1
)
