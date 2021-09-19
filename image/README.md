# tinygo.org/x/drivers/image

This is an image package that uses less RAM to run on a microcontroller.
Unlike Go's original image package, `image.Decode()` does not return `image.Image`.

Instead, a callback can be set to process the data corresponding to the image.

## How to use

First, use `SetCallback()` to set the callback.
Then call `png.Decode()` or `jpeg.Decode()`.
The callback will be called as many times as necessary to load the image.

`SetCallback()` needs to be given a Buffer to handle the callback and the actual function to be called.
The `data []uint16` in the callback is in RGB565 format.

The `io.Reader` to pass to `Decode()` specifies the binary data of the image.

```go
func drawPng(display *ili9341.Device) error {
	p := strings.NewReader(pngImage)
	png.SetCallback(buffer[:], func(data []uint16, x, y, w, h, width, height int16) {
		err := display.DrawRGBBitmap(x, y, data[:w*h], w, h)
		if err != nil {
			errorMessage(fmt.Errorf("error drawPng: %s", err))
		}
	})

	return png.Decode(p)
}
```

```go
func drawJpeg(display *ili9341.Device) error {
	p := strings.NewReader(jpegImage)
	jpeg.SetCallback(buffer[:], func(data []uint16, x, y, w, h, width, height int16) {
		err := display.DrawRGBBitmap(x, y, data[:w*h], w, h)
		if err != nil {
			errorMessage(fmt.Errorf("error drawJpeg: %s", err))
		}
	})

	return jpeg.Decode(p)
}
```

## How to create an image

The following program will output an image binary like the one in [images.go](./examples/ili9341/slideshow/images.go).  

```
go run ./cmd/convert2bin ./path/to/png_or_jpg.png
```

## Examples

An example can be found below.
Processing jpegs requires a minimum of 32KB of RAM.

* [./examples/ili9341/slideshow](./examples/ili9341/slideshow)
