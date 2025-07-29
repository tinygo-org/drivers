package tmc2209

func Constrain(value, low, high uint32) uint32 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func SetRunCurrent(percent uint8) {
	_ = PercentToCurrentSetting(percent)

	// Set the run current register to runCurrent value
}

func SetHoldCurrent(percent uint8) {
	_ = PercentToCurrentSetting(percent)
	// Set the hold current register to holdCurrent value
}
func PercentToCurrentSetting(percent uint8) uint8 {
	constrainedPercent := Constrain(uint32(percent), 0, 100)
	return uint8(Map(constrainedPercent, 0, 100, 0, 255))
}

func CurrentSettingToPercent(currentSetting uint8) uint8 {
	return uint8(Map(uint32(currentSetting), 0, 255, 0, 100))
}

func PercentToHoldDelaySetting(percent uint8) uint8 {
	constrainedPercent := Constrain(uint32(percent), 0, 100)
	return uint8(Map(constrainedPercent, 0, 100, 0, 255))
}

func HoldDelaySettingToPercent(holdDelaySetting uint8) uint8 {
	return uint8(Map(uint32(holdDelaySetting), 0, 255, 0, 100))
}
func Map(value, fromLow, fromHigh, toLow, toHigh uint32) uint32 {
	return (value-fromLow)*(toHigh-toLow)/(fromHigh-fromLow) + toLow
}
