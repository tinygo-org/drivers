package png

var (
	callback    Callback = func(data []uint16, x, y, w, h, width, height int16) {}
	callbackBuf []uint16
)

// A portion of the image data consisting of data, x, y, w, and h is passed to
// Callback. The size of the whole image is passed as width and height.
type Callback func(data []uint16, x, y, w, h, width, height int16)

// SetCallback registers the buffer and fn required for Callback. Callback can
// be called multiple times by calling Decode().
func SetCallback(buf []uint16, fn Callback) {
	callbackBuf = buf
	callback = fn
}
