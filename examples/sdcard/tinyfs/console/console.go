//go:build tinygo
// +build tinygo

package console

import (
	"fmt"
	"io"
	"machine"
	"os"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/tinyfs"
)

const consoleBufLen = 64

const startBlock = 0
const blockCount = 0

var (
	debug = false

	input [consoleBufLen]byte

	console  = machine.Serial
	readyLED = machine.LED

	flashdev tinyfs.BlockDevice
	fs       tinyfs.Filesystem

	currdir = "/"

	commands = map[string]cmdfunc{
		"":        noop,
		"dbg":     dbg,
		"help":    help,
		"lsblk":   lsblk,
		"mount":   mount,
		"umount":  umount,
		"format":  format,
		"xxd":     xxd,
		"ls":      ls,
		"samples": samples,
		"mkdir":   mkdir,
		"cat":     cat,
		"create":  create,
		"write":   write,
		"rm":      rm,
	}
)

type cmdfunc func(argv []string)

const (
	StateInput = iota
	StateEscape
	StateEscBrc
	StateCSI
)

func RunFor(dev tinyfs.BlockDevice, filesys tinyfs.Filesystem) {
	flashdev = dev
	fs = filesys

	readyLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	readyLED.High()

	readyLED.Low()
	println("SPI Configured. Reading flash info")

	/*
		lfsConfig := flashlfs.NewConfig()
		if blockCount == 0 {
			lfsConfig.BlockCount = (flashdev.Attrs().TotalSize / lfsConfig.BlockSize) - startBlock
		} else {
			lfsConfig.BlockCount = blockCount
		}
		println("Start block:", startBlock)
		println("Block count:", lfsConfig.BlockCount)

		blockdev = flashlfs.NewBlockDevice(flashdev, startBlock, lfsConfig.BlockSize)
	*/

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
					if console.Buffered() > 0 {
						data, _ := console.ReadByte()
						if data != 10 {
							println("\r\nunexpected: \r", int(data))
						}
					}
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

func help(argv []string) {
	fmt.Printf("help\r\n")
	fmt.Printf("  show help\r\n")
	fmt.Printf("dbg\r\n")
	fmt.Printf("  toggle debug mode\r\n")
	fmt.Printf("xxd <hex address, ex: 0xA0> <size of hexdump in bytes>\r\n")
	fmt.Printf("  hexdump the specified address\r\n")
	fmt.Printf("ls <target file>\r\n")
	fmt.Printf("  list information\r\n")
	fmt.Printf("samples\r\n")
	fmt.Printf("  write some files in the root directory\r\n")
	fmt.Printf("mkdir <target dir>\r\n")
	fmt.Printf("  create directory\r\n")
	fmt.Printf("cat <target file>\r\n")
	fmt.Printf("  print the contents of file\r\n")
	fmt.Printf("create <target file>\r\n")
	fmt.Printf("  create file\r\n")
	fmt.Printf("write <target file>\r\n")
	fmt.Printf("  write to file (press CTRL-D to exit)\r\n")
	fmt.Printf("rm\r\n")
}

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
	fmt.Printf("lsblk : not implement\r\n")
}

func mount(argv []string) {
	if err := fs.Mount(); err != nil {
		println("Could not mount LittleFS filesystem: " + err.Error() + "\r\n")
	} else {
		println("Successfully mounted LittleFS filesystem.\r\n")
	}
}

func format(argv []string) {
	if err := fs.Format(); err != nil {
		println("Could not format LittleFS filesystem: " + err.Error() + "\r\n")
	} else {
		println("Successfully formatted LittleFS filesystem.\r\n")
	}
}

func umount(argv []string) {
	if err := fs.Unmount(); err != nil {
		println("Could not unmount LittleFS filesystem: " + err.Error() + "\r\n")
	} else {
		println("Successfully unmounted LittleFS filesystem.\r\n")
	}
}

/*
	var err error
	if fatfs == nil {
		fatfs, err = fat.New(fatdisk)
		if err != nil {
			fatfs = nil
			println("could not load FAT filesystem: " + err.Error() + "\r\n")
		}
		fmt.Printf("loaded fs\r\n")
	}
	if rootdir == nil {
		rootdir, err = fatfs.RootDir()
		if err != nil {
			rootdir = nil
			println("could not load rootdir: " + err.Error() + "\r\n")
		}
		fmt.Printf("loaded rootdir\r\n")
	}
	if currdir == nil {
		currdir = rootdir
	}
*/

