package main

import (
	"fmt"
	"io"
	"machine"
	"os"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/drivers/sdcard"
)

const consoleBufLen = 64
const storageBufLen = 1024

var (
	debug = false

	input [consoleBufLen]byte
	store [storageBufLen]byte

	console = machine.Serial

	dev *sdcard.Device

	commands map[string]cmdfunc = map[string]cmdfunc{
		"":      cmdfunc(noop),
		"help":  cmdfunc(help),
		"dbg":   cmdfunc(dbg),
		"erase": cmdfunc(erase),
		"lsblk": cmdfunc(lsblk),
		"write": cmdfunc(write),
		"xxd":   cmdfunc(xxd),
	}

	his history
)

type history struct {
	buf [32]string
	wp  int
	idx int
}

func (h *history) Add(cmd string) {
	if len(cmd) == 0 {
		h.idx = h.wp
		return
	}

	if h.wp == len(h.buf)-1 {
		for i := 1; i < len(h.buf); i++ {
			h.buf[i-1] = h.buf[i]
		}
		h.wp--
	}

	h.buf[h.wp] = cmd
	h.wp++
	h.idx = h.wp
}

func (h *history) PeekPrev() string {
	if h.idx > 0 {
		h.idx--
	}
	return h.buf[h.idx]
}

func (h *history) PeekNext() string {
	if h.idx < h.wp {
		h.idx++
	}
	return h.buf[h.idx]
}

type cmdfunc func(argv []string)

const (
	StateInput = iota
	StateEscape
	StateEscBrc
	StateCSI
)

func RunFor(device *sdcard.Device) {

	dev = device

	prompt()

	var state = StateInput

	for i := 0; ; {
		if console.Buffered() > 0 {
			data, _ := console.ReadByte()
			if debug {
				fmt.Printf("\rdata: %x, his.idx: %d\r\n\r", data, his.idx)
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
					his.Add(string(input[:i]))
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
			case StateEscBrc:
				switch data {
				case 0x41:
					// up
					println()
					prompt()
					cmd := his.PeekPrev()
					i = len(cmd)
					copy(input[:i], []byte(cmd))
					console.Write(input[:i])
					state = StateInput
				case 0x42:
					//down
					println()
					prompt()
					cmd := his.PeekNext()
					i = len(cmd)
					copy(input[:i], []byte(cmd))
					console.Write(input[:i])
					state = StateInput
				default:
					// TODO: handle escape sequences
					state = StateInput
				}
			default:
				// TODO: handle escape sequences
				state = StateInput
			}
		} else {
			time.Sleep(1 * time.Millisecond)
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

func help(argv []string) {
	fmt.Printf("help\r\n")
	fmt.Printf("dbg\r\n")
	fmt.Printf("erase\r\n")
	fmt.Printf("lsblk\r\n")
	fmt.Printf("write <hex offset> <bytes>\r\n")
	fmt.Printf("xxd <start address> <length>\r\n")
}

func dbg(argv []string) {
	if debug {
		debug = false
		println("Console debbuging off")
	} else {
		debug = true
		println("Console debbuging on")
	}
}

func lsblk(argv []string) {
	csd := dev.CSD
	sectors, err := csd.Sectors()
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		return
	}
	cid := dev.CID

	fmt.Printf(
		"\r\n-------------------------------------\r\n"+
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
			"-Sectors:               %d\r\n"+
			"-Bytes (Sectors * 512)  %d\r\n"+
			"-ManufacturerID         %02X\r\n"+
			"-OEMApplicationID       %04X\r\n"+
			"-ProductName            %s\r\n"+
			"-ProductVersion         %s\r\n"+
			"-ProductSerialNumber    %08X\r\n"+
			"-ManufacturingYear      %02X\r\n"+
			"-ManufacturingMonth     %02X\r\n"+
			"-Always1                %d\r\n"+
			"-CRC                    %02X\r\n"+
			"-------------------------------------\r\n\r\n",
		"attrs.JedecID",         // attrs.JedecID,
		cid.ProductSerialNumber, // serialNumber1,
		0,                       // status1,
		0,                       // status2,
		csd.TRAN_SPEED,          // attrs.MaxClockSpeedMHz,
		false,                   // attrs.HasSectorProtection,
		false,                   // attrs.SupportsFastRead,
		false,                   // attrs.SupportsQSPI,
		false,                   // attrs.SupportsQSPIWrites,
		false,                   // attrs.WriteStatusSplit,
		false,                   // attrs.SingleStatusByte,
		sectors,
		csd.Size(),
		cid.ManufacturerID,
		cid.OEMApplicationID,
		cid.ProductName,
		cid.ProductVersion,
		cid.ProductSerialNumber,
		cid.ManufacturingYear,
		cid.ManufacturingMonth,
		cid.Always1,
		cid.CRC,
	)
}

func erase(argv []string) {
	fmt.Printf("erase - not impl\r\n")
}

var writeBuf [256]byte

func write(argv []string) {
	if len(argv) < 3 {
		println("usage: write <hex offset> <bytes>")
		return
	}
	var err error
	var addr uint64 = 0x0
	if addr, err = strconv.ParseUint(argv[1], 16, 32); err != nil {
		println("Invalid address: " + err.Error() + "\r\n")
		return
	}
	buf := writeBuf[:0]
	for i := 0; i < len(argv[2]); i += 2 {
		var b uint64
		if b, err = strconv.ParseUint(argv[2][i:i+2], 16, 8); err != nil {
			println("Invalid bytes: " + err.Error() + "\r\n")
			return
		}
		buf = append(buf, byte(b))
	}

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
	_, err = dev.ReadAt(buf, int64(addr))
	if err != nil {
		fmt.Printf("xxd err : %s\r\n", err.Error())
	}
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
			if j >= n || !strconv.IsPrint(rune(b[i+j])) || b[i+j] >= 0x80 {
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
