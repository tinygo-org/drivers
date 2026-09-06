package imu

import (
	"errors"
	"main/mpu6050"

	m "math"

	"tinygo.org/x/drivers"
)

type IMU struct {
	mpu mpu6050.Device
}

func New(bus drivers.I2C) (IMU, error) {
	imu := IMU{}
	imu.mpu = mpu6050.New(bus)
	// TODO: Setup I2C to talk to MPU6050 at 400kHz

	err := imu.mpu.Initialize()
	if err != nil {
		println("Bad MPU6050 device status.")
		return imu, err
	}

	err = imu.mpu.TestConnection()
	if err != nil {
		println("Test for MPU6050 failed.")
		return imu, err
	}

	err = imu.mpu.DMPinitialize()
	if err != nil {
		println("Initialize Digital Motion Processor (DMP) failed.")
		return imu, err
	}
	imu.GuessOffsets()

	println("MPU6050 connected!")
	return imu, nil
}

// Init initializes DMP on the mpu6050
// Note: New() must be called before Init()
func (imu *IMU) Init() error {
	if err := imu.mpu.SetDMPenabled(true); err != nil {
		println("Failed to enable Digital Motion Processor.")
		return err
	}
	return nil
}

// Note: New() must be called before this
func (imu *IMU) load_offsets_from_storage_and_start_dmp() error {
	if imu.IsCalibrated() {
		err := imu.loadCalibration()
		if err != nil {
			println("MPU6050 not calibrated.")
			return err
		}
		err = imu.mpu.SetDMPenabled(true)
		if err != nil {
			println("Failed to enable Digital Motion Processor.")
			return err
		}
	} else {
		println("MPU not calibrated.")
		return errors.New("MPU is not calibrated")
	}
	return nil
}
func (imu *IMU) IsCalibrated() bool {
	// TODO: read calibration flag from storage
	return true
}

func (imu *IMU) Calibrate() error {
	println("Calibration invoked ...")
	imu.mpu.CalibrateAccel(6)
	imu.mpu.CalibrateGyro(6)

	// TODO: store calibation to EEPROM or FLASH for
	// retieval on restarts.
	return nil
}

func (imu *IMU) PrintIMUOffsets() {
	ax, _ := imu.mpu.GetXAccelOffset()
	ay, _ := imu.mpu.GetYAccelOffset()
	az, _ := imu.mpu.GetZAccelOffset()
	gx, _ := imu.mpu.GetXGyroOffset()
	gy, _ := imu.mpu.GetYGyroOffset()
	gz, _ := imu.mpu.GetZGyroOffset()

	println("Accel X: ", ax,
		"\tAccel Y: ", ay,
		"\tAccel Z: ", az,
		"\tGyro X: ", gx,
		"\tGyro Y: ", gy,
		"\tGyro Z: ", gz,
	)

}

// use loadCalibration to get the calibration from
// EEPROM or FLASH storage - see Maker's Wharf for inspirational work.
func (imu *IMU) loadCalibration() error {
	// TODO: get calibration from storage
	return nil
}

func (imu *IMU) GetYawPitchRoll() (mpu6050.Angels, error) {
	fifo_buffer := make([]byte, 64)
	angle := mpu6050.Angels{}

	err := imu.mpu.DMPgetCurrentFIFOPacket(fifo_buffer)
	if err != nil {
		println("DMP buffer unavailable: ", err.Error())
		return angle, err
	}
	q, err := imu.mpu.DMPgetQuaternion(fifo_buffer)
	if err != nil {
		println("FIFO buffer to Quaternion conversion error")
		return angle, err
	}
	//print(q.W, "\t", q.X, "\t", q.Y, "\t", q.Z, "\n")
	q0 := float64(q.W)
	q1 := float64(q.X)
	q2 := float64(q.Y)
	q3 := float64(q.Z)

	// this conversion is based on Maker's Wharf corrected conversion
	// that does not have a gimbal lock  - check the video
	yr := -m.Atan2(-2.0*q1*q2+2.0*q0*q3, q2*q2-q3*q3-q1*q1+q0*q0)
	pr := m.Asin(2.0*q2*q3 + 2.0*q0*q1)
	rr := m.Atan2(-2.0*q1*q3+2.0*q0*q2, q3*q3-q2*q2-q1*q1+q0*q0)

	// convert to radians
	angle.Yaw = float32(yr * 180.0 / m.Pi)
	angle.Pitch = float32(pr * 180.0 / m.Pi)
	angle.Roll = float32(rr * 180.0 / m.Pi)

	return angle, nil
}

func (imu *IMU) GuessOffsets() {
	println("Apply Guess OFFSETS .....")
	/* Supply your gyro offsets here, scaled for min sensitivity */

	imu.mpu.SetXGyroOffset(294)
	imu.mpu.SetYGyroOffset(-24)
	imu.mpu.SetZGyroOffset(19)
	imu.mpu.SetXAccelOffset(-524)
	imu.mpu.SetYAccelOffset(-2261)
	imu.mpu.SetZAccelOffset(620)
}

// Use basic (NOT DMP mode) MPU functions to see if the I2C comms works
func (imu *IMU) Test_MPU6050() {
	imu.mpu.Initialize()
	x, y, z := imu.mpu.ReadRotation()
	println("Rotation x:", x, "  y:", y, " z:", z)

	x, y, z = imu.mpu.ReadAcceleration()
	println("Acceleration x:", x, "  y:", y, " z:", z)
}
