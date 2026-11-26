// This module owes its design to Jeff Rowberg
// https://github.com/ElectronicCats/mpu6050

// It is a stripped down  and minimilistic version for DMP purposes
// as used by Maker's Wharf

// The DMP module and functions was exported and translated from the C/C++
// implementation by Jeff Rowberg ( Electronic Cats) and
// converted for Tinygo use as an extened mpu6050 device (see Tinygo devices).

// The work was inspired by this video from Maker's Wharf
// https://www.youtube.com/watch?v=k5i-vE5rZR0

// Tested with the Arduino-Zero Tinygo machine (SAMD21)

package mpu6050

import (
	"errors"
	"strconv"
	"time"
)

var (
	print_debug_message = false
)

/** Power on and prepare for general usage.
 * This will activate the device and take it out of sleep mode (which must be done
 * after start-up). This function also sets both the accelerometer and the gyroscope
 * to their most sensitive settings, namely +/- 2g and +/- 250 degrees/sec, and sets
 * the clock source to use the X Gyro for reference, which is slightly better than
 * the default internal clock source.
 */
func (d Device) Initialize() error {
	err := d.SetClockSource(CLOCK_PLL_XGYRO)
	if err != nil {
		return err
	}

	err = d.SetFullScaleGyroRange(GYRO_FS_250)
	if err != nil {
		return err
	}

	err = d.SetFullScaleAccelRange(AFS_RANGE_2G)
	if err != nil {
		return err
	}

	err = d.SetSleepEnabled(false)
	if err != nil {
		return err
	}
	return nil
}

// SetSleepEnabled - toggle the mpu6050 sleep mode (0x00 == OFF, 0x01 == SLEEP MODE ON)
func (d Device) SetSleepEnabled(s bool) error {
	if s == true {
		return d.WriteBit(PWR_MGMT_1, PWR_MGMT_1_SLEEP_BIT, 0x01)
	} else {
		return d.WriteBit(PWR_MGMT_1, PWR_MGMT_1_SLEEP_BIT, 0x00)
	}
}

func (d Device) SetXGyroOffset(offset int16) error {
	return d.WriteWords(XG_OFFS_USRH, []int16{offset})
}

func (d Device) GetXGyroOffset() (int16, error) {
	buf := make([]int16, 1)
	if err := d.ReadWords(XG_OFFS_USRH, buf); err != nil {
		return buf[0], err
	}
	return buf[0], nil
}

func (d Device) SetYGyroOffset(offset int16) error {
	return d.WriteWords(YG_OFFS_USRH, []int16{offset})
}

func (d Device) GetYGyroOffset() (int16, error) {
	buf := make([]int16, 1)
	if err := d.ReadWords(YG_OFFS_USRH, buf); err != nil {
		return buf[0], err
	}
	return buf[0], nil
}

func (d Device) SetZGyroOffset(offset int16) error {
	return d.WriteWords(ZG_OFFS_USRH, []int16{offset})
}

func (d Device) GetZGyroOffset() (int16, error) {
	buf := make([]int16, 1)
	if err := d.ReadWords(ZG_OFFS_USRH, buf); err != nil {
		return buf[0], err
	}
	return buf[0], nil
}

func (d Device) SetXAccelOffset(offset int16) error {
	var saveAddress uint8
	id, _ := d.getDeviceID()
	if id < 0x38 {
		saveAddress = XA_OFFS_H
	} else {
		saveAddress = 0x77
	}
	return d.WriteWords(saveAddress, []int16{offset})
}

func (d Device) GetXAccelOffset() (int16, error) {
	var saveAddress uint8
	buf := make([]int16, 1)
	id, _ := d.getDeviceID()
	if id < 0x38 {
		saveAddress = XA_OFFS_H
	} else {
		saveAddress = 0x77
	}
	if err := d.ReadWords(saveAddress, buf); err != nil {
		return buf[0], err
	}
	return buf[0], nil
}

