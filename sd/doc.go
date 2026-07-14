// Package sd implements SD card drivers and the SD card specification:
// CID and CSD register decoding, command definitions and CRC7/CRC16 checksums.
//
// The [SPICard] type drives an SD card over a SPI bus and implements the
// [Card] interface, which exposes block-aligned I/O. [BlockDevice] wraps any
// [Card] with byte-addressed [io.ReaderAt]/[io.WriterAt] implementations
// suitable for filesystem libraries such as tinyfs.
package sd
