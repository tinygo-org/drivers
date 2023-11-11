// Port of Adafruit's "pyportal_boing" demo found here:
// https://github.com/adafruit/Adafruit_ILI9341/blob/master/examples/pyportal_boing
package main

import (
	"time"

	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/examples/ili9341/pyportal_boing/graphics"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/pixel"
)

const (
	BGCOLOR    = pixel.RGB565BE(0x75AD)
	GRIDCOLOR  = pixel.RGB565BE(0x15A8)
	BGSHADOW   = pixel.RGB565BE(0x8552)
	GRIDSHADOW = pixel.RGB565BE(0x0C60)
	RED        = pixel.RGB565BE(0x00F8)
	WHITE      = pixel.RGB565BE(0xFFFF)

	YBOTTOM = 123  // Ball Y coord at bottom
	YBOUNCE = -3.5 // Upward velocity on ball bounce

	_debug = false
)

var (
	frameBuffer = pixel.NewImage[pixel.RGB565BE](graphics.BALLWIDTH+8, graphics.BALLHEIGHT+8)

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
	palette [16]pixel.RGB565BE
)

var (
	display *ili9341.Device
)

func main() {
	display = initdisplay.InitDisplay()

	print("width, height == ")
	width, height := display.Size()
	println(width, height)

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

	for {

		balloldx = ballx // Save prior position
		balloldy = bally
		ballx += ballvx // Update position
		bally += ballvy
		ballvy += 0.06 // Update Y velocity
		if (ballx <= 15) || (ballx >= graphics.SCREENWIDTH-graphics.BALLWIDTH) {
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
		maxx = int16(ballx + graphics.BALLWIDTH - 1)
		if int16(balloldx+graphics.BALLWIDTH-1) > maxx {
			maxx = int16(balloldx + graphics.BALLWIDTH - 1)
		}
		maxy = int16(bally + graphics.BALLHEIGHT - 1)
		if int16(balloldy+graphics.BALLHEIGHT-1) > maxy {
			maxy = int16(balloldy + graphics.BALLHEIGHT - 1)
		}

		width = maxx - minx + 1
		height = maxy - miny + 1
		buffer := frameBuffer.Rescale(int(width), int(height))

		// Ball animation frame # is incremented opposite the ball's X velocity
		ballframe -= ballvx * 0.5
		if ballframe < 0 {
			ballframe += 14 // Constrain from 0 to 13
		} else if ballframe >= 14 {
			ballframe -= 14
		}

		// Set 7 palette entries to white, 7 to red, based on frame number.
		// This makes the ball spin
		for i := 0; i < 14; i++ {
			if (int(ballframe)+i)%14 < 7 {
				palette[i+2] = WHITE
			} else {
				palette[i+2] = RED
			} // Palette entries 0 and 1 aren't used (clear and shadow, respectively)
		}

		// Only the changed rectangle is drawn into the 'renderbuf' array...
		var c pixel.RGB565BE      //, *destPtr;
		bx := minx - int16(ballx) // X relative to ball bitmap (can be negative)
		by := miny - int16(bally) // Y relative to ball bitmap (can be negative)
		bgx := minx               // X relative to background bitmap (>= 0)
		bgy := miny               // Y relative to background bitmap (>= 0)
		var bx1, bgx1 int16       // Loop counters and working vars
		var p uint8               // 'packed' value of 2 ball pixels
		var bufIdx int8 = 0

		//tft.setAddrWindow(minx, miny, width, height)

		for y := 0; y < int(height); y++ { // For each row...
			//destPtr = &renderbuf[bufIdx][0];
			bx1 = bx   // Need to keep the original bx and bgx values,
			bgx1 = bgx // so copies of them are made here (and changed in loop below)
			for x := 0; x < int(width); x++ {
				var bgidx = int(bgy)*(graphics.SCREENWIDTH/8) + int(bgx1/8)
				if (bx1 >= 0) && (bx1 < graphics.BALLWIDTH) && // Is current pixel row/column
					(by >= 0) && (by < graphics.BALLHEIGHT) { // inside the ball bitmap area?
					// Yes, do ball compositing math...
					p = graphics.Ball[int(by*(graphics.BALLWIDTH/2))+int(bx1/2)] // Get packed value (2 pixels)
					var nibble uint8
					if (bx1 & 1) != 0 {
						nibble = p & 0xF
					} else {
						nibble = p >> 4
					} // Unpack high or low nybble
					if nibble == 0 { // Outside ball - just draw grid
						if graphics.Background[bgidx]&(0x80>>(bgx1&7)) != 0 {
							c = GRIDCOLOR
						} else {
							c = BGCOLOR
						}
					} else if nibble > 1 { // In ball area...
						c = palette[nibble]
					} else { // In shadow area...
						if graphics.Background[bgidx]&(0x80>>(bgx1&7)) != 0 {
							c = GRIDSHADOW
						} else {
							c = BGSHADOW
						}
					}
				} else { // Outside ball bitmap, just draw background bitmap...
					if graphics.Background[bgidx]&(0x80>>(bgx1&7)) != 0 {
						c = GRIDCOLOR
					} else {
						c = BGCOLOR
					}
				}
				buffer.Set(x, y, c)
				bx1++ // Increment bitmap position counters (X axis)
				bgx1++
			}
			//tft.dmaWait(); // Wait for prior line to complete
			//tft.writePixels(&renderbuf[bufIdx][0], width, false); // Non-blocking write
			bufIdx = 1 - bufIdx
			by++ // Increment bitmap position counters (Y axis)
			bgy++
		}

		display.DrawBitmap(minx, miny, buffer)

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
	byteWidth := (w + 7) / 8 // Bitmap scanline pad = whole byte
	var b uint8
	buffer := frameBuffer.Rescale(int(w), 1)
	for j := int16(0); j < h; j++ {
		for k := int16(0); k < w; k++ {
			if k&7 > 0 {
				b <<= 1
			} else {
				b = graphics.Background[j*byteWidth+k/8]
			}
			if b&0x80 == 0 {
				buffer.Set(int(k), 0, BGCOLOR)
			} else {
				buffer.Set(int(k), 0, GRIDCOLOR)
			}
		}
		display.DrawBitmap(0, j, buffer)
	}
}
