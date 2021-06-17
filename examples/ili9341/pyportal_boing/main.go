// Port of Adafruit's "pyportal_boing" demo found here:
// https://github.com/adafruit/Adafruit_ILI9341/blob/master/examples/pyportal_boing
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/examples/ili9341/pyportal_boing/graphics"
	"tinygo.org/x/drivers/ili9341"
)

const (
	BALLWIDTH  = 136
	BALLHEIGHT = 100
)

const (
	SCREENHEIGHT = 240
	SCREENWIDTH  = 320
)

const (
	invBGCOLOR    = 0x75AD
	invGRIDCOLOR  = 0x15A8
	invBGSHADOW   = 0x8552
	invGRIDSHADOW = 0x0C60
	invRED        = 0x00F8
	invWHITE      = 0xFFFF
)

const (
	BGCOLOR    = 0xAD75
	GRIDCOLOR  = 0xA815
	BGSHADOW   = 0x5285
	GRIDSHADOW = 0x600C
	RED        = 0xF800
	WHITE      = 0xFFFF

	YBOTTOM = 123  // Ball Y coord at bottom
	YBOUNCE = -3.5 // Upward velocity on ball bounce

	_debug = false
)

var (
	dbg5 = machine.D5
	dbg6 = machine.D6
)

var (
	frameBuffer = [2][(BALLHEIGHT + 8) * (BALLWIDTH + 8)]uint16{}

	startTime int64
	frame     int64

	// Ball coordinates are stored floating-point because screen refresh
	// is so quick, whole-pixel movements are just too fast!
	ballx     float32
	bally     float32
	ballvx    float32
	ballvy    float32
	ballframe float32
	balloldx  float32
	balloldy  float32

	// Color table for ball rotation effect
	palette [16]uint16
)

