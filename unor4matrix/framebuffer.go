package unor4matrix

const (
	// Width is the number of physical LED columns. Size reports the rotated area.
	Width = 12
	// Height is the number of physical LED rows. Size reports the rotated area.
	Height = 8

	framebufferSize = (Width*Height + 7) / 8
)

type framebuffer struct{ pixels [framebufferSize]byte }

func (f *framebuffer) get(index int) bool {
	return f.pixels[index/8]&(byte(1)<<(index%8)) != 0
}

func (f *framebuffer) set(index int, on bool) {
	mask := byte(1) << (index % 8)
	if on {
		f.pixels[index/8] |= mask
	} else {
		f.pixels[index/8] &^= mask
	}
}

func (f *framebuffer) clear() { f.pixels = [framebufferSize]byte{} }
