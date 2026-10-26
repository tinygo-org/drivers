// Demo for the Pimoroni Display-o-Tron HAT on a Waveshare RP2040-PiZero.
//
// Board wiring.
//
// The pin numbers below are only correct for the RP2040-PiZero. Any other
// board needs its own mapping. See the epd2in13-pizero demo in this
// examples tree for how that mapping was measured.
//
// The RP2040-PiZero does not map every header pin to the RP2040 GPIO of
// the same number. Waveshare rearranges the pins around the SPI block so
// that GPIO10 and GPIO11 land on the Raspberry Pi SCLK and MOSI positions,
// which shifts GPIO12 onto pin 21 and puts GPIO9 on pin 32. The reset line
// therefore has to be driven from GPIO9, not GPIO12.
// Schematic: https://files.waveshare.com/wiki/RP2040-PiZero/RP2040-PiZero.pdf
//
// Header to RP2040 GPIO for the Display-o-Tron HAT signals:
//
//	pin 3  SDA          GPIO2  (I2C1 SDA)
//	pin 5  SCL          GPIO3  (I2C1 SCL)
//	pin 19 LCD MOSI     GPIO11 (bit-banged data)
//	pin 22 LCD RS       GPIO25
//	pin 23 LCD SCLK     GPIO10 (bit-banged clock)
//	pin 24 LCD CS       GPIO8
//	pin 32 LCD RESET    GPIO9
//
// The LCD link is bit-banged rather than driven by the SPI1 peripheral.
// Both were tested on this board: the SPI1 version leaves the display
// blank, the bit-banged version works.
//
// The SN3218 backlight driver and CAP1166 touch and bar graph controller
// both sit on I2C1, at their fixed addresses 0x54 and 0x2C.
//
// Build and flash:
//
//	tinygo flash -target=waveshare-rp2040-pizero ./examples/dothat-pizero
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/cap1166"
	"tinygo.org/x/drivers/sn3218"
	"tinygo.org/x/drivers/st7036"
)

var (
	lcd   st7036.Device
	light sn3218.Device
	touch cap1166.Device
)

// bitbangSPI is a software SPI master, MSB first, mode 0 (clock idle low,
// data sampled on the rising edge). Used here instead of the RP2040's SPI1
// hardware peripheral, since that peripheral produced no visible signal on
// GPIO11 (the LCD data pin) during scope debugging.
type bitbangSPI struct {
	sck, sdo machine.Pin
}

func newBitbangSPI(sck, sdo machine.Pin) bitbangSPI {
	sck.Configure(machine.PinConfig{Mode: machine.PinOutput})
	sdo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	sck.Low()
	sdo.Low()
	return bitbangSPI{sck: sck, sdo: sdo}
}

func (b bitbangSPI) Transfer(out byte) (byte, error) {
	for i := 7; i >= 0; i-- {
		if out&(1<<uint(i)) != 0 {
			b.sdo.High()
		} else {
			b.sdo.Low()
		}
		delay()
		b.sck.High()
		delay()
		b.sck.Low()
	}
	return 0, nil
}

// delay holds each clock phase long enough to keep the bit-banged link
// near 100kHz. The ST7036 is rated for 1MHz at most, and the Pimoroni
// driver runs it at that speed.
func delay() {
	for i := 0; i < 40; i++ {
		machine.GPIO0.Get()
	}
}

