package console_example

import (
	"fmt"
	"io"
	"machine"
	"os"
	"strconv"
	"strings"

	"tinygo.org/x/drivers/flash"
)

const consoleBufLen = 64
const storageBufLen = 512

var (
	debug = false

	input [consoleBufLen]byte
	store [storageBufLen]byte

	console = machine.Serial

	dev *flash.Device

	commands map[string]cmdfunc = map[string]cmdfunc{
		"":      cmdfunc(noop),
		"erase": cmdfunc(erase),
		"lsblk": cmdfunc(lsblk),
		"write": cmdfunc(write),
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

func RunFor(device *flash.Device) {

	dev = device
	dev.Configure(&flash.DeviceConfig{
		Identifier: flash.DefaultDeviceIdentifier,
	})

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

func lsblk(argv []string) {
	attrs := dev.Attrs()
	status1, _ := dev.ReadStatus()
	status2, _ := dev.ReadStatus2()
	serialNumber1, _ := dev.ReadSerialNumber()
	fmt.Printf(
		"\n-------------------------------------\r\n"+
			" Device Information:  \r\n"+
			"-------------------------------------\r\n"+
			" JEDEC ID: %v\r\n"+
			" Serial:   %v\r\n"+
			" Status 1: %02x\r\n"+
			" Status 2: %02x\r\n"+
			" \r\n"+
			" Max clock speed (MHz): %d\r\n"+
			" Has Sector Protection: %t\r\n"+
			" Supports Fast Reads:   %t\r\n"+
			" Supports QSPI Reads:   %t\r\n"+
			" Supports QSPI Write:   %t\r\n"+
			" Write Status Split:    %t\r\n"+
			" Single Status Byte:    %t\r\n"+
			"-------------------------------------\r\n\r\n",
		attrs.JedecID,
		serialNumber1,
		status1,
		status2,
		attrs.MaxClockSpeedMHz,
		attrs.HasSectorProtection,
		attrs.SupportsFastRead,
		attrs.SupportsQSPI,
		attrs.SupportsQSPIWrites,
		attrs.WriteStatusSplit,
		attrs.SingleStatusByte,
	)
}

func erase(argv []string) {
	if len(argv) < 3 {
		println("usage: erase <chip|block|sector> <bytes>")
		return
	}
	var err error
	var addr uint64 = 0x0
	if addr, err = strconv.ParseUint(argv[2], 16, 32); err != nil {
		println("Invalid address: " + err.Error() + "\r\n")
		return
	}
	if argv[1] == "block" {
		if err = dev.EraseBlock(uint32(addr)); err != nil {
			println("Block erase error: " + err.Error() + "\r\n")
		}
	} else if argv[1] == "sector" {
		if err = dev.EraseSector(uint32(addr)); err != nil {
			println("Sector erase error: " + err.Error() + "\r\n")
		}
	} else if argv[1] == "chip" {
		if err = dev.EraseAll(); err != nil {
			println("Chip erase error: " + err.Error() + "\r\n")
		}
	} else {
		println("usage: erase <chip|block|sector> <bytes>")
	}
}

func write(argv []string) {
	if len(argv) < 3 {
		println("usage: write <hex offset> <bytes>")
	}
	var err error
	var addr uint64 = 0x0
	if addr, err = strconv.ParseUint(argv[1], 16, 32); err != nil {
		println("Invalid address: " + err.Error() + "\r\n")
		return
	}
	buf := []byte(argv[2])
	if _, err = dev.WriteAt(buf, int64(addr)); err != nil {
		println("Write error: " + err.Error() + "\r\n")
	}
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
	dev.ReadAt(buf, int64(addr))
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
	}
}

func prompt() {
	print("==> ")
}
