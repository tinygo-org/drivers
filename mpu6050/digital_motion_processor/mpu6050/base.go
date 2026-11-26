// Read and write to registers/address based on Jeff's utility functions
// These are very similar except for the fact that the length(size) of
// is implied by the length of the []byte. This works very well with
// the I2C read and write design in Tinygo device modules

// Note: not all of Jeff Rowberg's utilities are implemented - Only those
// required for the DMP configuration

package mpu6050

import (
	"encoding/binary"
	"errors"
	"strconv"
)

// WriteBit writes 0 or 1 to bit position 0-7
func (d Device) WriteBit(regAddr uint8, bitNum uint8, data uint8) error {
	b := make([]byte, 1)
	d.ReadBytes(regAddr, b)
	if data != 0 {
		b[0] |= (0x01 << bitNum)
	} else {
		b[0] &= 0xFF & ^(0x01 << bitNum)
	}
	return d.WriteBytes(regAddr, b)
}

// WriteBits writes a set of bits of length at starting position bitStart
//
//	010 value to write
//
// 76543210 bit numbers
//
//	xxx   args: bitStart=4, length=3
//
// 00011100 mask byte
// 10101111 original value (sample)
// 10100011 original & ~mask
// 10101011 masked | value
func (d Device) WriteBits(regAddr, bitStart, length, data uint8) error {
	b := []byte{0}
	err := d.ReadBytes(regAddr, b)
	if err != nil {
		return err
	}
	var mask uint8
	mask = ((1 << length) - 1) << (bitStart - length + 1)
	data = data << (bitStart - length + 1) // shift data into correct position
	data &= mask                           // zero all non-important bits in data
	b[0] &= ^(mask)                        // zero all important bits in existing byte
	b[0] |= data                           // combine data with existing byte
	return d.WriteBytes(regAddr, b)
}

// ReadBits reads a set of bits of length at starting position bitStart
//
// 01101001 read byte
// 76543210 bit numbers
//
//	 xxx   args: bitStart=4, length=3
//	 010   masked
//	-> 010 shifted
func (d Device) ReadBits(regAddr, bitStart, length uint8) (uint8, error) {
	buf := []byte{0}
	err := d.ReadBytes(regAddr, buf)
	if err != nil {
		return buf[0], err
	}
	var mask uint8
	mask = ((1 << length) - 1) << (bitStart - length + 1)
	buf[0] &= mask
	buf[0] = buf[0] >> (bitStart - length + 1)

	return buf[0], nil
}

// ReadBytes reads bytes from the I2C device for the length of data
func (d Device) ReadBytes(reg uint8, data []byte) error {
	return d.bus.Tx(uint16(d.Address), []byte{reg}, data)
}

// WriteBytes writes bytes to the I2C device for the length of data
func (d Device) WriteBytes(reg uint8, data []byte) error {
	buf := make([]uint8, len(data)+1)
	buf[0] = reg
	copy(buf[1:], data)
	return d.bus.Tx(uint16(d.Address), buf, nil)
}

// ReadWords reads int16 Arrays from the registers
func (d Device) ReadWords(reg uint8, data []int16) error {
	b := make([]byte, len(data)*2)
	err := d.ReadBytes(reg, b)
	if err != nil {
		return err
	}
	for i := 0; i < len(data); i++ {
		data[i] = int16(binary.BigEndian.Uint16(b[i*2:]))
		//data[i] = int16(b[2*i])<<8 | int16(b[2*i+1])
	}
	return nil
}

// WriteWords writes int16 Arrays to the registers
func (d *Device) WriteWords(regAddr uint8, data []int16) error {
	b := make([]byte, len(data)*2)
	for i := 0; i < len(data); i++ {
		binary.BigEndian.PutUint16(b[i*2:], uint16(data[i]))
		//b[i*2] = byte((data[i] >> 8) & 0xFF)
		//b[i*2+1] = byte(data[i] & 0xFF)
	}
	return d.WriteBytes(regAddr, b)
}

