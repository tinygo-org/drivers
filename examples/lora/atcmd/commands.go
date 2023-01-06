package main

import (
	"strings"
)

// Use to test if connection to module is OK.
func quicktest() {
	writeCommandOutput("AT", "OK")
}

// Check firmware version.
func version() {
	writeCommandOutput("VER", currentVersion())
}

// Use to check the ID of the LoRaWAN module, or change the ID.
func id(args string) error {
	cmd := "ID"

	// look for comma in args
	param, val, hasComma := strings.Cut(args, ",")
	if hasComma {
		// set
		switch param {
		case "DevAddr":
			writeCommandOutput(cmd, "DevAddr, "+val)
		case "DevEui":
			writeCommandOutput(cmd, "DevEui, "+val)
		case "AppEui":
			writeCommandOutput(cmd, "AppEui, "+val)
		default:
			return errInvalidCommand
		}

		return nil
	}

	// get
	switch param {
	case "DevAddr":
		writeCommandOutput(cmd, "DevAddr, xx:xx:xx:xx")
	case "DevEui":
		writeCommandOutput(cmd, "DevEui, xx:xx:xx:xx:xx:xx:xx:xx")
	case "AppEui":
		writeCommandOutput(cmd, "AppEui, xx:xx:xx:xx:xx:xx:xx:xx")
	default:
		writeCommandOutput(cmd, "DevAddr, xx:xx:xx:xx")
		writeCommandOutput(cmd, "DevEui, xx:xx:xx:xx:xx:xx:xx:xx")
		writeCommandOutput(cmd, "AppEui, xx:xx:xx:xx:xx:xx:xx:xx")
	}

	return nil
}

// Use to reset the module. If module returns error, then reset function is invalid.
func reset() error {
	radio.Reset()

	writeCommandOutput("RESET", "OK")
	return nil
}

// Use to send string format frame which is no need to be confirmed by the server.
func msg(data string) error {
	cmd := "MSG"
	writeCommandOutput(cmd, "Start")

	if err := radio.LoraTx([]byte(data), defaultTimeout); err != nil {
		writeCommandOutput(cmd, err.Error())

		return err
	}

	writeCommandOutput(cmd, "Done")
	return nil
}

// Use to send string format frame which must be confirmed by the server
func cmsg(data string) error {
	cmd := "CMSG"
	writeCommandOutput(cmd, "Start")

	if err := radio.LoraTx([]byte(data), defaultTimeout); err != nil {
		writeCommandOutput(cmd, err.Error())

		return err
	}

	// TODO: confirmation

	writeCommandOutput(cmd, "Done")
	return nil
}

// Use to send hex format frame which is no need to be confirmed by the server
func msghex(data string) error {
	cmd := "MSGHEX"
	writeCommandOutput(cmd, "Start")

	writeCommandOutput(cmd, "Done")
	return nil
}

// Use to send hex format frame which must be confirmed by the server.
func cmsghex(data string) error {
	cmd := "CMSGHEX"
	writeCommandOutput(cmd, "Start")

	writeCommandOutput(cmd, "Done")
	return nil
}

// Use to send string format LoRaWAN proprietary frames
func pmsg(data string) error {
	cmd := "PMSG"
	writeCommandOutput(cmd, "Start")

	writeCommandOutput(cmd, "Done")
	return nil
}

// Use to send hex format LoRaWAN proprietary frames.
func pmsghex(data string) error {
	cmd := "PMSGHEX"
	writeCommandOutput(cmd, "Start")

	writeCommandOutput(cmd, "Done")

	return nil
}

// Set PORT number which will be used by MSG/CMSG/MSGHEX/CMSGHEX command to send
// message, port number should range from 1 to 255. User should refer to LoRaWAN
// specification to choose port.
func port(p string) error {
	cmd := "PMSG"
	writeCommandOutput(cmd, p)

	return nil
}

// Set ADR function of LoRaWAN module
func adr(state string) error {
	cmd := "ADR"
	writeCommandOutput(cmd, state)

	return nil
}

// Use LoRaWAN defined DRx to set datarate of LoRaWAN AT modem.
func dr(rate string) error {
	cmd := "DR"
	writeCommandOutput(cmd, rate)

	return nil
}

// Channel Configuration
func ch(channel string) error {
	cmd := "CH"
	writeCommandOutput(cmd, channel)

	return nil
}

// Set and Check Power
func power(setting string) error {
	cmd := "POWER"
	writeCommandOutput(cmd, setting)

	return nil
}

// Unconfirmed message repeats times. 
func rept(setting string) error {
	cmd := "REPT"
	writeCommandOutput(cmd, setting)

	return nil
}

// Confirmed message retry times. Valid range 0~254, 
// if retry times is less than 2, only one message will
// be sent. Random delay 3 - 10s between each retry 
// (band duty cycle limitation has the priority)
func retry(setting string) error {
	cmd := "RETRY"
	writeCommandOutput(cmd, setting)

	return nil
}

func rxwin2(setting string) error {
	cmd := "RXWIN2"
	writeCommandOutput(cmd, setting)

	return nil
}

func rxwin1(setting string) error {
	cmd := "RXWIN1"
	writeCommandOutput(cmd, setting)

	return nil
}

func key(setting string) error {
	cmd := "KEY"
	writeCommandOutput(cmd, setting)

	return nil
}

func fdefault(setting string) error {
	cmd := "FDEFAULT"
	writeCommandOutput(cmd, "OK")

	return nil
}

func mode(setting string) error {
	cmd := "MODE"
	writeCommandOutput(cmd, setting)

	return nil
}

func join(setting string) error {
	cmd := "JOIN"
	writeCommandOutput(cmd, "Starting")

	writeCommandOutput(cmd, "Done")
	return nil
}

func beacon(setting string) error {
	cmd := "BEACON"
	writeCommandOutput(cmd, "Starting")

	writeCommandOutput(cmd, "Done")

	return nil
}

func class(setting string) error {
	cmd := "CLASS"
	writeCommandOutput(cmd, "Starting")

	writeCommandOutput(cmd, "Done")

	return nil
}

func delay(setting string) error {
	cmd := "DELAY"
	writeCommandOutput(cmd, setting)

	return nil
}

func lw(setting string) error {
	cmd := "LW"
	writeCommandOutput(cmd, setting)

	return nil
}

func wdt(setting string) error {
	cmd := "WDT"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func lowpower(setting string) error {
	cmd := "LOWPOWER"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func vdd(setting string) error {
	cmd := "VDD"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func temp(setting string) error {
	cmd := "TEMP"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func rtc(setting string) error {
	cmd := "RTC"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func eeprom(setting string) error {
	cmd := "EEPROM"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func uartcmd(setting string) error {
	cmd := "UART"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func test(setting string) error {
	cmd := "TEST"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func log(setting string) error {
	cmd := "LOG"
	writeCommandOutput(cmd, "Not implemented")

	return nil
}

func crlf()  {
	uart.Write([]byte("\r\n"))
}

func writeCommandOutput(cmd, data string) {
	uart.Write([]byte("+"+cmd+": "))
	uart.Write([]byte(data))
	crlf()
}