func (d Device) SetYAccelOffset(offset int16) error {
	var saveAddress uint8
	id, _ := d.getDeviceID()
	if id < 0x38 {
		saveAddress = YA_OFFS_H
	} else {
		saveAddress = 0x7A
	}
	return d.WriteWords(saveAddress, []int16{offset})
}

func (d Device) GetYAccelOffset() (int16, error) {
	var saveAddress uint8
	buf := make([]int16, 1)
	id, _ := d.getDeviceID()
	if id < 0x38 {
		saveAddress = YA_OFFS_H
	} else {
		saveAddress = 0x7A
	}
	if err := d.ReadWords(saveAddress, buf); err != nil {
		return buf[0], err
	}
	return buf[0], nil
}

func (d Device) SetZAccelOffset(offset int16) error {
	var saveAddress uint8
	id, _ := d.getDeviceID()
	if id < 0x38 {
		saveAddress = ZA_OFFS_H
	} else {
		saveAddress = 0x7D
	}
	return d.WriteWords(saveAddress, []int16{offset})
}

func (d Device) GetZAccelOffset() (int16, error) {
	var saveAddress uint8
	buf := make([]int16, 1)
	id, _ := d.getDeviceID()
	if id < 0x38 {
		saveAddress = ZA_OFFS_H
	} else {
		saveAddress = 0x7D
	}
	if err := d.ReadWords(saveAddress, buf); err != nil {
		return buf[0], err
	}
	return buf[0], nil
}

func (d Device) TestConnection() error {
	if !d.Connected() {
		return errors.New("Connection test for MPU6050 failed.")
	}
	return nil
}

func debug_print(str string) {
	if !print_debug_message {
		return
	}
	println(str)
}