// Alla Jeff Rowberg style - better than mine!
func (d Device) writeMemoryBlock(data []byte, dataSize int, bank uint8, address uint8, verify bool) error {
	// safety check - force user to specify length of data[]
	if int(dataSize) != len(data) {
		err_msg := "Size of datablock is invalid"
		return errors.New(err_msg)
	}
	d.setMemoryBank(bank, false, false)
	d.setMemoryStartAddress(address)
	for i := 0; i < dataSize; {

		chunkSize := DMP_MEMORY_CHUNK_SIZE

		if i+chunkSize > dataSize {
			chunkSize = dataSize - i
		}

		if chunkSize > 256-int(address) {
			chunkSize = 256 - int(address)
		}

		if err := d.WriteBytes(MEM_R_W, data[i:i+chunkSize]); err != nil {
			return err
		}

		if verify {
			d.setMemoryBank(bank, false, false)
			d.setMemoryStartAddress(address)
			buf := make([]byte, chunkSize)
			if err := d.ReadBytes(MEM_R_W, buf); err != nil {
				return err
			}
			for ii, b := range data[i : i+chunkSize] {
				if b != buf[ii] {
					// dump the buffer
					println("length: ", len(buf))
					for j := 0; j < len(buf); j++ {
						print(strconv.FormatInt(int64(buf[j]), 16))
					}
					println()
					// dump the data
					for j := 0; j < chunkSize; j++ {
						print(strconv.FormatInt(int64(data[i+j]), 16))
					}
					println()
					return errors.New("Verify writeMemoryBlock failed")
				}
			}
		}
		i += chunkSize
		address += uint8(chunkSize) // address will wrap automatically at 256 - swith to a new bank and restart a 0 offset
		if i < dataSize {
			if address == 0 {
				bank++
			}
			d.setMemoryBank(bank, false, false)
			d.setMemoryStartAddress(address)
		}
	}
	return nil
}

/* ------ MY version of the above -----
// writeMemoryBlock -dumps []byte to the MPU6050 Read and Write address - the MPU6050 will take the data and move it to
// the bank and address as indicated
func (d Device) writeMemoryBlock(data []byte, dataSize uint16, bank uint8, address uint8, verify bool) error {
	d.setMemoryBank(bank, false, false)
	d.setMemoryStartAddress(address)
	if int(dataSize) != len(data) {
		err_msg := "Size of datablock is invalid"
		return errors.New(err_msg)
	}
	n := dataSize
	for n > 0 {
		index := dataSize - n
		if n >= DMP_MEMORY_CHUNK_SIZE {
			if err := d.WriteBytes(MEM_R_W, data[index:index+DMP_MEMORY_CHUNK_SIZE]); err != nil {
				return err
			}
			time.Sleep(time.Millisecond * 30)
			if verify {
				d.setMemoryBank(bank, false, false)
				d.setMemoryStartAddress(address)
				buf := make([]byte, DMP_MEMORY_CHUNK_SIZE)
				if err := d.ReadBytes(MEM_R_W, buf); err != nil {
					return err
				}
				for i, b := range data[index : index+DMP_MEMORY_CHUNK_SIZE] {
					if b != buf[i] {
						// dump the buffer
						println("length: ", len(buf))
						for i := 0; i < len(buf); i++ {
							print(strconv.FormatInt(int64(buf[i]), 16))
						}
						println()
						// dump the data
						for i := 0; i < DMP_MEMORY_CHUNK_SIZE; i++ {
							print(strconv.FormatInt(int64(data[int(index)+i]), 16))
						}
						println()
						return errors.New("Verify writeMemoryBlock failed")
					}
				}
			}
			n -= DMP_MEMORY_CHUNK_SIZE
			address += DMP_MEMORY_CHUNK_SIZE // uint8 automatically wraps to 0 at 256

			if address == 0 { // Increase bank ie next 256 chunk
				bank++
			}
			if err := d.setMemoryBank(bank, false, false); err != nil {
				return err
			}
			if err := d.setMemoryStartAddress(address); err != nil {
				return err
			}
			print(".")
		} else {
			if err := d.WriteBytes(MEM_R_W, data[index:index+n]); err != nil {
				return err
			}
			if verify {
				d.setMemoryBank(bank, false, false)
				d.setMemoryStartAddress(address)
				buf := make([]byte, n)
				if err := d.ReadBytes(MEM_R_W, buf); err != nil {
					return err
				}
				for i, b := range data[index : index+n] {
					if b != buf[i] {
						// dump the buffer
						println("length: ", len(buf))
						for i := 0; i < len(buf); i++ {
							print(strconv.FormatInt(int64(buf[i]), 16))
						}
						println()
						// dump the data
						for i := 0; i < int(n); i++ {
							print(strconv.FormatInt(int64(data[int(index)+i]), 16))
						}
						println()
						return errors.New("Verify writeMemoryBlock failed")
					}
				}
			}
			n = 0
		}
	}
	return nil
}
*/
