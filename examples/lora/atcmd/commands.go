package main

import (
	"strings"
)

func crlf()  {
	uart.Write([]byte("\r\n"))
}

// Use to test if connection of module is OK.
func quicktest() {
	uart.Write([]byte("+AT: OK"))
	crlf()
}

// Check firmware version.
func version() {
	uart.Write([]byte("+VER: "))
	uart.Write([]byte(currentVersion()))
	crlf()
}

// Use to check the ID of the LoRaWAN module, or change the ID.
func id(args string) error {
	// look for comma in args
	param, val, hasComma := strings.Cut(args, ",")
	if hasComma {
		// set
		switch param {
		case "DevAddr":
			uart.Write([]byte("+ID: DevAddr, "))
			uart.Write([]byte(val))
			crlf()				
		case "DevEui":
			uart.Write([]byte("+ID: DevEui, "))
			uart.Write([]byte(val))
			crlf()					
		case "AppEui":
			uart.Write([]byte("+ID: AppEui, "))
			uart.Write([]byte(val))
			crlf()			
		default:
			return errInvalidCommand
		}

		return nil
	}

	// get
	switch param {
	case "DevAddr":
		uart.Write([]byte("+ID: DevAddr, xx:xx:xx:xx"))
		crlf()
	case "DevEui":
		uart.Write([]byte("+ID: DevEui, xx:xx:xx:xx:xx:xx:xx:xx"))
		crlf()	
	case "AppEui":
		uart.Write([]byte("+ID: AppEui, xx:xx:xx:xx:xx:xx:xx:xx"))
		crlf()
	default:
		uart.Write([]byte("+ID: DevAddr, xx:xx:xx:xx"))
		crlf()
		uart.Write([]byte("+ID: DevEui, xx:xx:xx:xx:xx:xx:xx:xx"))
		crlf()	
		uart.Write([]byte("+ID: AppEui, xx:xx:xx:xx:xx:xx:xx:xx"))
		crlf()
	}

	return nil
}

// Use to reset the module. If module returns error, then reset function is invalid.
func reset() error { 
	uart.Write([]byte("+RESET: OK"))
	crlf()
	return nil
}

// Use to send string format frame which is no need to be confirmed by the server.
func msg(data string) error {
	uart.Write([]byte("+MSG: Start"))
	crlf()
	uart.Write([]byte("+MSG: Done"))
	crlf()
	return nil
}

// Use to send string format frame which must be confirmed by the server
func cmsg(data string) error {
	uart.Write([]byte("+CMSG: Start"))
	crlf()
	uart.Write([]byte("+CMSG: Done"))
	crlf()
	return nil
}

// Use to send hex format frame which is no need to be confirmed by the server
func msghex(data string) error {
	uart.Write([]byte("+MSGHEX: Start"))
	crlf()
	uart.Write([]byte("+MSGHEX: Done"))
	crlf()
	return nil
}

// Use to send hex format frame which must be confirmed by the server.
func cmsghex(data string) error {
	uart.Write([]byte("+CMSGHEX: Start"))
	crlf()
	uart.Write([]byte("+CMSGHEX: Done"))
	crlf()
	return nil
}

// Use to send string format LoRaWAN proprietary frames
func pmsg(data string) error {
	uart.Write([]byte("+PMSG: Start"))
	crlf()
	uart.Write([]byte("+PMSG: Done"))
	crlf()
	return nil
}

// Use to send hex format LoRaWAN proprietary frames.
func pmsghex(data string) error {
	uart.Write([]byte("+PMSGHEX: Start"))
	crlf()
	uart.Write([]byte("+PMSGHEX: Done"))
	crlf()
	return nil
}

// Set PORT number which will be used by MSG/CMSG/MSGHEX/CMSGHEX command to send
// message, port number should range from 1 to 255. User should refer to LoRaWAN
// specification to choose port.
func port(p string) error {
	uart.Write([]byte("+PORT: "))
	uart.Write([]byte(p))
	crlf()
	return nil
}

