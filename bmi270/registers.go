package bmi270

const (
	reg_CHIP_ID         = 0x00
	reg_ACC_DATA        = 0x0C
	reg_GYR_DATA        = 0x12
	reg_INTERNAL_STATUS = 0x21
	reg_ACC_CONF        = 0x40
	reg_ACC_RANGE       = 0x41
	reg_GYR_CONF        = 0x42
	reg_GYR_RANGE       = 0x43
	reg_INIT_CTRL       = 0x59
	reg_INIT_ADDR_0     = 0x5B
	reg_INIT_ADDR_1     = 0x5C
	reg_INIT_DATA       = 0x5E
	reg_PWR_CONF        = 0x7C
	reg_PWR_CTRL        = 0x7D
	reg_CMD             = 0x7E

	chipIDBMI270 = 0x24
)