func (d Device) DMPinitialize() error {
	// reset device
	debug_print("\n\nResetting MPU6050...\n")
	d.reset()
	time.Sleep(time.Millisecond * 30)

	// disable sleep mode
	d.SetSleepEnabled(false)

	// get MPU hardware revision
	d.setMemoryBank(0x10, true, true)
	d.setMemoryStartAddress(0x06)
	debug_print("Checking hardware revision...")
	debug_print("Revision @ user[16][6] = " + strconv.Itoa(int(d.readMemoryByte())))
	debug_print("Resetting memory bank selection to 0...")
	d.setMemoryBank(0, false, false)

	// check OTP bank valid
	debug_print("Reading OTP bank valid flag...")
	if d.getOTPBankValid() {
		debug_print("OTP bank is valid!")
	} else {
		debug_print("OTP bank is invalid!")
	}

	// setup weird slave stuff (?)
	debug_print("Setting slave 0 address to 0x7F...")
	d.setSlaveAddress(0, 0x7F)
	debug_print("Disabling I2C Master mode...")
	d.setI2CMasterModeEnabled(false)
	debug_print("Setting slave 0 address to 0x68 (self)...")
	d.setSlaveAddress(0, 0x68)
	debug_print("Resetting I2C Master control...")
	d.resetI2CMaster()
	time.Sleep(time.Millisecond * 20)

	debug_print("Setting clock source to Z Gyro...")
	d.setClockSource(CLOCK_PLL_ZGYRO)

	debug_print("Setting DMP and FIFO_OFLOW interrupts enabled...")
	d.setIntEnabled(1<<INTERRUPT_FIFO_OFLOW_BIT | 1<<INTERRUPT_DMP_INT_BIT)

	debug_print("Setting sample rate to 200Hz...")
	d.setRate(4) // 1khz / (1 + 4) = 200 Hz

	debug_print("Setting external frame sync to TEMP_OUT_L[0]...")
	d.setExternalFrameSync(SYNC_TEMP_OUT_L)

	debug_print("Setting DLPF bandwidth to 42Hz...")
	d.setDLPFMode(DLPF_BW_42)

	debug_print("Setting gyro sensitivity to +/- 2000 deg/sec...")
	d.setFullScaleGyroRange(GYRO_FS_2000)

	// load DMP code into memory banks
	debug_print("Writing DMP code to MPU memory banks (" + strconv.Itoa(DMP_CODE_SIZE) + "bytes)")
	//if (!writeProgMemoryBlock(dmpMemory, MPU6050_DMP_CODE_SIZE)) return 1; // Failed
	if err := d.writeMemoryBlock(PROGMEM, DMP_CODE_SIZE, 0x00, 0x00, true); err != nil {
		println("Failed write DMP firmware: ", err.Error())
		return err // Failed
	}
	debug_print("Success! DMP code written and verified.")

	// Set the FIFO Rate Divisor into the DMP Firmware Memory
	dmpUpdate := []byte{0x00, DMP_FIFO_RATE_DIVISOR}
	d.writeMemoryBlock(dmpUpdate, 0x02, 0x02, 0x16, true) // Lets write the dmpUpdate data to the Firmware image, we have 2 bytes to write in bank 0x02 with the Offset 0x16

	//write start address MSB into register
	d.setDMPConfig1(0x03)
	//write start address LSB into register
	d.setDMPConfig2(0x00)

	debug_print("Clearing OTP Bank flag...")
	d.setOTPBankValid(false)

	debug_print("Setting motion detection threshold to 2...")
	d.setMotionDetectionThreshold(2)

	debug_print("Setting zero-motion detection threshold to 156...")
	d.setZeroMotionDetectionThreshold(156)

	debug_print("Setting motion detection duration to 80...")
	d.setMotionDetectionDuration(80)

	debug_print("Setting zero-motion detection duration to 0...")
	d.setZeroMotionDetectionDuration(0)

	debug_print("Enabling FIFO...")
	d.setFIFOEnabled(true)

	debug_print("Resetting DMP...")
	d.resetDMP()

	debug_print("DMP is good to go! Finally.")

	debug_print("Disabling DMP (you turn it on later)...")
	d.setDMPEnabled(false)

	debug_print("Setting up internal 42-byte (default) DMP packet buffer...")
	d.dmpPacketSize = 42

	debug_print("Resetting FIFO and clearing INT status one last time...")
	d.resetFIFO()
	init_status := d.getIntStatus()
	debug_print("DMP initialize status: " + strconv.Itoa(int(init_status)))

	return nil
}

func (d Device) reset() error {
	return d.WriteBit(PWR_MGMT_1, PWR_MGMT_1_DEVICE_RESET_BIT, 0x01)
}

func (d Device) setMemoryBank(bank uint8, prefetchEnabled, userBank bool) error {
	bank &= 0x1F
	if userBank {
		bank |= 0x20
	}
	if prefetchEnabled {
		bank |= 0x40
	}
	return d.WriteBytes(BANK_SEL, []byte{bank})
}

func (d Device) setMemoryStartAddress(address uint8) error {
	return d.WriteBytes(MEM_START_ADDR, []byte{address})
}

func (d Device) readMemoryByte() uint8 {
	buf := []byte{0}
	d.ReadBytes(MEM_R_W, buf)
	return buf[0]
}

func (d Device) getOTPBankValid() bool {
	buf := []byte{0}
	d.ReadBytes(XG_OFFS_TC, buf)
	return (buf[0] & (0x01 << TC_OTP_BNK_VLD_BIT)) > 0
}

func (d Device) setSlaveAddress(num uint8, address uint8) {
	if num > 3 {
		return
	}
	d.WriteBytes(I2C_SLV0_ADDR+num*3, []byte{address})
}

func (d Device) setI2CMasterModeEnabled(s bool) {
	if s == true {
		d.WriteBit(USER_CTRL, USER_CTRL_I2C_MST_EN_BIT, 0x01)
	} else {
		d.WriteBit(USER_CTRL, USER_CTRL_I2C_MST_EN_BIT, 0x00)
	}
}