// Set ADR function of LoRaWAN module
func adr(state string) error {
	uart.Write([]byte("+ADR: "))
	uart.Write([]byte(state))
	crlf()
	return nil
}

// Use LoRaWAN defined DRx to set datarate of LoRaWAN AT modem.
func dr(rate string) error {
	uart.Write([]byte("+ADR: "))
	uart.Write([]byte(rate))
	crlf()
	return nil
}

// Channel Configuration
func ch(channel string) error {
	uart.Write([]byte("+CH: "))
	uart.Write([]byte(channel))
	crlf()
	return nil
}

// Set and Check Power
func power(setting string) error {
	uart.Write([]byte("+POWER: "))
	uart.Write([]byte(setting))
	crlf()
	return nil
}

// Unconfirmed message repeats times. 
func rept(setting string) error {
	uart.Write([]byte("+REPT: "))
	uart.Write([]byte(setting))
	crlf()
	return nil
}

// Confirmed message retry times. Valid range 0~254, 
// if retry times is less than 2, only one message will
// be sent. Random delay 3 - 10s between each retry 
// (band duty cycle limitation has the priority)
func retry(setting string) error {
	uart.Write([]byte("+RETRY: "))
	uart.Write([]byte(setting))
	crlf()
	return nil
}

func rxwin2(setting string) error {
	uart.Write([]byte("+RXWIN2: "))
	uart.Write([]byte(setting))
	crlf()
	return nil
}

func rxwin1(setting string) error {
	uart.Write([]byte("+RXWIN1: "))
	uart.Write([]byte(setting))
	crlf()
	return nil
}

func key(setting string) error {
	uart.Write([]byte("+RETRY: "))
	uart.Write([]byte(setting))
	crlf()
	return nil
}

func fdefault(setting string) error {
	uart.Write([]byte("+FDEFAULT: OK"))
	crlf()
	return nil
}

func mode(setting string) error {
	uart.Write([]byte("+MODE: "))
	uart.Write([]byte(setting))
	crlf()
	return nil
}

func join(setting string) error {
	uart.Write([]byte("+JOIN: Starting"))
	crlf()
	uart.Write([]byte("+JOIN: Done"))
	crlf()
	return nil
}

func beacon(setting string) error {
	uart.Write([]byte("+BEACON: Starting"))
	crlf()
	uart.Write([]byte("+BEACON: Done"))
	crlf()
	return nil
}

func class(setting string) error {
	uart.Write([]byte("+CLASS: Starting"))
	crlf()
	uart.Write([]byte("+CLASS: Done"))
	crlf()
	return nil
}

func delay(setting string) error {
	uart.Write([]byte("+DELAY: Starting"))
	crlf()
	return nil
}

func lw(setting string) error {
	uart.Write([]byte("+LW: Starting"))
	crlf()
	return nil
}

func wdt(setting string) error {
	uart.Write([]byte("+WDT: Not implemented"))
	crlf()
	return nil
}

func lowpower(setting string) error {
	uart.Write([]byte("+LOWPOWER: Not implemented"))
	crlf()
	return nil
}

func vdd(setting string) error {
	uart.Write([]byte("+VDD: Not implemented"))
	crlf()
	return nil
}

func temp(setting string) error {
	uart.Write([]byte("+TEMP: Not implemented"))
	crlf()
	return nil
}

func rtc(setting string) error {
	uart.Write([]byte("+RTC: Not implemented"))
	crlf()
	return nil
}

func eeprom(setting string) error {
	uart.Write([]byte("+EEPROM: Not implemented"))
	crlf()
	return nil
}

func uartcmd(setting string) error {
	uart.Write([]byte("+UART: Not implemented"))
	crlf()
	return nil
}

func test(setting string) error {
	uart.Write([]byte("+TEST: Not implemented"))
	crlf()
	return nil
}

func log(setting string) error {
	uart.Write([]byte("+LOG: Not implemented"))
	crlf()
	return nil
}
