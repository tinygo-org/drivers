package ili9341

type Rotation uint8

const (

	// source:
	// https://github.com/adafruit/Adafruit_ILI9341/blob/master/Adafruit_ILI9341.h
	/*!
	 * @file Adafruit_ILI9341.h
	 *
	 * This is the documentation for Adafruit's ILI9341 driver for the
	 * Arduino platform.
	 *
	 * This library works with the Adafruit 2.8" Touch Shield V2 (SPI)
	 *    http://www.adafruit.com/products/1651
	 * Adafruit 2.4" TFT LCD with Touchscreen Breakout w/MicroSD Socket - ILI9341
	 *    https://www.adafruit.com/product/2478
	 * 2.8" TFT LCD with Touchscreen Breakout Board w/MicroSD Socket - ILI9341
	 *    https://www.adafruit.com/product/1770
	 * 2.2" 18-bit color TFT LCD display with microSD card breakout - ILI9340
	 *    https://www.adafruit.com/product/1770
	 * TFT FeatherWing - 2.4" 320x240 Touchscreen For All Feathers
	 *    https://www.adafruit.com/product/3315
	 *
	 * These displays use SPI to communicate, 4 or 5 pins are required
	 * to interface (RST is optional).
	 *
	 * Adafruit invests time and resources providing this open source code,
	 * please support Adafruit and open-source hardware by purchasing
	 * products from Adafruit!
	 *
	 *
	 * This library depends on <a href="https://github.com/adafruit/Adafruit_GFX">
	 * Adafruit_GFX</a> being present on your system. Please make sure you have
	 * installed the latest version before using this library.
	 *
	 * Written by Limor "ladyada" Fried for Adafruit Industries.
	 *
	 * BSD license, all text here must be included in any redistribution.
	 */

	TFTWIDTH  = 240 ///< ILI9341 max TFT width
	TFTHEIGHT = 320 ///< ILI9341 max TFT height

	NOP     = 0x00 ///< No-op register
	SWRESET = 0x01 ///< Software reset register
	RDDID   = 0x04 ///< Read display identification information
	RDDST   = 0x09 ///< Read Display Status

	SLPIN  = 0x10 ///< Enter Sleep Mode
	SLPOUT = 0x11 ///< Sleep Out
	PTLON  = 0x12 ///< Partial Mode ON
	NORON  = 0x13 ///< Normal Display Mode ON

	RDMODE     = 0x0A ///< Read Display Power Mode
	RDMADCTL   = 0x0B ///< Read Display MADCTL
	RDPIXFMT   = 0x0C ///< Read Display Pixel Format
	RDIMGFMT   = 0x0D ///< Read Display Image Format
	RDSELFDIAG = 0x0F ///< Read Display Self-Diagnostic Result

	INVOFF   = 0x20 ///< Display Inversion OFF
	INVON    = 0x21 ///< Display Inversion ON
	GAMMASET = 0x26 ///< Gamma Set
	DISPOFF  = 0x28 ///< Display OFF
	DISPON   = 0x29 ///< Display ON

	CASET = 0x2A ///< Column Address Set
	PASET = 0x2B ///< Page Address Set
	RAMWR = 0x2C ///< Memory Write
	RAMRD = 0x2E ///< Memory Read

	PTLAR    = 0x30 ///< Partial Area
	VSCRDEF  = 0x33 ///< Vertical Scrolling Definition
	MADCTL   = 0x36 ///< Memory Access Control
	VSCRSADD = 0x37 ///< Vertical Scrolling Start Address
	PIXFMT   = 0x3A ///< COLMOD: Pixel Format Set

	FRMCTR1 = 0xB1 ///< Frame Rate Control (In Normal Mode/Full Colors)
	FRMCTR2 = 0xB2 ///< Frame Rate Control (In Idle Mode/8 colors)
	FRMCTR3 = 0xB3 ///< Frame Rate control (In Partial Mode/Full Colors)
	INVCTR  = 0xB4 ///< Display Inversion Control
	DFUNCTR = 0xB6 ///< Display Function Control

	PWCTR1 = 0xC0 ///< Power Control 1
	PWCTR2 = 0xC1 ///< Power Control 2
	PWCTR3 = 0xC2 ///< Power Control 3
	PWCTR4 = 0xC3 ///< Power Control 4
	PWCTR5 = 0xC4 ///< Power Control 5
	VMCTR1 = 0xC5 ///< VCOM Control 1
	VMCTR2 = 0xC7 ///< VCOM Control 2

	RDID1 = 0xDA ///< Read ID 1
	RDID2 = 0xDB ///< Read ID 2
	RDID3 = 0xDC ///< Read ID 3
	RDID4 = 0xDD ///< Read ID 4

	GMCTRP1 = 0xE0 ///< Positive Gamma Correction
	GMCTRN1 = 0xE1 ///< Negative Gamma Correction
	//PWCTR6     0xFC

	// Color definitions
	BLACK       = 0x0000 ///<   0,   0,   0
	NAVY        = 0x000F ///<   0,   0, 123
	DARKGREEN   = 0x03E0 ///<   0, 125,   0
	DARKCYAN    = 0x03EF ///<   0, 125, 123
	MAROON      = 0x7800 ///< 123,   0,   0
	PURPLE      = 0x780F ///< 123,   0, 123
	OLIVE       = 0x7BE0 ///< 123, 125,   0
	LIGHTGREY   = 0xC618 ///< 198, 195, 198
	DARKGREY    = 0x7BEF ///< 123, 125, 123
	BLUE        = 0x001F ///<   0,   0, 255
	GREEN       = 0x07E0 ///<   0, 255,   0
	CYAN        = 0x07FF ///<   0, 255, 255
	RED         = 0xF800 ///< 255,   0,   0
	MAGENTA     = 0xF81F ///< 255,   0, 255
	YELLOW      = 0xFFE0 ///< 255, 255,   0
	WHITE       = 0xFFFF ///< 255, 255, 255
	ORANGE      = 0xFD20 ///< 255, 165,   0
	GREENYELLOW = 0xAFE5 ///< 173, 255,  41
	PINK        = 0xFC18 ///< 255, 130, 198
)

const (
	Rotation0   Rotation = 0
	Rotation90  Rotation = 1 // 90 degrees clock-wise rotation
	Rotation180 Rotation = 2
	Rotation270 Rotation = 3
)
