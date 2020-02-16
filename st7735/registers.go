package st7735

// Registers
const (
	NOP        = 0x00
	SWRESET    = 0x01
	RDDID      = 0x04
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
	RAMRD      = 0x2E
	PTLAR      = 0x30
	COLMOD     = 0x3A
	MADCTL     = 0x36
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
	RDID4      = 0xDD
	FRMCTR1    = 0xB1
	FRMCTR2    = 0xB2
	FRMCTR3    = 0xB3
	INVCTR     = 0xB4
	DISSET5    = 0xB6
	PWCTR1     = 0xC0
	PWCTR2     = 0xC1
	PWCTR3     = 0xC2
	PWCTR4     = 0xC3
	PWCTR5     = 0xC4
	VMCTR1     = 0xC5
	PWCTR6     = 0xFC
	GMCTRP1    = 0xE0
	GMCTRN1    = 0xE1
	VSCRDEF    = 0x33
	VSCRSADD   = 0x37

	GREENTAB   Model = 0
	MINI80x160 Model = 1

	NO_ROTATION  Rotation = 0
	ROTATION_90  Rotation = 1 // 90 degrees clock-wise rotation
	ROTATION_180 Rotation = 2
	ROTATION_270 Rotation = 3
)
