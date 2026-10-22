# INA3221

This package reads the three channels of the TI INA3221 shunt and bus voltage
monitor.

Use `New(bus)` with a `drivers.I2C` bus, then call `Configure(DefaultConfig)`.
`Address` is `0x40` with A0 connected to GND. Set `dev.Address` to `0x41`,
`0x42`, or `0x43` if A0 is connected to VS, SDA, or SCL.

Channels are numbered 1 to 3, as in the datasheet. `BusVoltage` and
`ShuntVoltage` return µV. `Current` returns µA and `Power` returns µW. Both use
the shunt resistance of the channel, which is 100 milliohms by default. Use
`SetShuntResistance(channel, milliohms)` if your board uses a different value.

In a single-shot mode, call `Trigger()` to start a conversion and poll
`ConversionReady()` before reading.

See the [INA3221 datasheet](https://www.ti.com/lit/ds/symlink/ina3221.pdf)
for the register map.
