# BMI270

This package reads the accelerometer, gyroscope, and temperature sensor of the
Bosch BMI270 IMU over I2C or SPI.

Use `NewI2C(bus)` or `NewSPI(bus, cs)`, then call `Configure(Config{})`. The
zero `Config` uses ±8 g at 100 Hz and ±2000 °/s at 200 Hz. `Address` is `0x68`
with SDO connected to GND. Set `dev.Address` to `AddressAlt` (`0x69`) if SDO is
connected to VDDIO. For SPI, configure the bus in mode 0 or 3 at up to 10 MHz,
and configure `cs` as an output set high.

`Configure` performs a soft reset and uploads the 8 kB config file required by
the device. The file is from the
[Bosch BMI270 Sensor API](https://github.com/boschsensortec/BMI270_SensorAPI)
and is covered by the license in `LICENSE-config`.

`ReadAcceleration` returns µg, `ReadRotation` returns µ°/s, and
`ReadTemperature` returns milli °C. The device also implements
`drivers.Sensor` with `Update`, `Acceleration`, `AngularVelocity`, and
`Temperature`.

See the [BMI270 datasheet](https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmi270-ds000.pdf)
for the register map.
