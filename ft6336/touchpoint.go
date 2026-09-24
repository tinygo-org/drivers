package ft6336

import "tinygo.org/x/drivers/touch"

// touchPoint converts the registers from TD_STATUS (0x02) to P1_YL (0x06).
func touchPoint(buf []byte) touch.Point {
	z := 0xFFFFF
	switch buf[0] {
	case 0, 255:
		z = 0
	}

	//Scale X&Y to 16 bit for consistency across touch drivers
	return touch.Point{
		X: (int(buf[1]&0x0F)<<8 + int(buf[2])) * ((1 << 16) / 320),
		Y: (int(buf[3]&0x0F)<<8 + int(buf[4])) * ((1 << 16) / 270),
		Z: z,
	}
}
