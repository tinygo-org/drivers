package pixel

import (
	"unsafe"
)

// Image buffer, used for working with the native image format of various
// displays. It works a lot like a slice: it can be rescaled while reusing the
// underlying buffer and should be passed around by value.
type Image[T Color] struct {
	width  int16
	height int16
	data   unsafe.Pointer
}

// NewImage creates a new image of the given size.
func NewImage[T Color](width, height int) Image[T] {
	if width < 0 || height < 0 || int(int16(width)) != width || int(int16(height)) != height {
		// The width/height are stored as 16-bit integers and should never be
		// negative.
		panic("NewImage: width/height out of bounds")
	}
	var zeroColor T
	var data unsafe.Pointer
	if zeroColor.BitsPerPixel()%8 == 0 {
		// Typical formats like RGB888 and RGB565.
		// Each color starts at a whole byte offset from the start.
		buf := make([]T, width*height)
		data = unsafe.Pointer(&buf[0])
	} else {
		// Formats like RGB444 that have 12 bits per pixel.
		// We access these as bytes, so allocate the buffer as a byte slice.
		bufBits := width * height * zeroColor.BitsPerPixel()
		bufBytes := (bufBits + 7) / 8
		buf := make([]byte, bufBytes)
		data = unsafe.Pointer(&buf[0])
	}
	return Image[T]{
		width:  int16(width),
		height: int16(height),
		data:   data,
	}
}

// Rescale returns a new Image buffer based on the img buffer.
// The contents is undefined after the Rescale operation, and any modification
// to the returned image will overwrite the underlying image buffer in undefined
// ways. It will panic if width*height is larger than img.Len().
func (img Image[T]) Rescale(width, height int) Image[T] {
	if width*height > img.Len() {
		panic("Image.Rescale size out of bounds")
	}
	return Image[T]{
		width:  int16(width),
		height: int16(height),
		data:   img.data,
	}
}

// LimitHeight returns a subimage with the bottom part cut off, as specified by
// height.
func (img Image[T]) LimitHeight(height int) Image[T] {
	if height < 0 || height > int(img.height) {
		panic("Image.LimitHeight: out of bounds")
	}
	return Image[T]{
		width:  img.width,
		height: int16(height),
		data:   img.data,
	}
}

// Len returns the number of pixels in this image buffer.
func (img Image[T]) Len() int {
	return int(img.width) * int(img.height)
}

// RawBuffer returns a byte slice that can be written directly to the screen
// using DrawRGBBitmap8.
func (img Image[T]) RawBuffer() []uint8 {
	var zeroColor T
	var numBytes int
	if zeroColor.BitsPerPixel()%8 == 0 {
		// Each color starts at a whole byte offset.
		numBytes = int(unsafe.Sizeof(zeroColor)) * int(img.width) * int(img.height)
	} else {
		// Formats like RGB444 that aren't a whole number of bytes.
		numBits := zeroColor.BitsPerPixel() * int(img.width) * int(img.height)
		numBytes = (numBits + 7) / 8 // round up (see NewImage)
	}
	return unsafe.Slice((*byte)(img.data), numBytes)
}

// Size returns the image size.
func (img Image[T]) Size() (int, int) {
	return int(img.width), int(img.height)
}

func (img Image[T]) setPixel(index int, c T) {
	var zeroColor T

	if zeroColor.BitsPerPixel()%8 == 0 {
		// Each color starts at a whole byte offset.
		// This is the easy case.
		offset := index * int(unsafe.Sizeof(zeroColor))
		ptr := unsafe.Add(img.data, offset)
		*((*T)(ptr)) = c
		return
	}

	if c, ok := any(c).(RGB444BE); ok {
		// Special case for RGB444.
		bitIndex := index * zeroColor.BitsPerPixel()
		if bitIndex%8 == 0 {
			byteOffset := bitIndex / 8
			ptr := (*[2]byte)(unsafe.Add(img.data, byteOffset))
			ptr[0] = uint8(c >> 4)
			ptr[1] = ptr[1]&0x0f | uint8(c)<<4 // change top bits
		} else {
			byteOffset := bitIndex / 8
			ptr := (*[2]byte)(unsafe.Add(img.data, byteOffset))
			ptr[0] = ptr[0]&0xf0 | uint8(c>>8) // change bottom bits
			ptr[1] = uint8(c)
		}
		return
	}

	// TODO: the code for RGB444 should be generalized to support any bit size.
	panic("todo: setPixel for odd bits per pixel")
}

