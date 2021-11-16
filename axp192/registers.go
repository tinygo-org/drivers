package axp192

// power supply control class
// 0x00 Power supply status register
// 0x01  Power supply mode/charging status register
// 0x04  OTG VBUS status register
// 0x06‐09  Data buffer register
// 0x10  EXTEN & DC‐DC2 switch register
// 0x12  DC‐DC1/3 & LDO2/3switch register
// 0x23  DC‐DC2 voltage set register
// 0x25  DC‐DC2 voltage slope set register
// 0x26  DC‐DC1voltage set register
// 0x27  DC‐DC3 voltage set register
// 0x28  LDO2/3 voltage set register
// 0x30  VBUS‐IPSOUT access set register
// 0x31  VOFF power off voltage set register
// 0x32  Power off、battery detect、CHGLED control register
// 0x33  Charging control register1
// 0x34  Charging control register2
// 0x35  Backup battery charging control register
// 0x36  PEK parameter set register
// 0x37  DCDC switch frequency set register
// 0x38  Battery charging under temperature warning set register
// 0x39  Battery charging over temperature warning set register
// 0x3A  APS under voltage Level1 set register
// 0x3B  APS under voltage Level2 set register
// 0x3C  Battery discharging under temperature warning set register
// 0x3D  Battery discharging over temperature warning set register
// 0x80  DCDC mode set register
// 0x82  ADC enable set register 1
// 0x83  ADC enable set register 2
// 0x84  ADC sample frequency set, TS pin control register
// 0x85  GPIO [3:0] input range set register
// 0x8A  Timer control register
// 0x8B  VBUS monitor set register
// 0x8F  Over temperature power off control register

// GPIO control class
// 0x90  GPIO0 control register
// 0x91  GPIO0 LDO mode output voltage set register
// 0x92  GPIO1 control register
// 0x93  GPIO2 control register
// 0x94  GPIO[2:0] signal status register
// 0x95  GPIO[4:3] function control register
// 0x96  GPIO[4:3] signal status register
// 0x97  GPIO[2:0] pull down control register
// 0x98  PWM1 frequency set register
// 0x99  PWM1 duty ratio set register 1
// 0x9A  PWM1 duty ratio set register 2
// 0x9B  PWM2 frequency set register
// 0x9C  PWM2 duty ratio set register 1
// 0x9D  PWM2 duty ratio set register 2
// 0x9E  GPIO5 control register

// IRQ control class
// 0x40  IRQ enable control register 1
// 0x41  IRQ enable control register 2
// 0x42  IRQ enable control register 3
// 0x43  IRQ enable control register 4
// 0x44  IRQ status register 1
// 0x45  IRQ status register 2
// 0x46  IRQ status register 3
// 0x47  IRQ status register 4

// ADC data class
// 0x56  ACIN voltage ADC data high 8 bit
// 0x57  ACIN voltage ADC data low 4 bit
// 0x58  ACIN current ADC data high 8 bit
// 0x59  ACIN current ADC data low 4 bit
// 0x5A  VBUS voltage ADC data high 8 bit
// 0x5B  VBUS voltage ADC data low 4 bit
// 0x5C  VBUS current ADC data high 8 bit
// 0x5D  VBUS current ADC data low 4 bit
// 0x5E  AXP192 internal temperature monitor ADC data High 8 bit
// 0x5F  AXP192 internal temperature monitor ADC data low 4 bit
// 0x62  TS input ADC data High 8 bit，monitor battery temperature by default
// 0x63  TS input ADC data low 4 bit，monitor battery temperature by default
// 0x64  GPIO0 voltage ADC data high 8 bit
// 0x65  GPIO0 voltage ADC data low 4 bit
// 0x66  GPIO1 voltage ADC data high 8 bit
// 0x67  GPIO1 voltage ADC data low 4 bit
// 0x68  GPIO2 voltage ADC data high 8 bit
// 0x69  GPIO2 voltage ADC data low 4 bit
// 0x6A  GPIO[3] voltage ADC data high 8 bit
// 0x6B  GPIO[3] voltage ADC data low 4 bit
// 0x70  Battery instantaneous power high 8 bit
// 0x71  Battery instantaneous power middle 8 bit
// 0x72  Battery instantaneous power low 8 bit
// 0x78  Battery voltage high 8 bit
// 0x79  Battery voltage low 4 bit
// 0x7A  Battery charging current high 8 bit
// 0x7B  Battery charging current low 5 bit
// 0x7C  Battery discharging current high 8 bit
// 0x7D  Battery discharging current low 5 bit
// 0x7E  APS voltage high 8 bit
// 0x7F  APS voltage low 4 bit
// 0xB0  Battery charging coulomb counter data register 3
// 0xB1  Battery charging coulomb counter data register 2
// 0xB2  Battery charging coulomb counter data register 1
// 0xB3  Battery charging coulomb counter data register 0
// 0xB4  Battery discharging coulomb counter data register 3
// 0xB5  Battery discharging coulomb counter data register 2
// 0xB6  Battery discharging coulomb counter data register 1
// 0xB7  Battery discharging coulomb counter data register 0
// 0xB8  Coulomb counter control register

const (
	// Address is default I2C address.
	Address = 0x34

	RegPowerSupplyStatus            = 0x00
	RegDCDC13LDO23Switch            = 0x12
	RegVbusIPSOutAccessManagement   = 0x30
	RegBackupBatteryChargingControl = 0x35
	RegDCDC2VoltageSet              = 0x25
	RegDCDC1VoltageSet              = 0x26
	RegDCDC3VoltageSet              = 0x27
	RegLDO23VoltageSet              = 0x28
	RegPEKParameterSet              = 0x36
	RegADCEnableSet                 = 0x82

	RegGPIO1Control          = 0x92
	RegGPIO2Control          = 0x93
	RegGPIO20SignalStatus    = 0x94
	RegGPIO43FunctionControl = 0x95
	RegGPIO43SignalStatus    = 0x96
)
