package st7789

import "tinygo.org/x/drivers"

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
	RGBCTRL    = 0xB1
	FRMCTR2    = 0xB2
	PORCTRL    = 0xB2
	FRMCTR3    = 0xB3
	INVCTR     = 0xB4
	DISSET5    = 0xB6
	PWCTR1     = 0xC0
	PWCTR2     = 0xC1
	PWCTR3     = 0xC2
	PWCTR4     = 0xC3
	PWCTR5     = 0xC4
	VMCTR1     = 0xC5
	FRCTRL2    = 0xC6
	PWCTR6     = 0xFC
	GMCTRP1    = 0xE0
	GMCTRN1    = 0xE1
	GSCAN      = 0x45
	VSCRDEF    = 0x33
	VSCRSADD   = 0x37

	ColorRGB444 ColorFormat = 0b011
	ColorRGB565 ColorFormat = 0b101
	ColorRGB666 ColorFormat = 0b111

	NO_ROTATION  = drivers.Rotation0
	ROTATION_90  = drivers.Rotation90 // 90 degrees clock-wise rotation
	ROTATION_180 = drivers.Rotation180
	ROTATION_270 = drivers.Rotation270

	// Allowable frame rate codes for FRCTRL2 (Identifier is in Hz)
	FRAMERATE_111 FrameRate = 0x01
	FRAMERATE_105 FrameRate = 0x02
	FRAMERATE_99  FrameRate = 0x03
	FRAMERATE_94  FrameRate = 0x04
	FRAMERATE_90  FrameRate = 0x05
	FRAMERATE_86  FrameRate = 0x06
	FRAMERATE_82  FrameRate = 0x07
	FRAMERATE_78  FrameRate = 0x08
	FRAMERATE_75  FrameRate = 0x09
	FRAMERATE_72  FrameRate = 0x0A
	FRAMERATE_69  FrameRate = 0x0B
	FRAMERATE_67  FrameRate = 0x0C
	FRAMERATE_64  FrameRate = 0x0D
	FRAMERATE_62  FrameRate = 0x0E
	FRAMERATE_60  FrameRate = 0x0F // 60 is default
	FRAMERATE_58  FrameRate = 0x10
	FRAMERATE_57  FrameRate = 0x11
	FRAMERATE_55  FrameRate = 0x12
	FRAMERATE_53  FrameRate = 0x13
	FRAMERATE_52  FrameRate = 0x14
	FRAMERATE_50  FrameRate = 0x15
	FRAMERATE_49  FrameRate = 0x16
	FRAMERATE_48  FrameRate = 0x17
	FRAMERATE_46  FrameRate = 0x18
	FRAMERATE_45  FrameRate = 0x19
	FRAMERATE_44  FrameRate = 0x1A
	FRAMERATE_43  FrameRate = 0x1B
	FRAMERATE_42  FrameRate = 0x1C
	FRAMERATE_41  FrameRate = 0x1D
	FRAMERATE_40  FrameRate = 0x1E
	FRAMERATE_39  FrameRate = 0x1F

	MAX_VSYNC_SCANLINES = 254
)
