// This example demonstrates how to use the STUSB4500 driver to monitor cable
// attach/detach, read the USB PD source capabilities, and negotiate different
// power profiles.
//
// An external button is used to cycle through each of the power profiles
// discovered.
//
// An LED is used to indicate USB PD negotiation activity.
//
// Note that your STUSB4500 should be powered via VBUS, which is connected to
// the USB Type-C cable itself, and NOT via VSYS/VCC (which should be connected
// to ground).
package main

import (
	"machine"
	"runtime"
	"time"

	"tinygo.org/x/drivers/stusb4500"
	"tinygo.org/x/drivers/stusb4500/conf"
)

// STUSB4500 auxiliary pins
const (
	// not connecting these pins is possible (with reduced functionality) if the
	// connection is managed manually, but the Monitor function will be disabled.
	resetPin  = machine.D6 // GPIO output
	alertPin  = machine.D5 // GPIO input pullup (with external interrupt)
	attachPin = machine.D4 // GPIO input pullup (with external interrupt)
)

// Other pins for demo program functionality
const (
	// the LED will turn on the instant a new power profile is requested and will
	// turn off once the new power contract is negotiated (successfully or not).
	//
	// there is no way to verify the PD source is actually providing the requested
	// power; we can only verify it was accepted. you must use a voltage meter of
	// some sort (e.g. handheld digital multimeter, voltage/current sensor IC, or
	// something else...? just don't try using an ADC on your MCU for this, unless
	// you're absolutely 100% positive it is +20V tolerant (it isn't)).
	ledPin    = machine.LED // GPIO output
	buttonPin = machine.D10 // GPIO input pullup (with external interrupt)
)

// declare our device globally for easy reference from the callback routines
var usbpd *stusb4500.Device

func main() {

	// register at most 1 button press per second
	const buttonBounceDuration = time.Second

	// state variables associated with button debounce logic
	var (
		isButtonPressed  bool
		wasButtonPressed bool
		whenButtonReady  time.Time
	)

	// configure the GPIO pins connected to the LED and external button
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	buttonPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	buttonPin.SetInterrupt(machine.PinFalling,
		func(machine.Pin) { isButtonPressed = true })

	println("-- initializing STUSB4500 on I2C0")

	// configure the I2C interface
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
		SDA:       machine.SDA_PIN,
		SCL:       machine.SCL_PIN,
	})

	// configure the auxiliary STUSB4500 pins
	resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	alertPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	attachPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// create and initialize the STUSB4500 device connection
	usbpd = stusb4500.New(machine.I2C0).Configure(conf.Configuration{
		// register the GPIO pins for auxiliary connections
		ResetPin:  resetPin,
		AlertPin:  alertPin,
		AttachPin: attachPin,
		// install callback functions
		OnInitFail:     deviceError,
		OnResetFail:    deviceError,
		OnConnect:      deviceConnected,
		OnConnectFail:  deviceError,
		OnError:        deviceError,
		OnCableAttach:  cableAttached,
		OnCableDetach:  cableDetached,
		OnCapabilities: capabilitiesReceived,
	})

	// watch indefinitely for button presses
	go func(dev *stusb4500.Device) {
		// index of the currently selected power profile
		pdoIndex := 0
		for {
			if !wasButtonPressed && isButtonPressed {
				wasButtonPressed = true
				whenButtonReady = time.Now().Add(buttonBounceDuration)

				// turn on the LED to indicate a new power profile is being requested
				ledPin.High()

				// select the next power profile
				if pdoCount := len(dev.SrcPDO); pdoCount > 0 {
					pdoIndex = (pdoIndex + 1) % pdoCount
					println("-- requesting new power profile")
					println("      " + dev.SrcPDO[pdoIndex].String())
					dev.SetPower(dev.SrcPDO[pdoIndex].Voltage, dev.SrcPDO[pdoIndex].Current)
				} else {
					println("!! no USB PD source capabilities received")
				}

			} else if wasButtonPressed && time.Now().After(whenButtonReady) {
				wasButtonPressed = false
				isButtonPressed = false
			}
			// yield to the monitor goroutine
			runtime.Gosched()
		}
	}(usbpd)

	println("-- starting USB PD monitor")

	// `Monitor` does not return until `StopMonitor` is called from another
	// goroutine. See its godoc comments for usage details.
	usbpd.Monitor()
}

func deviceError(err error) {
	if nil != err {
		println("!! " + err.Error())
	}
}

func deviceConnected() { println("-- connected!") }
func cableAttached()   { println("-- cable attached") }
func cableDetached()   { println("-- cable detached") }

func capabilitiesReceived() {
	println("-- source capabilities received")

	// turn off the LED once we've received source capabilities again, which will
	// occur after we have set a new power profile.
	ledPin.Low()

	println("   source PDOs:")
	for _, pdo := range usbpd.SrcPDO {
		println("      " + pdo.String())
	}

	println("   sink PDOs:")
	for _, pdo := range usbpd.SnkPDO {
		println("      " + pdo.String())
	}
}