func (b bitbangSPI) Tx(w, r []byte) error {
	for _, out := range w {
		if _, err := b.Transfer(out); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// Wait for the USB serial console to attach so the log is not lost.
	time.Sleep(3 * time.Second)

	lcdBus := newBitbangSPI(machine.GPIO10, machine.GPIO11)

	lcd = st7036.New(lcdBus, machine.GPIO8, machine.GPIO25, machine.GPIO9)
	if err := lcd.Configure(st7036.Config{Rows: 3, Columns: 16}); err != nil {
		println("LCD configure failed:", err.Error())
		return
	}
	println("LCD configure ok")

	err := machine.I2C1.Configure(machine.I2CConfig{
		Frequency: 400000,
		SDA:       machine.GPIO2,
		SCL:       machine.GPIO3,
	})
	if err != nil {
		println("I2C1 configure failed:", err.Error())
		return
	}

	light = sn3218.New(machine.I2C1)
	if err := light.Configure(); err != nil {
		println("backlight configure failed:", err.Error())
		return
	}

	touch = cap1166.New(machine.I2C1, cap1166.DefaultAddress)
	if err := touch.Configure(); err != nil {
		println("touch configure failed:", err.Error())
		return
	}
	// The white bar graph LEDs are uncomfortably bright, keep them off.
	touch.GraphOff()

	println("Display-o-Tron HAT bring-up")
	// 50 reads well head on with this panel. The Pimoroni default of 40
	// leaves the characters looking like solid blocks.
	contrast := uint8(50)
	lcd.SetContrast(contrast)
	lcd.SetDisplayMode(true, false, false)
	lcd.SetCursorPosition(0, 0)
	lcd.Write([]byte("Display-o-Tron"))
	lcd.SetCursorPosition(0, 1)
	lcd.Write([]byte("up/dn: contrast"))

	lastButton := ""
	var hue float32

	for {
		// Sweep the 6 backlight zones through the colour wheel.
		hue += 0.01
		if hue >= 1 {
			hue -= 1
		}
		setBacklightSweep(hue)

		// Poll the 6 touch buttons and report the first one pressed.
		status, err := touch.InputStatus()
		if err == nil && status != 0 {
			name := buttonName(status)
			if name != lastButton {
				lastButton = name
				println("pressed:", name)

				// Up and down tune the contrast, so the best value
				// for a given panel can be found by eye.
				switch name {
				case "up":
					if contrast < 0x3F {
						contrast++
						lcd.SetContrast(contrast)
					}
				case "down":
					if contrast > 0 {
						contrast--
						lcd.SetContrast(contrast)
					}
				}

				lcd.SetCursorPosition(0, 2)
				lcd.Write([]byte(pad(name+" c="+itoa(int(contrast)), 16)))
				println("contrast:", contrast)
			}
			touch.ClearInterrupt()
		} else {
			lastButton = ""
		}

		time.Sleep(20 * time.Millisecond)
	}
}

// pad returns s padded with spaces to width, so a shorter line fully
// overwrites whatever was on the display before.
func pad(s string, width int) string {
	for len(s) < width {
		s += " "
	}
	return s[:width]
}

// itoa formats a small non-negative integer.
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [3]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

// buttonName returns the label of the lowest numbered touched input in
// status, matching the silk-screened button names on the HAT.
func buttonName(status uint8) string {
	switch {
	case status&(1<<cap1166.Cancel) != 0:
		return "cancel"
	case status&(1<<cap1166.Up) != 0:
		return "up"
	case status&(1<<cap1166.Down) != 0:
		return "down"
	case status&(1<<cap1166.Left) != 0:
		return "left"
	case status&(1<<cap1166.Button) != 0:
		return "select"
	case status&(1<<cap1166.Right) != 0:
		return "right"
	default:
		return ""
	}
}

// setBacklightSweep lights the 6 backlight zones with a gradient of hues
// centred on hue.
func setBacklightSweep(hue float32) {
	const sweepRange = 0.0833

	var values [sn3218.NumChannels]uint8
	for zone := 0; zone < 6; zone++ {
		h := hue + sweepRange*float32(zone)
		for h >= 1 {
			h -= 1
		}
		r, g, b := hueToRGB(h)
		// Channel order per zone is B, G, R.
		values[zone*3+0] = b
		values[zone*3+1] = g
		values[zone*3+2] = r
	}
	light.SetChannels(values)
}

// hueToRGB converts a hue in the range 0.0 to 1.0 to full brightness,
// fully saturated RGB values.
func hueToRGB(hue float32) (r, g, b uint8) {
	h6 := hue * 6
	i := int(h6)
	f := h6 - float32(i)
	q := 1 - f
	t := f

	var rf, gf, bf float32
	switch i % 6 {
	case 0:
		rf, gf, bf = 1, t, 0
	case 1:
		rf, gf, bf = q, 1, 0
	case 2:
		rf, gf, bf = 0, 1, t
	case 3:
		rf, gf, bf = 0, q, 1
	case 4:
		rf, gf, bf = t, 0, 1
	default:
		rf, gf, bf = 1, 0, q
	}
	return uint8(rf * 255), uint8(gf * 255), uint8(bf * 255)
}
