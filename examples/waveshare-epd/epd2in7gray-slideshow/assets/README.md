# Slideshow images

Add packed 2bpp `.bin` images here. Images are embedded at build time and
shown in filename order every 10 seconds. Numbered filenames control the order.

The format matches the monochrome slideshow's header, followed by 2bpp pixels.

| Offset | Content |
| --- | --- |
| 0 | Width as a little-endian uint16 |
| 2 | Height as a little-endian uint16 |
| 4 | Packed pixels in row-major order, most significant pixel first |

Each byte holds four pixels in bits 7-6, 5-4, 3-2, and 1-0.
Shade values are 0 black, 1 dark gray, 2 light gray, and 3 white.

The example uses Rotation270 and expects **264x176** landscape images.
Each file is 11,620 bytes, including the four-byte header. For example, the
header is `08 01 b0 00` and the pixel byte `1b` means black, dark gray,
light gray, white.

The supplied images are included in this order:

1. `00-macintosh-classic.bin` Macintosh Classic
2. `01-guinea-pig.bin` Guinea pig
3. `02-cccp-stamp.bin` Soviet space stamp
4. `03-kanagawa.bin` The Great Wave off Kanagawa

Images are fitted without cropping on a white canvas, then quantized to
four shades using Floyd-Steinberg dithering to retain photographic detail.
The Macintosh and guinea pig images have increased contrast.

With no `.bin` assets, the example reports an error before initializing or
clearing the panel.

Build and flash:

```sh
tinygo flash -target=waveshare-rp2040-pizero ./examples/waveshare-epd/epd2in7gray-slideshow
```

The SPI configuration and pins match the original `epd2in7` slideshow.
Initialization tracing is disabled in this example.
