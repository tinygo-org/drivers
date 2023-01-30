package main

import (
	"encoding/hex"
	"strings"

	"tinygo.org/x/drivers/examples/lora/lorawan/common"
	"tinygo.org/x/drivers/lora/lorawan"
)

// Use to test if connection to module is OK.
func quicktest() {
	writeCommandOutput("AT", "OK")
}

// Check firmware version.
func version() {
	writeCommandOutput("VER", common.CurrentVersion()+" ("+common.FirmwareVersion()+")")
}

// Use to check the ID of the LoRaWAN module, or change the ID.
func id(args string) error {
	cmd := "ID"

	// look for comma in args
	param, val, hasComma := strings.Cut(args, ",")
	if hasComma {
		// set
		data := strings.Trim(val, "\"'")

		// convert data from hex formatted string
		data = strings.ReplaceAll(data, " ", "")
		hexdata, err := hex.DecodeString(data)
		if err != nil {
			writeCommandOutput(cmd, err.Error())

			return err
		}

		switch param {
		case "DevAddr":
			if err := session.SetDevAddr(hexdata); err != nil {
				writeCommandOutput(cmd, err.Error())

				return err
			}

			writeCommandOutput(cmd, "DevAddr, "+session.GetDevAddr())
		case "DevEui":
			if err := otaa.SetDevEUI(hexdata); err != nil {
				writeCommandOutput(cmd, err.Error())

				return err
			}

			writeCommandOutput(cmd, "DevEui, "+otaa.GetDevEUI())
		case "AppEui":
			if err := otaa.SetAppEUI(hexdata); err != nil {
				writeCommandOutput(cmd, err.Error())

				return err
			}

			writeCommandOutput(cmd, "AppEui, "+otaa.GetAppEUI())
		default:
			return errInvalidCommand
		}

		return nil
	}

	// get
	switch param {
	case "DevAddr":
		writeCommandOutput(cmd, "DevAddr, "+session.GetDevAddr())
	case "DevEui":
		writeCommandOutput(cmd, "DevEui, "+otaa.GetDevEUI())
	case "AppEui":
		writeCommandOutput(cmd, "AppEui, "+otaa.GetAppEUI())
	default:
		writeCommandOutput(cmd, "DevAddr, "+session.GetDevAddr())
		writeCommandOutput(cmd, "DevEui, "+otaa.GetDevEUI())
		writeCommandOutput(cmd, "AppEui, "+otaa.GetAppEUI())
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

	if err := radio.Tx([]byte(data), defaultTimeout); err != nil {
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

	if err := radio.Tx([]byte(data), defaultTimeout); err != nil {
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

	// look for comma in args
	param, val, hasComma := strings.Cut(setting, ",")
	if hasComma {
		// set
		data := strings.Trim(val, "\"'")

		// convert data from hex formatted string
		data = strings.ReplaceAll(data, " ", "")
		hexdata, err := hex.DecodeString(data)
		if err != nil {
			writeCommandOutput(cmd, err.Error())

			return err
		}

		switch param {
		case "APPKEY":
			if err := otaa.SetAppKey(hexdata); err != nil {
				writeCommandOutput(cmd, err.Error())

				return err
			}

			writeCommandOutput(cmd, "APPKEY, "+otaa.GetAppKey())
		default:
			return errInvalidCommand
		}
	}

	// cannot get keys
	return errInvalidCommand
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
	// TODO: check that DevEUI, AppEUI, and AppKey have values

	cmd := "JOIN"
	writeCommandOutput(cmd, "Starting")
	if err := lorawan.Join(otaa, session); err != nil {
		writeCommandOutput(cmd, err.Error())

		return err
	}

	writeCommandOutput(cmd, "Network joined")
	writeCommandOutput(cmd, "DevEui, "+otaa.GetDevEUI())
	writeCommandOutput(cmd, "AppEui, "+otaa.GetAppEUI())
	writeCommandOutput(cmd, "DevAddr, "+session.GetDevAddr())
	writeCommandOutput(cmd, "NetID, "+otaa.GetNetID())
	writeCommandOutput(cmd, "NwkSKey, "+session.GetNwkSKey())
	writeCommandOutput(cmd, "AppSKey, "+session.GetAppSKey())
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

	param, val, hasComma := strings.Cut(setting, ",")
	if hasComma {
		if param == "NET" {
			if val == "ON" {
				lorawan.SetPublicNetwork(true)
			} else {
				lorawan.SetPublicNetwork(false)
			}
		}
	}
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

func send(data string) error {
	cmd := "SEND"
	writeCommandOutput(cmd, "Start")

	// remove leading/trailing quotes
	data = strings.Trim(data, "\"'")

	if err := radio.Tx([]byte(data), defaultTimeout); err != nil {
		writeCommandOutput(cmd, err.Error())

		return err
	}

	writeCommandOutput(cmd, "Done")
	return nil
}

func sendhex(data string) error {
	cmd := "SENDHEX"
	writeCommandOutput(cmd, "Start")

	// remove leading/trailing quotes
	data = strings.Trim(data, "\"'")

	// convert data from hex formatted string
	data = strings.ReplaceAll(data, " ", "")
	payload, err := hex.DecodeString(data)
	if err != nil {
		writeCommandOutput(cmd, err.Error())

		return err
	}

	if err := radio.Tx(payload, defaultTimeout); err != nil {
		writeCommandOutput(cmd, err.Error())

		return err
	}

	writeCommandOutput(cmd, "Done")
	return nil
}

func recv(setting string) error {
	cmd := "RECV"

	data, err := common.Lorarx()
	if err != nil {
		writeCommandOutput(cmd, "ERROR "+err.Error())
		return err
	}
	writeCommandOutput(cmd, string(data))

	return nil
}

func recvhex(setting string) error {
	cmd := "RECVHEX"

	data, err := common.Lorarx()
	if err != nil {
		writeCommandOutput(cmd, "ERROR "+err.Error())
		return err
	}
	writeCommandOutput(cmd, string(data))

	return nil
}

func crlf() {
	uart.Write([]byte("\r\n"))
}

func writeCommandOutput(cmd, data string) {
	uart.Write([]byte("+" + cmd + ": "))
	uart.Write([]byte(data))
	crlf()
}
