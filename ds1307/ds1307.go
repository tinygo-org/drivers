package ds1307

import (
	"time"

	"machine"
)

// Device wraps an I2C connection to a DS1307 device.
type Device struct {
	bus         machine.I2C
	Address     uint8
	AddressSRAM uint8
	data        []byte
}

// New creates a new DS1307 connection. I2C bus must be already configured.
func New(bus machine.I2C) Device {
	return Device{bus: bus,
		Address:     uint8(I2CAddress),
		AddressSRAM: SRAMBeginAddres,
		data:        make([]byte, 7),
	}
}

// SetTime sets the time and date
func (d *Device) SetTime(t time.Time) error {
	d.data[0] = decToBcd(t.Second())
	d.data[1] = decToBcd(t.Minute())
	d.data[2] = decToBcd(t.Hour())
	d.data[3] = decToBcd(int(t.Weekday() + 1))
	d.data[4] = decToBcd(t.Day())
	d.data[5] = decToBcd(int(t.Month()))
	d.data[6] = decToBcd(t.Year() - 2000)
	err := d.bus.WriteRegister(d.Address, uint8(TimeDate), d.data)
	return err
}

// Time returns the time and date
func (d *Device) Time() (time.Time, error) {
	err := d.bus.ReadRegister(d.Address, uint8(TimeDate), d.data)
	if err != nil {
		return time.Time{}, err
	}
	seconds := bcdToDec(d.data[0] & 0x7F)
	minute := bcdToDec(d.data[1])
	hour := hoursBCDToInt(d.data[2])
	day := bcdToDec(d.data[4])
	month := time.Month(bcdToDec(d.data[5]))
	year := bcdToDec(d.data[6])
	year += 2000

	t := time.Date(year, month, day, hour, minute, seconds, 0, time.UTC)
	return t, nil
}

// SetSRAMAddress sets SRAM register address. Range (SRAMBeginAddres, SRAMEndAddress)
func (d *Device) SetSRAMAddress(address uint8) {
	d.AddressSRAM = address
	if d.AddressSRAM < SRAMBeginAddres || d.AddressSRAM > SRAMEndAddress {
		d.AddressSRAM = SRAMBeginAddres
	}
}

// SRAMAddress returns current SRAM address
func (d *Device) SRAMAddress() uint8 {
	return d.AddressSRAM
}

// Write writes len(data) bytes to SRAM starting from SRAMAddress
func (d *Device) Write(data []byte) (n int, err error) {
	err = d.bus.WriteRegister(d.Address, d.AddressSRAM, data)
	if err != nil {
		return 0, err
	}
	d.SetSRAMAddress(d.AddressSRAM + uint8(len(data)))
	return len(data), nil
}

// Read reads len(data) from SRAM starting from SRAMAddress
func (d *Device) Read(data []uint8) (n int, err error) {
	err = d.bus.ReadRegister(d.Address, d.AddressSRAM, data)
	if err != nil {
		return 0, err
	}
	d.SetSRAMAddress(d.AddressSRAM + uint8(len(data)))
	return len(data), nil
}

// SetSQW sets square wave output of DS1307
// Available modes: SQW_OFF, SQW_1HZ, SQW_4KHZ, SQW_8KHZ, SQW_32KHZ
func (d *Device) SetSQW(sqw uint8) error {
	err := d.bus.WriteRegister(d.Address, uint8(Control), []byte{sqw})
	return err
}

// IsRunning returns if the oscillator is running
func (d *Device) IsRunning() bool {
	data := []byte{0}
	err := d.bus.ReadRegister(d.Address, uint8(TimeDate), data)
	if err != nil {
		return false
	}
	return (data[0] & (1 << CH)) == 0
}

// SetRunning starts/stops internal oscillator by toggling halt bit
func (d *Device) SetRunning(running bool) error {
	data := []byte{0}
	err := d.bus.ReadRegister(d.Address, uint8(TimeDate), data)
	if err != nil {
		return err
	}
	if running {
		data[0] &^= (1 << CH)
	} else {
		data[0] |= (1 << CH)
	}
	err = d.bus.WriteRegister(d.Address, uint8(TimeDate), data)
	return err
}

// decToBcd converts int to BCD
func decToBcd(dec int) uint8 {
	return uint8(dec + 6*(dec/10))
}

// bcdToDec converts BCD to int
func bcdToDec(bcd uint8) int {
	return int(bcd - 6*(bcd>>4))
}

// hoursBCDToInt converts the BCD hours to int
func hoursBCDToInt(value uint8) (hour int) {
	if value&0x40 != 0x00 {
		hour = bcdToDec(value & 0x1F)
		if (value & 0x20) != 0x00 {
			hour += 12
		}
	} else {
		hour = bcdToDec(value)
	}
	return
}
