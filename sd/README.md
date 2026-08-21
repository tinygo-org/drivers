## `sd` package

File map:
* `blockdevice.go`: Contains logic for creating an `io.WriterAt` and `io.ReaderAt` with the `sd.BlockDevice` concrete type
     from the `sd.Card` interface which is intrinsically a blocked reader and writer. 

* `spicard.go`: Contains the `sd.SpiCard` driver for controlling an SD card over SPI using the most commonly available circuit boards. 

* `responses.go`: Contains a currently unused SD response implementations as per the latest specification.

* `definitions.go`: Contains SD Card specification definitions such as the CSD and CID types as well as encoding/decoding logic, as well as CRC logic.
