package main

import (
	"errors"
	"strings"
)

var (
	errInvalidCommand = errors.New("Invalid command")
)

func parse(data []byte) error {
	switch {
	case len(data) < 2, string(data[0:2]) != "AT":
		return errInvalidCommand
	case len(data) == 2:
		// just the AT command by itself
		quicktest()
	case len(data) < 6 || data[2] != '+':
		return errInvalidCommand
	default:
		// parse the rest of the command
		cmd, args, _ := strings.Cut(string(data[3:]), "=")
		return parseCommand(cmd, args)
	}

	return nil
}

func parseCommand(cmd, args string) error {
	switch cmd {
	case "VER":
		version()
	case "ID":
		id(args)
	case "RESET":
		reset()
	case "MSG":
		msg(args)
	case "CMSG":
		cmsg(args)
	case "MSGHEX":
		msghex(args)
	case "CMSGHEX":
		cmsghex(args)
	case "PMSG":
		pmsg(args)
	case "PMSGHEX":
		pmsg(args)
	case "PORT":
		port(args)
	case "ADR":
		adr(args)
	case "DR":
		dr(args)
	case "CH":
		ch(args)
	case "POWER":
		power(args)
	case "REPT":
		rept(args)
	case "RETRY":
		retry(args)
	case "RXWIN2":
		rxwin2(args)
	case "RXWIN1":
		rxwin1(args)
	case "KEY":
		key(args)
	case "FDEFAULT":
		fdefault(args)
	case "MODE":
		mode(args)
	case "JOIN":
		join(args)
	case "BEACON":
		join(args)
	case "CLASS":
		class(args)
	case "DELAY":
		delay(args)
	case "LW":
		lw(args)
	case "WDT":
		wdt(args)
	case "LOWPOWER":
		lowpower(args)
	case "VDD":
		vdd(args)
	case "TEMP":
		temp(args)
	case "RTC":
		rtc(args)
	case "EEPROM":
		eeprom(args)
	case "UART":
		uartcmd(args)
	case "TEST":
		test(args)
	case "LOG":
		log(args)
	case "RECV":
		recv(args)
	case "RECVHEX":
		recvhex(args)
	case "SEND":
		send(args)
	case "SENDHEX":
		sendhex(args)
	default:
		return errInvalidCommand
	}

	return nil
}
