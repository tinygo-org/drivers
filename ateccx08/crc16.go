// from https://github.com/usbarmory/armoryctl/blob/master/atecc608/atecc608.go#L104
// thank you!
package ateccx08

const (
	CRC16Poly uint16 = 0x8005
)

func crc16(data []byte) []byte {
	var crc uint16

	for i := 0; i < len(data); i++ {
		for shift := uint8(0x01); shift > 0x00; shift <<= 1 {
			// data and crc bits
			var d uint8
			var c uint8

			if uint8(data[i])&uint8(shift) != 0 {
				d = 1
			}

			c = uint8(crc >> 15)
			crc <<= 1

			if d != c {
				crc ^= CRC16Poly
			}
		}
	}

	return []byte{byte(crc & 0xff), byte(crc >> 8)}
}
