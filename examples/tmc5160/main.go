// Connects to SPI1 on a RP2040 (Pico)
package main

import (
	"machine"

	"tinygo.org/x/drivers/tmc5160"
)

func main() {
	// Step 1. Setup your protocol.  SPI setup shown below
	spi := machine.SPI1
	spi.Configure(machine.SPIConfig{
		Frequency: 12000000, // Upto 12 MHZ is pretty stable. Reduce to 5 or 6 Mhz if you are experiencing issues
		Mode:      3,
		LSBFirst:  false,
	})
	// Step 2. Set up all associated Pins
	csPin0 := machine.GPIO13
	csPin0.Configure(machine.PinConfig{Mode: machine.PinOutput})
	enn0 := machine.GPIO18
	enn0.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// csPins is a map of all chip select pins in a multi driver setup.
	//Only one pin csPin0 mapped to "0"is shown in this example, but add more mappings as required
	csPins := map[uint8]machine.Pin{0: csPin0}
	//bind csPin to driverAdddress
	driverAddress := uint8(0) // Let's assume we are working with driver at address 0x01
	// Step 3. Bind the communication interface to the protocol
	comm := tmc5160.NewSPIComm(spi, csPins)
	// Step 4. Define your stepper like this below
	//stepper := tmc5160.NewStepper(angle , gearRatio  vSupply  rCoil , lCoil , iPeak , rSense , mSteps, fclk )
	stepper := tmc5160.NewDefaultStepper() // Default Stepper should be used only for testing.
	// Step 5. Instantiate your driver
	driver := tmc5160.NewDriver(
		comm,
		driverAddress,
		enn0,
		stepper)

	// Setting and getting mode
	rampMode := tmc5160.NewRAMPMODE(comm, driverAddress)
	err := rampMode.SetMode(tmc5160.PositioningMode)
	if err != nil {
		return
	}
	mode, err := rampMode.GetMode()
	if err != nil {
		println("Error getting mode:", err)
	} else {
		println("Current Mode:", mode)
	}

	// Read GCONF register
	GCONF := tmc5160.NewGCONF()
	gconfVal, err := driver.ReadRegister(tmc5160.GCONF)
	// Uppack the register to get all the bits and bytes of the register
	GCONF.Unpack(gconfVal)
	//E.g. MultiStepFlit is retrieved from the GCONF register
	println("GCONF:MultiStepFlit:", GCONF.MultistepFilt)
}