func (d Device) resetI2CMaster() {
	d.WriteBit(USER_CTRL, USER_CTRL_I2C_MST_RESET_BIT, 0x01)
}

func (d Device) setClockSource(source uint8) error {
	return d.WriteBits(PWR_MGMT_1, PWR_MGMT_1_CLKSEL_BIT, PWR_MGMT_1_CLKSEL_LENGTH, source)
}

func (d Device) setIntEnabled(interrupt uint8) error {
	return d.WriteBytes(INT_ENABLE, []byte{interrupt})
}

func (d Device) setRate(rate uint8) error {
	return d.WriteBytes(SMPLRT_DIV, []byte{rate})
}

func (d Device) setExternalFrameSync(sync uint8) error {
	return d.WriteBits(CONFIG, CONFIG_EXT_SYNC_SET_BIT, CONFIG_EXT_SYNC_SET_LENGTH, sync)
}

func (d Device) setDLPFMode(mode uint8) error {
	return d.WriteBits(CONFIG, CONFIG_DLPF_CFG_BIT, CONFIG_DLPF_CFG_LENGTH, mode)
}

func (d Device) setFullScaleGyroRange(rnge uint8) error {
	return d.WriteBits(GYRO_CONFIG, GYRO_CONFIG_FS_SEL_BIT, GYRO_CONFIG_FS_SEL_LENGTH, rnge)
}

func (d Device) setDMPConfig1(config uint8) error {
	return d.WriteBytes(DMP_CFG_1, []byte{config})
}

func (d Device) setDMPConfig2(config uint8) error {
	return d.WriteBytes(DMP_CFG_2, []byte{config})
}

func (d Device) setOTPBankValid(s bool) error {
	if s == true {
		return d.WriteBit(XG_OFFS_TC, TC_OTP_BNK_VLD_BIT, 0x01) // 0x01 == set enabled
	} else {
		return d.WriteBit(XG_OFFS_TC, TC_OTP_BNK_VLD_BIT, 0x00) // 0x00 == set disabled
	}
}

func (d Device) setMotionDetectionThreshold(threshold uint8) error {
	return d.WriteBytes(MOT_THR, []byte{threshold})
}

func (d Device) setZeroMotionDetectionThreshold(threshold uint8) error {
	return d.WriteBytes(ZRMOT_THR, []byte{threshold})
}

func (d Device) setMotionDetectionDuration(duration uint8) error {
	return d.WriteBytes(MOT_DUR, []byte{duration})
}

func (d Device) setZeroMotionDetectionDuration(duration uint8) error {
	return d.WriteBytes(ZRMOT_DUR, []byte{duration})
}

func (d Device) setFIFOEnabled(s bool) error {
	if s == true {
		return d.WriteBit(USER_CTRL, USER_CTRL_FIFO_EN_BIT, 0x01) // 0x01 enabled
	} else {
		return d.WriteBit(USER_CTRL, USER_CTRL_FIFO_EN_BIT, 0x00) // 0x00 disabled
	}
}

func (d Device) resetDMP() error {
	return d.WriteBit(USER_CTRL, USER_CTRL_DMP_RESET_BIT, 0x01)
}

func (d Device) setDMPEnabled(s bool) error {
	if s == true {
		return d.WriteBit(USER_CTRL, USER_CTRL_DMP_EN_BIT, 0x01) // 0x01 set enablrd
	} else {
		return d.WriteBit(USER_CTRL, USER_CTRL_DMP_EN_BIT, 0x00) // 0x00 set disablrd
	}
}

func (d Device) resetFIFO() error {
	return d.WriteBit(USER_CTRL, USER_CTRL_FIFO_RESET_BIT, 0x01)
}

func (d Device) getIntStatus() uint8 {
	buf := []byte{0}
	d.ReadBytes(INT_STATUS, buf)
	return buf[0]
}

