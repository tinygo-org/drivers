package ssd1289

type Command byte

const (
	OSCILLATIONSTART                Command = 0x00
	DRIVEROUTPUTCONTROL                     = 0x01
	POWERCONTROL1                           = 0x03
	POWERCONTROL2                           = 0x0C
	POWERCONTROL3                           = 0x0D
	POWERCONTROL4                           = 0x0E
	POWERCONTROL5                           = 0x1E
	DISPLAYCONTROL                          = 0x07
	SLEEPMODE                               = 0x10
	ENTRYMODE                               = 0x11
	LCDDRIVEACCONTROL                       = 0x02
	HORIZONTALRAMADDRESSPOSITION            = 0x44
	VERTICALRAMADDRESSSTARTPOSITION         = 0x45
	VERTICALRAMADDRESSENDPOSITION           = 0x46
	SETGDDRAMYADDRESSCOUNTER                = 0x4F
	SETGDDRAMXADDRESSCOUNTER                = 0x4E
	RAMDATAREADWRITE                        = 0x22
)
