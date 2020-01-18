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

	tr1  *flash.Transport
	dev1 *flash.Device
	/*
		fatdisk fs.BlockDevice
		fatboot *fat.BootSectorCommon
		fatfs   *fat.FileSystem
		rootdir fs.Directory
		currdir fs.Directory
	*/
	//fatfsys *fat.FAT

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

	tr1 = &flash.Transport{
		SPI:  machine.SPI1,
		MOSI: machine.SPI1_MOSI_PIN,
		MISO: machine.SPI1_MISO_PIN,
		SCK:  machine.SPI1_SCK_PIN,
		SS:   machine.SPI1_CS_PIN,
	}
	tr1.Begin()
	dev1 = &flash.Device{Transport: tr1}

	readyLED.Low()
	write("SPI Configured. Reading flash info")

	dev1.Begin()

	var err error
	if err != nil {
		println("could not decode boot sector: " + err.Error() + "\r\n")
	}

	//mnt(nil)

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
	status, _ := dev1.ReadStatus()
	serialNumber1, _ := dev1.ReadSerialNumber()
	fmt.Printf(
		"\n-------------------------------------\r\n"+
			" Device Information:  \r\n"+
			"-------------------------------------\r\n"+
			" JEDEC ID: %v\r\n"+
			"   Serial: %v\r\n"+
			"   Status: %2x\r\n"+
			"-------------------------------------\r\n\r\n",
		dev1.ID,
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
	dev1.ReadBuffer(uint32(addr), buf)
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

/*
const FlashBlockDeviceSectorSize = 512

type FlashBlockDevice struct {
	flashdev *flash.Device
	buf      []byte
	bufaddr  uint32
	bufvalid bool
}

func (fbd *FlashBlockDevice) Close() error {
	// no-op
	return nil
}

func (fbd *FlashBlockDevice) Len() int64 {
	// hard-coded for now
	return 4096
}

func (fbd *FlashBlockDevice) SectorSize() int {
	// hard-coded for now
	return FlashBlockDeviceSectorSize
}

func (fbd *FlashBlockDevice) ReadAt(p []byte, addr int64) (n int, err error) {

	if debug {
		fmt.Printf(" -- reading %d from %08x", len(p), addr)
	}

	// this is the offset from the start of the first sector that we will read
	offset := addr % FlashBlockDeviceSectorSize

	// this is the address of the start of the first sector
	start := uint32(addr - int64(offset))

	// if a buffer does not already exist, create it and mark it as invalid
	if fbd.buf == nil {
		fbd.buf = make([]byte, FlashBlockDeviceSectorSize)
		fbd.bufvalid = false
	}

	// for the first sector we'll check if it is already cached or not
	if !fbd.bufvalid || start != fbd.bufaddr {
		if debug {
			fmt.Printf(" (not cached)\r\n")
		}
		fbd.bufvalid = false
		fbd.bufaddr = start
		if err = fbd.flashdev.ReadBuffer(fbd.bufaddr, fbd.buf); err != nil {
			return
		}
		fbd.bufvalid = true
	} else if debug {
		fmt.Printf(" (cached)\r\n")
	}

	if debug {
		fmt.Printf("    address: %08x, offset: %d, n before: %d", start, offset, n)
	}

	// copy the first section of bytes into the destination buffer
	n += copy(p[n:], fbd.buf[offset:])
	start += FlashBlockDeviceSectorSize

	if debug {
		fmt.Printf(" - n after: %d\r\n", n)
	}

	// keep looping over subsequent sectors until we've read n bytes
	for c := len(p); n < c; start += FlashBlockDeviceSectorSize {
		if debug {
			fmt.Printf("    address: %08x, n before: %d", start, n)
		}
		fbd.bufvalid = false
		fbd.bufaddr = start
		if err = fbd.flashdev.ReadBuffer(fbd.bufaddr, fbd.buf); err != nil {
			return
		}
		fbd.bufvalid = true
		n += copy(p[n:], fbd.buf)
		if debug {
			fmt.Printf(" - n after: %d\r\n", n)
		}
	}

	return
}

/*
	if err = fbd.flashdev.ReadBuffer(uint32(off), p); err == nil {
		return len(p), nil
	}

func min(a int, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

func (fbd *FlashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("Writes not yet supported")
}

*/