func (d Device) SetDMPenabled(s bool) error {
	if s == true {
		if err := d.WriteBit(USER_CTRL, USER_CTRL_DMP_EN_BIT, 0x01); err != nil {
			return err
		}
	} else {
		if err := d.WriteBit(USER_CTRL, USER_CTRL_DMP_EN_BIT, 0x00); err != nil {
			return err
		}
	}
	return nil
}

// the DMP is dumping Quaternions as fast as it can.
// Maybe faster than we would like. So try and get a relative
// fresh complete packet from the FIFO buffer.
func (d Device) getGurrentFIFOPacket(buf []byte, length int16) {
	var fifoC int16
	packetReceived := false

	iGetFIFOCount := func() int16 {
		fifoC = int16(d.getFIFOCount())
		return fifoC
	}

	for !packetReceived {
		if iGetFIFOCount() > length {
			if fifoC > 5*length {
				d.resetFIFO()
				time.Sleep(time.Millisecond)
				// try to strip fragmented head
				if iGetFIFOCount()%length != 0 {
					trash := make([]byte, fifoC%length)
					d.ReadBytes(FIFO_R_W, trash)
				}
				continue
			}
			if fifoC%length == 0 {
				err := d.getFIFOBytes(buf, uint8(length))
				if err == nil {
					packetReceived = true
				}
			}
		}
	}
}

func (d Device) getFIFOTimeout() int64 {
	return int64(FIFO_DEFAULT_TIMEOUT)
}

func (d Device) getFIFOCount() uint16 {
	buf := make([]byte, 2)
	d.ReadBytes(FIFO_COUNTH, buf)
	return (uint16(buf[0]) << 8) | uint16(buf[1])
}

func (d Device) getFIFOBytes(buf []byte, length uint8) error {
	tbuf := make([]byte, length)
	err := d.ReadBytes(FIFO_R_W, tbuf)
	if err != nil {
		return err
	}
	copy(buf, tbuf)
	return nil
}

func (d Device) DMPgetCurrentFIFOPacket(buf []byte) error {
	dmpPacketSize := int16(d.dmpPacketSize)
	if len(buf) < int(dmpPacketSize) {
		return errors.New("Buffer size is too small")
	}
	d.getGurrentFIFOPacket(buf, dmpPacketSize)

	// dump the buffer
	//for i := 0; i < int(dmpPacketSize); i++ {
	//	print(strconv.FormatInt(int64(buf[i]), 16))
	//}
	//println()
	return nil
}

type Quaternion struct {
	W float32
	X float32
	Y float32
	Z float32
}

type Angels struct {
	Yaw   float32
	Pitch float32
	Roll  float32
}

func (d Device) DMPgetQuaternion(buf []byte) (Quaternion, error) {
	data := make([]int16, 4)
	q := Quaternion{}
	data[0] = int16(uint16(buf[0])<<8 | uint16(buf[1]))
	data[1] = int16(uint16(buf[4])<<8 | uint16(buf[5]))
	data[2] = int16(uint16(buf[8])<<8 | uint16(buf[9]))
	data[3] = int16(uint16(buf[12])<<8 | uint16(buf[13]))

	q.W = float32(data[0]) / 16384.0
	q.X = float32(data[1]) / 16384.0
	q.Y = float32(data[2]) / 16384.0
	q.Z = float32(data[3]) / 16384.0

	// Not sure why this does not work - I guess we have to convert from uint to
	// int before we cast to float32 as in the case above
	//q.W = float32(binary.BigEndian.Uint32(buf[0:])) / 16384.0
	//q.X = float32(binary.BigEndian.Uint32(buf[4:])) / 16384.0
	//q.Y = float32(binary.BigEndian.Uint32(buf[8:])) / 16384.0
	//q.Z = float32(binary.BigEndian.Uint32(buf[12:])) / 16384.0
	return q, nil
}
