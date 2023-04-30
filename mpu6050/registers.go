package mpu6050

// read reads register reg and writes n bytes to b where
// n is the length of b.
func (p *Device) read(reg uint8, buff []byte) error {
	return p.conn.Tx(uint16(p.address), []byte{reg}, buff)
}

// write8 writes a registry byte
func (p *Device) write8(reg uint8, datum byte) error {
	err := p.write(reg, []byte{datum})
	return err
}

// write write a byte buffer to sequential incrementing adresses (1 byte per address) in individual
// write operations.
func (p *Device) write(addr uint8, buff []byte) error {
	p.conn.Tx(uint16(p.address), append([]byte{addr}, buff...), nil)
	return nil
}

// MPU 6050 REGISTER ADRESSES
const (
	_SMPRT_DIV    uint8 = 0x19
	_CONFIG       uint8 = 0x1A
	_GYRO_CONFIG  uint8 = 0x1B
	_ACCEL_CONFIG uint8 = 0x1C
	_FIFO_EN      uint8 = 0x23
	_I2C_MST_CTRL uint8 = 0x24
	_PWR_MGMT_1   uint8 = 0x6B
	_ACCEL_XOUT_H uint8 = 0x3B
	_TEMP_OUT_H   uint8 = 0x41
	_GYRO_XOUT_H  uint8 = 0x43
)

// MPU 6050 MASKS
const (
	_G_FS_SEL     uint8 = 0x18
	_AFS_SEL      uint8 = 0x18
	_CLK_SEL_MASK uint8 = 0x07
	_SLEEP_MASK   uint8 = 0x40
)

// MPU 6050 SHIFTS
const (
	_AFS_SHIFT   uint8 = 3
	_G_FS_SHIFT  uint8 = 3
	_SLEEP_SHIFT uint8 = 6
)

// Gyroscope ranges for Init configuration
const (
	GYRO_RANGE_250  byte = 0x00
	GYRO_RANGE_500  byte = 0x01
	GYRO_RANGE_1000 byte = 0x02
	GYRO_RANGE_2000 byte = 0x03
)

// Accelerometer ranges for Init configuration
const (
	ACCEL_RANGE_2  byte = 0x00
	ACCEL_RANGE_4  byte = 0x01
	ACCEL_RANGE_8  byte = 0x02
	ACCEL_RANGE_16 byte = 0x03
)
