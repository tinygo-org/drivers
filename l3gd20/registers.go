package l3gd20

// Expected identification number for L3GD20.
const (
	// For L3GD20
	expectedWHOAMI = 0xD4
	// For L3GD20H
	expectedWHOAMI_H = 0xD7
)

// Register bits masks
const (
	reg5RebootBit     = 1 << 7
	reg5FIFOEnableBit = 1 << 6
	reg1NormalBits    = 0b1111
)

// Register addresses. Comments from https://github.com/adafruit/Adafruit_L3GD20_U/blob/master/Adafruit_L3GD20_U.cpp
const (
	// The Slave ADdress (SAD) associated with the L3GD20 is 110101xb
	I2CAddr = 0b1101011
	// The SDO pin can be used to modify the less significant bit of the device address.
	I2CAddrSDOLow       = 0b1101010
	WHOAMI        uint8 = 0x0F
	// 	CTRL_REG1 (0x20)
	//    ====================================================================
	//    BIT  Symbol    Description                                   Default
	//    ---  ------    --------------------------------------------- -------
	//    7-6  DR1/0     Output data rate                                   00
	//    5-4  BW1/0     Bandwidth selection                                00
	//      3  PD        0 = Power-down mode, 1 = normal/sleep mode          0
	//      2  ZEN       Z-axis enable (0 = disabled, 1 = enabled)           1
	//      1  YEN       Y-axis enable (0 = disabled, 1 = enabled)           1
	//      0  XEN       X-axis enable (0 = disabled, 1 = enabled)           1
	CTRL_REG1 uint8 = 0x20
	// 	Set CTRL_REG2 (0x21)
	//    ====================================================================
	//    BIT  Symbol    Description                                   Default
	//    ---  ------    --------------------------------------------- -------
	//    5-4  HPM1/0    High-pass filter mode selection                    00
	//    3-0  HPCF3..0  High-pass filter cutoff frequency selection      0000
	CTRL_REG2 uint8 = 0x21
	// 	CTRL_REG3 (0x22)
	//    ====================================================================
	//    BIT  Symbol    Description                                   Default
	//    ---  ------    --------------------------------------------- -------
	//      7  I1_Int1   Interrupt enable on INT1 (0=disable,1=enable)       0
	//      6  I1_Boot   Boot status on INT1 (0=disable,1=enable)            0
	//      5  H-Lactive Interrupt active config on INT1 (0=high,1=low)      0
	//      4  PP_OD     Push-Pull/Open-Drain (0=PP, 1=OD)                   0
	//      3  I2_DRDY   Data ready on DRDY/INT2 (0=disable,1=enable)        0
	//      2  I2_WTM    FIFO wtrmrk int on DRDY/INT2 (0=dsbl,1=enbl)        0
	//      1  I2_ORun   FIFO overrun int on DRDY/INT2 (0=dsbl,1=enbl)       0
	//      0  I2_Empty  FIFI empty int on DRDY/INT2 (0=dsbl,1=enbl)         0
	CTRL_REG3 uint8 = 0x22
	// 	CTRL_REG4 (0x23)
	//    ====================================================================
	//    BIT  Symbol    Description                                   Default
	//    ---  ------    --------------------------------------------- -------
	//      7  BDU       Block Data Update (0=continuous, 1=LSB/MSB)         0
	//      6  BLE       Big/Little-Endian (0=Data LSB, 1=Data MSB)          0
	//    5-4  FS1/0     Full scale selection                               00
	//                                   00 = 250 dps
	//                                   01 = 500 dps
	//                                   10 = 2000 dps
	//                                   11 = 2000 dps
	//      0  SIM       SPI Mode (0=4-wire, 1=3-wire)                       0
	CTRL_REG4 uint8 = 0x23
	// 	CTRL_REG5 (0x24)
	//    ====================================================================
	//    BIT  Symbol    Description                                   Default
	//    ---  ------    --------------------------------------------- -------
	//      7  BOOT      Reboot memory content (0=normal, 1=reboot)          0
	//      6  FIFO_EN   FIFO enable (0=FIFO disable, 1=enable)              0
	//      4  HPen      High-pass filter enable (0=disable,1=enable)        0
	//    3-2  INT1_SEL  INT1 Selection config                              00
	//    1-0  OUT_SEL   Out selection config                               00
	CTRL_REG5     uint8 = 0x24
	REFERENCE     uint8 = 0x25
	OUT_TEMP      uint8 = 0x26
	STATUS_REG    uint8 = 0x27
	OUT_X_L       uint8 = 0x28
	OUT_X_H       uint8 = 0x29
	OUT_Y_L       uint8 = 0x2A
	OUT_Y_H       uint8 = 0x2B
	OUT_Z_L       uint8 = 0x2C
	OUT_Z_H       uint8 = 0x2D
	FIFO_CTRL_REG uint8 = 0x2E
	FIFO_SRC_REG  uint8 = 0x2F
	INT1_CFG      uint8 = 0x30
	INT1_SRC      uint8 = 0x31
	INT1_TSH_XH   uint8 = 0x32
	INT1_TSH_XL   uint8 = 0x33
	INT1_TSH_YH   uint8 = 0x34
	INT1_TSH_YL   uint8 = 0x35
	INT1_TSH_ZH   uint8 = 0x36
	INT1_TSH_ZL   uint8 = 0x37
	INT1_DURATION uint8 = 0x38
)