func ls(argv []string) {
	path := "/"
	if len(argv) > 1 {
		path = strings.TrimSpace(argv[1])
	}
	dir, err := fs.Open(path)
	if err != nil {
		fmt.Printf("Could not open directory %s: %v\r\n", path, err)
		return
	}
	defer dir.Close()
	infos, err := dir.Readdir(0)
	_ = infos
	if err != nil {
		fmt.Printf("Could not read directory %s: %v\r\n", path, err)
		return
	}
	for _, info := range infos {
		s := "-rwxrwxrwx"
		if info.IsDir() {
			s = "drwxrwxrwx"
		}
		fmt.Printf("%s %5d %s\r\n", s, info.Size(), info.Name())
	}
}

func mkdir(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying mkdir to " + tgt)
	}
	if tgt == "" {
		println("Usage: mkdir <target dir>")
		return
	}
	err := fs.Mkdir(tgt, 0777)
	if err != nil {
		println("Could not mkdir " + tgt + ": " + err.Error())
	}
}

func rm(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying rm to " + tgt)
	}
	if tgt == "" {
		println("Usage: rm <target dir>")
		return
	}
	err := fs.Remove(tgt)
	if err != nil {
		println("Could not rm " + tgt + ": " + err.Error())
	}
}

func samples(argv []string) {
	buf := make([]byte, 90)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("file%d.txt", i)
		if bytes, err := createSampleFile(name, buf); err != nil {
			fmt.Printf("%s\r\n", err)
			return
		} else {
			fmt.Printf("wrote %d bytes to %s\r\n", bytes, name)
		}
	}
}

func create(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying create to " + tgt)
	}
	buf := make([]byte, 90)
	if bytes, err := createSampleFile(tgt, buf); err != nil {
		fmt.Printf("%s\r\n", err)
		return
	} else {
		fmt.Printf("wrote %d bytes to %s\r\n", bytes, tgt)
	}
}

func write(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying receive to " + tgt)
	}
	buf := make([]byte, 1)
	f, err := fs.OpenFile(tgt, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		fmt.Printf("error opening %s: %s\r\n", tgt, err.Error())
		return
	}
	defer f.Close()
	var n int
	for {
		if console.Buffered() > 0 {
			data, _ := console.ReadByte()
			switch data {
			case 0x04:
				fmt.Printf("wrote %d bytes to %s\r\n", n, tgt)
				return
			default:
				// anything else, just echo the character if it is printable
				if strconv.IsPrint(rune(data)) {
					console.WriteByte(data)
				}
				buf[0] = data
				if _, err := f.Write(buf); err != nil {
					fmt.Printf("\nerror writing: %s\r\n", err)
					return
				}
				n++
			}
		}
	}
}

func createSampleFile(name string, buf []byte) (int, error) {
	for j := uint8(0); j < uint8(len(buf)); j++ {
		buf[j] = 0x20 + j
	}
	f, err := fs.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return 0, fmt.Errorf("error opening %s: %s", name, err.Error())
	}
	defer f.Close()
	bytes, err := f.Write(buf)
	if err != nil {
		return 0, fmt.Errorf("error writing %s: %s", name, err.Error())
	}
	return bytes, nil
}

/*
func cd(argv []string) {

	if fatfs == nil || rootdir == nil {
		mnt(nil)
	}
	if len(argv) == 1 {
		currdir = rootdir
		return
	}
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying to cd to " + tgt)
	}
	if tgt == "" {
		println("Usage: cd <target dir>")
		return
	}
	if debug {
		println("Getting entry")
	}
	entry := currdir.Entry(tgt)
	if entry == nil {
		println("File not found: " + tgt)
		return
	}
	if !entry.IsDir() {
		println("Not a directory: " + tgt)
		return
	}
	if debug {
		println("Getting dir")
	}
	cd, err := entry.Dir()
	if err != nil {
		println("Could not cd to " + tgt + ": " + err.Error())
	}
	currdir = cd
}
*/

func cat(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying to cat to " + tgt)
	}
	if tgt == "" {
		println("Usage: cat <target file>")
		return
	}
	if debug {
		println("Getting entry")
	}
	f, err := fs.Open(tgt)
	if err != nil {
		println("Could not open: " + err.Error())
		return
	}
	defer f.Close()
	if f.IsDir() {
		println("Not a file: " + tgt)
		return
	}
	off := 0x0
	buf := make([]byte, 64)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			println("Error reading " + tgt + ": " + err.Error())
			time.Sleep(1 * time.Second)
		}
		xxdfprint(os.Stdout, uint32(off), buf[:n])
		off += n
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
		if size > 512 || size < 1 {
			fmt.Printf("Size of hexdump must be greater than 0 and less than %d\r\n", 512)
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
	buf := make([]byte, size)
	//bsz := uint64(flash.SectorSize)
	//blockdev.ReadBlock(uint32(addr/bsz), uint32(addr%bsz), buf)
	flashdev.ReadAt(buf, int64(addr))
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
