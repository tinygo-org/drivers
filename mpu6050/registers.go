package mpu6050

// read reads register reg and writes n bytes to b where
// n is the length of b.
func (p *Dev) read(reg uint8, buff []byte) error {
	return p.conn.Tx(uint16(p.address), []byte{reg}, buff)
}

// write8 writes a registry byte
func (p *Dev) write8(reg uint8, datum byte) error {
	err := p.write(reg, []byte{datum})
	return err
}

// write write a byte buffer to sequential incrementing adresses (1 byte per address) in individual
// write operations.
func (p *Dev) write(addr uint8, buff []byte) error {
	p.conn.Tx(uint16(p.address), append([]byte{addr}, buff...), nil)
	return nil
}

// MPU 6050 REGISTER ADRESSES
const (
	SMPRT_DIV    uint8 = 0x19
	CONFIG       uint8 = 0x1A
	GYRO_CONFIG  uint8 = 0x1B
	ACCEL_CONFIG uint8 = 0x1C
	FIFO_EN      uint8 = 0x23
	I2C_MST_CTRL uint8 = 0x24
	PWR_MGMT_1   uint8 = 0x6B
	ACCEL_XOUT_H uint8 = 0x3B
	TEMP_OUT_H   uint8 = 0x41
	GYRO_XOUT_H  uint8 = 0x43
)

// MPU 6050 MASKS
const (
	G_FS_SEL     uint8 = 0x18
	AFS_SEL      uint8 = 0x18
	CLK_SEL_MASK uint8 = 0x07
	SLEEP_MASK   uint8 = 0x40
)

// MPU 6050 SHIFTS
const (
	AFS_SHIFT   uint8 = 3
	G_FS_SHIFT  uint8 = 3
	SLEEP_SHIFT uint8 = 6
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
