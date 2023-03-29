package sh1106

// Registers
const (
	Address = 0x3C

	SETCONTRAST                          = 0x81
	DISPLAYALLON_RESUME                  = 0xA4
	DISPLAYALLON                         = 0xA5
	NORMALDISPLAY                        = 0xA6
	INVERTDISPLAY                        = 0xA7
	DISPLAYOFF                           = 0xAE
	DISPLAYON                            = 0xAF
	SETDISPLAYOFFSET                     = 0xD3
	SETCOMPINS                           = 0xDA
	SETVCOMDETECT                        = 0xDB
	SETDISPLAYCLOCKDIV                   = 0xD5
	SETPRECHARGE                         = 0xD9
	SETMULTIPLEX                         = 0xA8
	SETLOWCOLUMN                         = 0x00
	SETHIGHCOLUMN                        = 0x10
	SETSTARTLINE                         = 0x40
	MEMORYMODE                           = 0x20
	COLUMNADDR                           = 0x21
	PAGEADDR                             = 0x22
	COMSCANINC                           = 0xC0
	COMSCANDEC                           = 0xC8
	SEGREMAP                             = 0xA0
	CHARGEPUMP                           = 0x8D
	ACTIVATE_SCROLL                      = 0x2F
	DEACTIVATE_SCROLL                    = 0x2E
	SET_VERTICAL_SCROLL_AREA             = 0xA3
	RIGHT_HORIZONTAL_SCROLL              = 0x26
	LEFT_HORIZONTAL_SCROLL               = 0x27
	VERTICAL_AND_RIGHT_HORIZONTAL_SCROLL = 0x29
	VERTICAL_AND_LEFT_HORIZONTAL_SCROLL  = 0x2A

	EXTERNALVCC  VccMode = 0x1
	SWITCHCAPVCC VccMode = 0x2
)