func main() {

	// configure backlight
	backlight.Configure(machine.PinConfig{machine.PinOutput})

	// configure display
	display.Configure(ili9341.Config{})
	print("width, height == ")
	width, height := display.Size()
	println(width, height)

	backlight.High()

	display.SetRotation(ili9341.Rotation270)
	DrawBackground()

	startTime = time.Now().UnixNano()
	frame = 0

	ballx = 20.0
	bally = YBOTTOM // Current ball position
	ballvx = 0.8
	ballvy = YBOUNCE // Ball velocity
	ballframe = 3    // Ball animation frame #
	balloldx = ballx
	balloldy = bally // Prior ball position

	var bufIdx int8 = 0

	for {
		dbg5.High()
		bufIdx = 1 - bufIdx

		balloldx = ballx // Save prior position
		balloldy = bally
		ballx += ballvx // Update position
		bally += ballvy
		ballvy += 0.06 // Update Y velocity
		if (ballx <= 15) || (ballx >= SCREENWIDTH-BALLWIDTH) {
			ballvx *= -1 // Left/right bounce
		}
		if bally >= YBOTTOM { // Hit ground?
			bally = YBOTTOM  // Clip and
			ballvy = YBOUNCE // bounce up
		}

		// Determine screen area to update.  This is the bounds of the ball's
		// prior and current positions, so the old ball is fully erased and new
		// ball is fully drawn.
		var minx, miny, maxx, maxy, width, height int16

		// Determine bounds of prior and new positions
		minx = int16(ballx)
		if int16(balloldx) < minx {
			minx = int16(balloldx)
		}
		miny = int16(bally)
		if int16(balloldy) < miny {
			miny = int16(balloldy)
		}
		maxx = int16(ballx + BALLWIDTH - 1)
		if int16(balloldx+BALLWIDTH-1) > maxx {
			maxx = int16(balloldx + BALLWIDTH - 1)
		}
		maxy = int16(bally + BALLHEIGHT - 1)
		if int16(balloldy+BALLHEIGHT-1) > maxy {
			maxy = int16(balloldy + BALLHEIGHT - 1)
		}

		width = maxx - minx + 1
		height = maxy - miny + 1

		// Ball animation frame # is incremented opposite the ball's X velocity
		ballframe -= ballvx * 0.5
		if ballframe < 0 {
			ballframe += 14 // Constrain from 0 to 13
		} else if ballframe >= 14 {
			ballframe -= 14
		}

		//// Set 7 palette entries to white, 7 to red, based on frame number.
		//// This makes the ball spin
		for i := 0; i < 14; i++ {
			if (int(ballframe)+i)%14 < 7 {
				palette[i+2] = invWHITE
			} else {
				palette[i+2] = invRED
			} // Palette entries 0 and 1 aren't used (clear and shadow, respectively)
		}

		// Only the changed rectangle is drawn into the 'renderbuf' array...
		var c uint16              //, *destPtr;
		bx := minx - int16(ballx) // X relative to ball bitmap (can be negative)
		by := miny - int16(bally) // Y relative to ball bitmap (can be negative)
		bgx := minx               // X relative to background bitmap (>= 0)
		bgy := miny               // Y relative to background bitmap (>= 0)
		//var bufIdx int8 = 0

		//tft.setAddrWindow(minx, miny, width, height)
		//dbg5.Low()
		dbg6.High()
		//fmt.Printf("%d < %d < %d < %d\r\n", by, 0, BALLHEIGHT, height)

		y := 0
		if by < 0 {
			max := -1 * int(by)
			for y = 0; y < max; y++ { // For each row...
				var bgidxBase = int(bgy)*(SCREENWIDTH) + int(bgx)
				var yBase = y * int(width)
				for x := 0; x < int(width); x++ {
					frameBuffer[bufIdx][yBase+x] = graphics.Background[bgidxBase+x]
				}
				bgy++
			}
		}

		y2 := y
		max := 0
		if bx < 0 {
			max = -1 * int(bx)
			bgy2 := bgy
			for y = y2; y < y2+int(BALLHEIGHT); y++ { // For each row...
				var bgidxBase = int(bgy2)*(SCREENWIDTH) + int(bgx)
				var yBase = y * int(width)
				//fmt.Printf("- %d %d %d %d %d %d\r\n", bgy, y, bgx, max, yBase, bgidxBase)
				for x := 0; x < int(max); x++ {
					//fmt.Printf("  %d %d\r\n", yBase+x, bgidxBase+x)
					frameBuffer[bufIdx][yBase+x] = graphics.Background[bgidxBase+x]
				}
				bgy2++
			}
			//fmt.Printf("(%d, %d) - (%d, %d)\r\n", bx, 0, -1, BALLHEIGHT-1)
		}

		{
			bgy2 := bgy
			//fmt.Printf("(%d, %d) - (%d, %d)\r\n", 0, 0, BALLWIDTH-1, BALLHEIGHT-1)
			for y = y2; y < y2+int(BALLHEIGHT); y++ { // For each row...
				var bgidxBase = int(bgy2)*(SCREENWIDTH) + int(bgx)
				var byBase = (y - y2) * BALLWIDTH
				var yBase = y * int(width)
				for x := max; x < int(BALLWIDTH)+max; x++ {
					//fmt.Printf("%d %d %d %d\r\n", byBase, x, bgidxBase, yBase)
					//time.Sleep(1 * time.Millisecond)
					// Yes, do ball compositing math...
					c = uint16(graphics.Ball[int(byBase)+x-max]) // Get packed value (2 pixels)

					if c == 0 { // Outside ball - just draw grid
						c = graphics.Background[bgidxBase+x]
					} else if c > 1 { // In ball area...
						c = palette[c]
					} else { // In shadow area...
						c = graphics.BackgroundShadow[bgidxBase+x]
					}
					frameBuffer[bufIdx][yBase+x] = c
				}
				bgy2++
			}
		}

		{
			bgy2 := bgy
			for y = y2; y < y2+int(BALLHEIGHT); y++ { // For each row...
				var bgidxBase = int(bgy2)*(SCREENWIDTH) + int(bgx)
				var yBase = y * int(width)
				//fmt.Printf("+ %d %d %d %d %d\r\n", bgy, y, bgx, yBase, bgidxBase)
				for x := int(BALLWIDTH) + max; x < int(width); x++ {
					frameBuffer[bufIdx][yBase+x] = graphics.Background[bgidxBase+x]
				}
				bgy2++
			}
		}

		y = y2 + int(BALLHEIGHT)
		bgy += BALLHEIGHT
		{
			for ; y < int(height); y++ { // For each row...
				//destPtr = &renderbuf[bufIdx][0];
				var bgidxBase = int(bgy)*(SCREENWIDTH) + int(bgx)
				var yBase = y * int(width)
				for x := 0; x < int(width); x++ {
					frameBuffer[bufIdx][yBase+x] = graphics.Background[bgidxBase+x]
				}
				bgy++
			}
		}
		dbg6.Low()

		display.DrawRGBBitmap(minx, miny, frameBuffer[bufIdx][:width*height], width, height)
		//time.Sleep(10 * time.Millisecond)

		// Show approximate frame rate
		frame++
		if frame&255 == 0 { // Every 256 frames...
			elapsed := (time.Now().UnixNano() - startTime) / int64(time.Second)
			if elapsed > 0 {
				println(frame/elapsed, " fps")
			}
		}
	}
}

func DrawBackground() {
	w, h := display.Size()
	var bufIdx int8 = 0
	for j := 0; j < int(h); j++ {
		bufIdx = 1 - bufIdx
		for k := 0; k < int(w); k++ {
			frameBuffer[bufIdx][k] = graphics.Background[j*int(w)+k]
		}
		display.DrawRGBBitmap(0, int16(j), frameBuffer[bufIdx][0:w], w, 1)
		time.Sleep(1 * time.Millisecond)
	}
}