// Set sets the pixel at x, y to the given color.
// Use FillSolidColor to efficiently fill the entire image buffer.
func (img Image[T]) Set(x, y int, c T) {
	if uint(x) >= uint(int(img.width)) || uint(y) >= uint(int(img.height)) {
		panic("Image.Set: out of bounds")
	}
	index := y*int(img.width) + x
	img.setPixel(index, c)
}

// Get returns the color at the given index.
func (img Image[T]) Get(x, y int) T {
	if uint(x) >= uint(int(img.width)) || uint(y) >= uint(int(img.height)) {
		panic("Image.Get: out of bounds")
	}
	var zeroColor T
	index := y*int(img.width) + x // index into img.data

	if zeroColor.BitsPerPixel()%8 == 0 {
		// Colors like RGB565, RGB888, etc.
		offset := index * int(unsafe.Sizeof(zeroColor))
		ptr := unsafe.Add(img.data, offset)
		return *((*T)(ptr))
	}

	if _, ok := any(zeroColor).(RGB444BE); ok {
		// Special case for RGB444 that isn't stored in a neat byte multiple.
		bitIndex := index * zeroColor.BitsPerPixel()
		var c RGB444BE
		if bitIndex%8 == 0 {
			byteOffset := bitIndex / 8
			ptr := (*[2]byte)(unsafe.Add(img.data, byteOffset))
			c |= RGB444BE(ptr[0]) << 4
			c |= RGB444BE(ptr[1] >> 4) // load top bits
		} else {
			byteOffset := bitIndex / 8
			ptr := (*[2]byte)(unsafe.Add(img.data, byteOffset))
			c |= RGB444BE(ptr[0]&0x0f) << 8 // load bottom bits
			c |= RGB444BE(ptr[1])
		}
		return any(c).(T)
	}

	// TODO: generalize the above code.
	panic("todo: Image.Get for odd bits per pixel")
}

// FillSolidColor fills the entire image with the given color.
// This may be faster than setting individual pixels.
func (img Image[T]) FillSolidColor(color T) {
	var zeroColor T

	// Fast pass for colors of 8, 16, 24, etc bytes in size.
	if zeroColor.BitsPerPixel()%8 == 0 {
		ptr := img.data
		for i := 0; i < img.Len(); i++ {
			// TODO: this can be optimized a lot.
			// - The store can be done as a 32-bit integer, after checking for
			//   alignment.
			// - Perhaps the loop can be unrolled to improve copy performance.
			*(*T)(ptr) = color
			ptr = unsafe.Add(ptr, unsafe.Sizeof(zeroColor))
		}
		return
	}

	// Special case for RGB444.
	if c, ok := any(color).(RGB444BE); ok {
		// RGB444 can be stored in a more optimized way, by storing two colors
		// at a time instead of setting each color individually. This avoids
		// loading and masking the old color bits for the half-bytes.
		var buf [3]uint8
		buf[0] = uint8(c >> 4)
		buf[1] = uint8(c)<<4 | uint8(c>>8)
		buf[2] = uint8(c)
		rawBuf := unsafe.Slice((*[3]byte)(img.data), img.Len()/2)
		for i := 0; i < len(rawBuf); i++ {
			rawBuf[i] = buf
		}
		if img.Len()%2 != 0 {
			// The image contains an uneven number of pixels.
			// This is uncommon, but it can happen and we have to handle it.
			img.setPixel(img.Len()-1, color)
		}
		return
	}

	// Fallback for other color formats.
	for i := 0; i < img.Len(); i++ {
		img.setPixel(i, color)
	}
}
