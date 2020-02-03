package main

import (
	"fmt"
	"io"
	"machine"
	"os"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/drivers/flash"
)

const consoleBufLen = 64
const storageBufLen = 512

var (
	debug = false

	input [consoleBufLen]byte
	store [storageBufLen]byte

	console  = machine.UART0
	readyLED = machine.LED

	dev *flash.Device = flash.NewSPI(
		&machine.SPI1,
		machine.SPI1_MISO_PIN,
		machine.SPI1_MOSI_PIN,
		machine.SPI1_SCK_PIN,
		machine.SPI1_CS_PIN,
	)

	commands map[string]cmdfunc = map[string]cmdfunc{
		"":      cmdfunc(noop),
		"dbg":   cmdfunc(dbg),
		"lsblk": cmdfunc(lsblk),
		"xxd":   cmdfunc(xxd),
	}
)

type cmdfunc func(argv []string)

const (
	StateInput = iota
	StateEscape
	StateEscBrc
	StateCSI
)

func main() {

	time.Sleep(3 * time.Second)

	readyLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	readyLED.High()

	readyLED.Low()
	write("spi Configured. Reading flash info")

	dev.Begin()

	prompt()

	var state = StateInput

	for i := 0; ; {
		if console.Buffered() > 0 {
			data, _ := console.ReadByte()
			if debug {
				fmt.Printf("\rdata: %x\r\n\r", data)
				prompt()
				console.Write(input[:i])
			}
			switch state {
			case StateInput:
				switch data {
				case 0x8:
					fallthrough
				case 0x7f: // this is probably wrong... works on my machine tho :)
					// backspace
					if i > 0 {
						i -= 1
						console.Write([]byte{0x8, 0x20, 0x8})
					}
				case 13:
					// return key
					console.Write([]byte("\r\n"))
					runCommand(string(input[:i]))
					prompt()

					i = 0
					continue
				case 27:
					// escape
					state = StateEscape
				default:
					// anything else, just echo the character if it is printable
					if strconv.IsPrint(rune(data)) {
						if i < (consoleBufLen - 1) {
							console.WriteByte(data)
							input[i] = data
							i++
						}
					}
				}
			case StateEscape:
				switch data {
				case 0x5b:
					state = StateEscBrc
				default:
					state = StateInput
				}
			default:
				// TODO: handle escape sequences
				state = StateInput
			}
			//time.Sleep(10 * time.Millisecond)
		}
	}
}

func runCommand(line string) {
	argv := strings.SplitN(strings.TrimSpace(line), " ", -1)
	cmd := argv[0]
	cmdfn, ok := commands[cmd]
	if !ok {
		println("unknown command: " + line)
		return
	}
	cmdfn(argv)
}

func noop(argv []string) {}

func dbg(argv []string) {
	if debug {
		debug = false
		println("Console debugging off")
	} else {
		debug = true
		println("Console debugging on")
	}
}

func lsblk(argv []string) {
	id, _ := dev.ReadJEDEC()
	status, _ := dev.ReadStatus()
	serialNumber1, _ := dev.ReadSerialNumber()
	fmt.Printf(
		"\n-------------------------------------\r\n"+
			" Device Information:  \r\n"+
			"-------------------------------------\r\n"+
			" JEDEC ID: %v\r\n"+
			"   Serial: %v\r\n"+
			"   Status: %2x\r\n"+
			"-------------------------------------\r\n\r\n",
		id,
		serialNumber1,
		status,
	)
}

func xxd(argv []string) {
	var err error
	var addr uint64 = 0x0
	var size int = 64
	switch len(argv) {
	case 3:
		if size, err = strconv.Atoi(argv[2]); err != nil {
			println("Invalid size argument: " + err.Error() + "\r\n")
			return
		}
		if size > storageBufLen || size < 1 {
			fmt.Printf("Size of hexdump must be greater than 0 and less than %d\r\n", storageBufLen)
			return
		}
		fallthrough
	case 2:
		/*
			if argv[1][:2] != "0x" {
				println("Invalid hex address (should start with 0x)")
				return
			}
		*/
		if addr, err = strconv.ParseUint(argv[1], 16, 32); err != nil {
			println("Invalid address: " + err.Error() + "\r\n")
			return
		}
		fallthrough
	case 1:
		// no args supplied, so nothing to do here, just use the defaults
	default:
		println("usage: xxd <hex address, ex: 0xA0> <size of hexdump in bytes>\r\n")
		return
	}
	buf := store[0:size]
	//fatdisk.ReadAt(buf, int64(addr))
	dev.ReadBuffer(uint32(addr), buf)
	xxdfprint(os.Stdout, uint32(addr), buf)
}

func xxdfprint(w io.Writer, offset uint32, b []byte) {
	var l int
	var buf16 = make([]byte, 16)
	for i, c := 0, len(b); i < c; i += 16 {
		l = i + 16
		if l >= c {
			l = c
		}
		fmt.Fprintf(w, "%08x: % x    ", offset+uint32(i), b[i:l])
		for j, n := 0, l-i; j < 16; j++ {
			if j >= n || !strconv.IsPrint(rune(b[i+j])) {
				buf16[j] = '.'
			} else {
				buf16[j] = b[i+j]
			}
		}
		console.Write(buf16)
		println()
		//	"%s\r\n", b[i:l], "")
	}
}

func write(s string) {
	println(s)
}

func prompt() {
	print("==> ")
}
