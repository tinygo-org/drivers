package unor4wifi

import (
	"errors"
	"fmt"
	"machine"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	errModemError   = errors.New("unor4wifi: modem returned ERROR")
	errModemTimeout = errors.New("unor4wifi: modem response timeout")
	errModemParse   = errors.New("unor4wifi: modem response parse error")
)

const (
	tokenOK    = "OK"
	tokenError = "ERROR"
)

// modem is the AT transport to the ESP32-S3 bridge firmware.
// The protocol is defined in uno-r4-wifi-usb-bridge/UNOR4USBBridge/commands.h.
// Responses have the form "+CMD:<data>\r\nOK\r\n" or "ERROR\r\n".
// Binary responses have the form "+CMD:<len>|<len bytes>OK\r\n".
type modem struct {
	uart    *machine.UART
	timeout time.Duration
	line    []byte
}

func (m *modem) readByte(deadline time.Time) (byte, error) {
	for m.uart.Buffered() == 0 {
		if time.Now().After(deadline) {
			return 0, errModemTimeout
		}
		runtime.Gosched()
	}
	return m.uart.ReadByte()
}

// readLine reads a line and strips the trailing CR LF.
func (m *modem) readLine(deadline time.Time) (string, error) {
	m.line = m.line[:0]
	for {
		c, err := m.readByte(deadline)
		if err != nil {
			return "", err
		}
		if c == '\n' {
			n := len(m.line)
			if n > 0 && m.line[n-1] == '\r' {
				n--
			}
			return string(m.line[:n]), nil
		}
		m.line = append(m.line, c)
	}
}

func (m *modem) writeString(s string) {
	m.uart.Write([]byte(s))
}

func (m *modem) flushInput() {
	for m.uart.Buffered() > 0 {
		m.uart.ReadByte()
	}
}

// waitResult reads response lines until OK or ERROR.
// Data lines are collected into a single string. Multiple data lines are
// joined with "\r\n".
func (m *modem) waitResult(deadline time.Time) (string, error) {
	var sb strings.Builder
	first := true
	for {
		line, err := m.readLine(deadline)
		if err != nil {
			return sb.String(), err
		}
		if debugging(debugCmd) {
			fmt.Printf("    [<--] %s\r\n", line)
		}
		switch {
		case line == tokenOK:
			return strings.TrimSpace(sb.String()), nil
		case line == tokenError:
			return strings.TrimSpace(sb.String()), errModemError
		case first && len(line) > 0 && line[0] == '+':
			if i := strings.IndexAny(line, ":="); i >= 0 {
				line = line[i+1:]
			} else {
				line = ""
			}
			sb.WriteString(strings.TrimLeft(line, " "))
			first = false
		case first && line == "":
			// Skip empty lines before the prompt.
		default:
			if !first {
				sb.WriteString("\r\n")
			}
			sb.WriteString(line)
			first = false
		}
	}
}

// exec sends a command and returns the response data.
func (m *modem) exec(cmd string) (string, error) {
	return m.execTimeout(cmd, m.timeout)
}

func (m *modem) execTimeout(cmd string, timeout time.Duration) (string, error) {
	m.flushInput()
	if debugging(debugCmd) {
		fmt.Printf("    [-->] %s\r\n", cmd)
	}
	m.writeString(cmd)
	m.writeString("\r\n")
	return m.waitResult(time.Now().Add(timeout))
}

// execInt sends a command and parses the response as an integer.
func (m *modem) execInt(cmd string) (int, error) {
	res, err := m.exec(cmd)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(res)
	if err != nil {
		return 0, errModemParse
	}
	return n, nil
}

// execSend sends a command followed by raw payload bytes and waits for OK.
func (m *modem) execSend(cmd string, payload []byte) error {
	m.flushInput()
	if debugging(debugCmd) {
		fmt.Printf("    [-->] %s, len(payload): %d\r\n", cmd, len(payload))
	}
	m.writeString(cmd)
	m.writeString("\r\n")
	m.uart.Write(payload)
	_, err := m.waitResult(time.Now().Add(m.timeout))
	return err
}

// execSized sends a command whose response carries binary data in the form
// "+CMD:<len>|<len bytes>". The bytes are stored in dst. Returns the count.
func (m *modem) execSized(cmd string, dst []byte) (int, error) {
	m.flushInput()
	if debugging(debugCmd) {
		fmt.Printf("    [-->] %s\r\n", cmd)
	}
	m.writeString(cmd)
	m.writeString("\r\n")
	deadline := time.Now().Add(m.timeout)

	// Find the prompt. A bare ERROR line can arrive instead.
	for {
		c, err := m.readByte(deadline)
		if err != nil {
			return 0, err
		}
		if c == '+' {
			break
		}
		if c == 'E' {
			line, err := m.readLine(deadline)
			if err != nil {
				return 0, err
			}
			if "E"+line == tokenError {
				return 0, errModemError
			}
			return 0, errModemParse
		}
	}
	// Skip the command name up to ':'.
	for {
		c, err := m.readByte(deadline)
		if err != nil {
			return 0, err
		}
		if c == ':' || c == '=' {
			break
		}
	}
	// Read the decimal length up to '|'.
	size := 0
	for {
		c, err := m.readByte(deadline)
		if err != nil {
			return 0, err
		}
		if c == '|' {
			break
		}
		if c < '0' || c > '9' {
			if c == ' ' {
				continue
			}
			return 0, errModemParse
		}
		size = size*10 + int(c-'0')
	}
	n := 0
	for i := 0; i < size; i++ {
		c, err := m.readByte(deadline)
		if err != nil {
			return n, err
		}
		if i < len(dst) {
			dst[i] = c
			n++
		}
	}
	_, err := m.waitResult(deadline)
	return n, err
}
