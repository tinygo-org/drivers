package sdcard

import "errors"

var (
	errNoCard            = errors.New("sdcard: no card")
	errCardV1Unsupported = errors.New("sdcard: v1 cards not supported")
	errCMD8Voltage       = errors.New("sdcard: CMD8 bad voltage range")
	errCMD8Pattern       = errors.New("sdcard: CMD8 bad check pattern")
	errACMD41            = errors.New("sdcard: ACMD41 failed")
	errCMD58             = errors.New("sdcard: CMD58 failed")
	errCMD16             = errors.New("sdcard: CMD16 failed")
	errStartBlock        = errors.New("sdcard: start block token not received")
	errReadReg           = errors.New("sdcard: register read failed")
	errBufferTooSmall    = errors.New("sdcard: buffer must be at least 512 bytes")
	errCMD17             = errors.New("sdcard: CMD17 failed")
	errCMD24             = errors.New("sdcard: CMD24 failed")
	errCMD25             = errors.New("sdcard: CMD25 failed")
	errWrite             = errors.New("sdcard: write rejected")
	errWriteTimeout      = errors.New("sdcard: write timeout")
	errCSDv1Unsupported  = errors.New("sdcard: CSD version 1.0 not supported")
	errUnknownCSD        = errors.New("sdcard: unknown CSD format")
)
