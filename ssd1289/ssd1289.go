// Package ssd1289 implements a driver for the SSD1289 led matrix controller as packaged on the TFT_320QVT board
//
// Datasheet: http://aitendo3.sakura.ne.jp/aitendo_data/product_img/lcd/tft2/M032C1289TP/3.2-SSD1289.pdf
package ssd1289

import (
	"image/color"
	"machine"
	"time"
)

type Bus interface {
	Set(data uint16)
}

type Device struct {
	rs  machine.Pin
	wr  machine.Pin
	cs  machine.Pin
	rst machine.Pin
	bus Bus
}

const width = int16(240)
const height = int16(320)

func New(rs machine.Pin, wr machine.Pin, cs machine.Pin, rst machine.Pin, bus Bus) Device {
	d := Device{
		rs:  rs,
		wr:  wr,
		cs:  cs,
		rst: rst,
		bus: bus,
	}

	rs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	wr.Configure(machine.PinConfig{Mode: machine.PinOutput})
	cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rst.Configure(machine.PinConfig{Mode: machine.PinOutput})

	cs.High()
	rst.High()
	wr.High()

	return d
}

func (d *Device) lcdWriteCom(cmd Command) {
	d.rs.Low()
	d.lcdWriteBusInt(uint16(cmd))
}

func (d *Device) lcdWriteDataInt(data uint16) {
	d.rs.High()
	d.lcdWriteBusInt(data)
}

func (d *Device) lcdWriteComData(cmd Command, data uint16) {
	d.lcdWriteCom(cmd)
	d.lcdWriteDataInt(data)
}

func (d *Device) tx() {
	d.wr.Low()
	d.wr.High()
}

func (d *Device) lcdWriteBusInt(data uint16) {
	d.bus.Set(data)
	d.tx()
}

func (d *Device) Configure() {
	d.rst.High()
	time.Sleep(time.Millisecond * 5)
	d.rst.Low()
	time.Sleep(time.Millisecond * 15)
	d.rst.High()
	time.Sleep(time.Millisecond * 15)
	d.cs.Low()

	//Power supply setting
	d.lcdWriteComData(POWERCONTROL1, 0xA8A4)
	d.lcdWriteComData(POWERCONTROL2, 0x0000)
	d.lcdWriteComData(POWERCONTROL3, 0x080C)
	d.lcdWriteComData(POWERCONTROL4, 0x2B00)
	d.lcdWriteComData(POWERCONTROL5, 0x00B7)

	//Set R07h at 0021h
	d.lcdWriteComData(DISPLAYCONTROL, 0x021)

	//Set R00h at 0001h
	d.lcdWriteComData(OSCILLATIONSTART, 0x0001)

	//Set R07h at 0021h
	d.lcdWriteComData(DISPLAYCONTROL, 0x023)

	//Set R10h at 0000h, Exit sleep mode
	d.lcdWriteComData(SLEEPMODE, 0x0000)

	//Wait 30ms
	time.Sleep(time.Millisecond * 30)

	//Set R07h at 0033h
	d.lcdWriteComData(DISPLAYCONTROL, 0x033)

	//Entry Mode setting (R11h)
	//DFM   11 --> 65k
	//TRANS  0
	//OEDEF  0
	//WMODE  0 --> Normal data bus
	//DMODE 00 --> Ram
	//TY    01 --> 262k Type A not used as we are in 65k mode.
	//ID    11 --> Horizontal & Vertical increment
	//AM     0 --> Horizontal
	//LG   000 --> No compare register usage
	d.lcdWriteComData(ENTRYMODE, 0x6030)

	//LCD Driver AC Setting
	//I couldn't make sense of the documentation fortunately 0 seems to
	//FLD 0 --> Normal driving
	//ENWS 0 --> POR mode
	//BC   1 --> Less flicker
	//EOR  1 --> Less stripey
	//WSMD 0 --> not used in POR mode
	//NW   0 --> Least flicker
	d.lcdWriteComData(LCDDRIVEACCONTROL, 0x0600)

	//End of documented init

	//RL    0 --> Output shift direction
	//REV   1 --> Reverse colors
	//CAD   0 --> Cs on common
	//BGR   0 --> use RGB color assignment
	//SM    0 --> standard gate scan sequence
	//TB    1 --> Display is mirrored with 0
	//MUX 319 --> Number of lines in display
	d.lcdWriteComData(DRIVEROUTPUTCONTROL, 0x233F)

	d.cs.High()

}

func (d *Device) setXY(x1 uint16, y1 uint16, x2 uint16, y2 uint16) {
	d.lcdWriteComData(HORIZONTALRAMADDRESSPOSITION, (x2<<8)+x1)
	d.lcdWriteComData(VERTICALRAMADDRESSSTARTPOSITION, y1)
	d.lcdWriteComData(VERTICALRAMADDRESSENDPOSITION, y2)
	d.lcdWriteComData(SETGDDRAMXADDRESSCOUNTER, x1)
	d.lcdWriteComData(SETGDDRAMYADDRESSCOUNTER, y1)
	d.lcdWriteCom(RAMDATAREADWRITE)
}

func (d *Device) ClearDisplay() {
	d.FillDisplay(color.RGBA{0, 0, 0, 255})
}

func (d *Device) FillDisplay(c color.RGBA) {
	d.FillRect(0, 0, width, height, c)
}

func encodeColor(c color.RGBA) uint16 {
	encoded := (uint16(c.B)&248)<<8 | (uint16(c.G)&252)<<3 | (uint16(c.R)&248)>>3
	return encoded
}

func (d *Device) SetPixel(x, y int16, c color.RGBA) {

	encoded := encodeColor(c)

	d.cs.Low()
	d.setXY(uint16(x), uint16(y), uint16(x), uint16(y))
	d.rs.High()
	d.lcdWriteBusInt(encoded)
	d.cs.High()
}

func (d *Device) FillRect(x, y, w, h int16, c color.RGBA) {
	encoded := encodeColor(c)

	d.cs.Low()
	d.setXY(uint16(x), uint16(y), uint16(x+(w-1)), uint16(y+(h-1)))
	d.rs.High()
	d.bus.Set(encoded)
	for i := int64(0); i < int64(w)*int64(h); i++ {
		d.tx()
	}
	d.cs.High()
	d.rs.Low()

}

func (d *Device) Display() error {
	//Not enough memory to store an entire screen on most microcontrollers
	return nil
}

func (d *Device) Size() (x, y int16) {
	return width, height
}
