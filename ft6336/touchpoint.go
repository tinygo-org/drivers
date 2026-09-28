package ft6336

import "tinygo.org/x/drivers/touch"

const (
	defaultWidth  = 320
	defaultHeight = 270
)

// panelSize returns the panel size, with the default for each value less than 2.
func panelSize(width, height int) (int, int) {
	if width < 2 {
		width = defaultWidth
	}
	if height < 2 {
		height = defaultHeight
	}
	return width, height
}

// touchPoint converts the registers TD_STATUS (0x02) through P1_YL (0x06).
// See sections 3.1.3 to 3.1.7 of the application note in the FT6236 datasheet.
func touchPoint(buf []byte, width, height int) touch.Point {
	z := 0xFFFFF
	switch buf[0] {
	case 0, 255:
		z = 0
	}

	x := int(buf[1]&0x0F)<<8 + int(buf[2])
	y := int(buf[3]&0x0F)<<8 + int(buf[4])

	//Scale X&Y to 16 bit for consistency across touch drivers
	return touch.Point{
		X: min(x*0xFFFF/(width-1), 0xFFFF),
		Y: min(y*0xFFFF/(height-1), 0xFFFF),
		Z: z,
	}
}
