//go:build tinygo

// Package e220 implements a driver for the e220 LoRa module.
//
// Datasheet(E220-900T22S JP): https://dragon-torch.tech/wp-content/uploads/2022/08/data_sheet.pdf
// Datasheet(E220-900T2S): https://www.ebyte.com/en/downpdf.aspx?id=1211
// Datasheet(LLCC68): https://semtech.my.salesforce.com/sfc/p/#E0000000JelG/a/2R000000HTJR/Tem0gUxGfOZ2Qn3bUzmV2zKNQRYJ3bpobPfOQ7B.erE
//
// LoraWAN 1.1 Specification: https://hz137b.p3cdn1.secureserver.net/wp-content/uploads/2020/11/lorawantm_specification_-v1.1.pdf?time=1673563697
package e220

import (
	"bytes"
	"context"
	"fmt"
	"machine"
	"runtime"
	"strings"
	"time"
)

// targetUARTbaud represents the UART baud rate of the target board
const (
	TargetUARTBaud1200kbps   = 1200
	TargetUARTBaud2400kbps   = 2400
	TargetUARTBaud4800kbps   = 4800
	TargetUARTBaud9600kbps   = 9600
	TargetUARTBaud19200kbps  = 19200
	TargetUARTBaud38400kbps  = 38400
	TargetUARTBaud57600kbps  = 57600
	TargetUARTBaud115200kbps = 115200
)

// Mode represents the operating mode of the E220
const (
	Mode0 uint8 = iota
	Mode1
	Mode2
	Mode3
)

// Device wraps E220 LoRa module
type Device struct {
	uart               *machine.UART
	m0                 machine.Pin
	m1                 machine.Pin
	tx                 machine.Pin
	rx                 machine.Pin
	aux                machine.Pin
	targetUARTBaudRate uint32

	parameters []byte
	payload    []byte

	txAddr   uint16
	txCh     uint8
	txMethod uint8
}

const (
	writingRegisterSize = 8
	bufferSize          = 400
)

// New returns a new E220 driver. Pass UART object and each used machine.Pin
func New(
	uart *machine.UART,
	m0, m1, tx, rx, aux machine.Pin,
	targetUARTBaudRate uint32,
) *Device {
	d := &Device{
		uart:               uart,
		m0:                 m0,
		m1:                 m1,
		tx:                 tx,
		rx:                 rx,
		aux:                aux,
		targetUARTBaudRate: targetUARTBaudRate,
		parameters:         make([]byte, writingRegisterSize),
		payload:            make([]byte, bufferSize),
		// Set to factory setting
		txAddr:   0xFFFF,
		txCh:     15,
		txMethod: TxMethodTransparent,
	}
	return d
}

// Configure sets up the device for communication
func (d *Device) Configure(startupE220Mode uint8) error {
	d.uart.Configure(machine.UARTConfig{BaudRate: d.targetUARTBaudRate, TX: d.tx, RX: d.rx})
	d.m0.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.m1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.aux.Configure(machine.PinConfig{Mode: machine.PinInput})

	err := d.SetMode(startupE220Mode)
	if err != nil {
		return err
	}

	// AUX Pin becomes Low Level for about 30ms after the E220 is started.
	// therefore, wait until transmission/reception becomes possible.
	time.Sleep(50 * time.Millisecond)

	return nil
}

// SetMode sets operation mode of E220
func (d *Device) SetMode(mode uint8) error {
	switch mode {
	case Mode0:
		d.m1.Low()
		d.m0.Low()
	case Mode1:
		d.m1.Low()
		d.m0.High()
	case Mode2:
		d.m1.High()
		d.m0.Low()
	case Mode3:
		d.m1.High()
		d.m0.High()
	default:
		return fmt.Errorf("Invalid mode: %d", mode)
	}

	return nil
}

const (
	writeCmd = 0xC0
	readCmd  = 0xC1
)

// ReadRegister reads and returns length register values from startAddr
func (d *Device) ReadRegister(startAddr uint8, length uint8) ([]byte, error) {
	response, err := d.writeRegister(readCmd, startAddr, length, []byte{})
	if err != nil {
		return nil, err
	}
	// response must start at 0xC1 and be the length requested
	if (len(response) != int(length+3)) || (response[0] != 0xC1) {
		return nil, fmt.Errorf("unexpected response: %X", response)
	}

	return response[3:], nil
}

// WriteRegister writes the register value from startAddr to the value of params
func (d *Device) WriteRegister(startAddr uint8, params []byte) error {
	response, err := d.writeRegister(writeCmd, startAddr, uint8(len(params)), params)
	if err != nil {
		return err
	}
	// response must start with C1 and the subsequent values must be the same as the requested values
	if (response[0] != 0xC1) || ((response[1]) != startAddr) ||
		(int(response[2]) != len(params)) || !bytes.Equal(response[3:], params) {
		return fmt.Errorf("unexpected response: want=%X%X got=%X",
			[]byte{0xC1, startAddr, byte(len(params))}, params, response)
	}

	return nil
}

func (d *Device) writeRegister(cmd, startAddr, length uint8, params []byte) ([]byte, error) {
	d.payload[0] = cmd
	d.payload[1] = startAddr
	d.payload[2] = length
	if len(params) > writingRegisterSize {
		return nil, fmt.Errorf("params must be greater than or equal to %d: got=%d", writingRegisterSize, len(params))
	}
	copy(d.payload[3:], params)
	writeLength := uint8(3)
	if cmd == writeCmd {
		writeLength += length
	}
	// Wait for device is ready
	ctxWrite, cancelWrite := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancelWrite()
	for !d.IsReady() {
		select {
		case <-ctxWrite.Done():
			return nil, fmt.Errorf("writing register timed out")
		default:
		}
		runtime.Gosched()
	}
	_, err := d.uart.Write(d.payload[:writeLength])
	if err != nil {
		return nil, err
	}

	ctxRead, cancelRead := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancelRead()
	readIndex := 0
	for {
		if d.uart.Buffered() > 0 {
			d.payload[readIndex], err = d.uart.ReadByte()
			if err != nil {
				return nil, err
			}
			readIndex++
			if readIndex == 3+int(length) {
				break
			}
		}
		select {
		case <-ctxRead.Done():
			return nil, fmt.Errorf("waiting for response from E220 timed out")
		default:
		}
		runtime.Gosched()
	}
	// Wait for a while because the next register access
	// cannot be received immediately after register access
	time.Sleep(50 * time.Millisecond)

	return d.payload[:readIndex], nil
}

// WriteConfig writes configuration values to E220
func (d *Device) WriteConfig(cfg Config) error {

	cfg.Validate()
	errors := cfg.Errors()
	if len(errors) > 0 {
		s := make([]string, 0, 8)
		for _, e := range errors {
			s = append(s, e.Error())
		}
		return fmt.Errorf("configuration parameter validation failed:\n%s", strings.Join(s, "\n"))
	}
	err := cfg.paramsToBytes(&d.parameters)
	if err != nil {
		return err
	}

	return d.WriteRegister(0x00, d.parameters)
}

// ReadConfig reads configuration values from E220
func (d *Device) ReadConfig() (*Config, error) {
	// "+1" is read only register (addr = 0x08)
	registerBytes, err := d.ReadRegister(0x00, writingRegisterSize+1)
	if err != nil {
		return nil, err
	}
	config := Config{}
	err = config.bytesToParams(registerBytes)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// IsReady returns whether the device is ready for new commands
// This API returns false in the following cases.
//   - Self checking
//   - TX/RX buffer is not empty
//   - In mode3 or sleep mode
//   - 1 ms has not passed since the operation mode was changed from mode3 (or sleep mode) to another mode
func (d *Device) IsReady() bool {
	return d.aux.Get()
}

// TxMethod represents transmit method of E220
// This value should be set to the same value that the user set for the E220.
const (
	TxMethodTransparent = iota
	TxMethodFixed
)

// SetTxInfo sets information such as the destination address
func (d *Device) SetTxInfo(addr uint16, ch uint8, txMethod uint8) error {
	d.txAddr = addr
	d.txCh = ch
	if (txMethod != TxMethodTransparent) && (txMethod != TxMethodFixed) {
		return fmt.Errorf("invalid Transmit Method: %d", d.txMethod)
	}
	d.txMethod = txMethod

	return nil
}

// Write sends msg to the device corresponding to addr and ch
func (d *Device) Write(p []byte) (n int, err error) {
	var offset int
	if d.txMethod == TxMethodTransparent {
		offset = 0
	} else if d.txMethod == TxMethodFixed {
		d.payload[0] = byte((d.txAddr & 0xFF00) >> 8)
		d.payload[1] = byte((d.txAddr & 0x00FF) >> 0)
		d.payload[2] = d.txCh
		offset = 3
	} else {
		return 0, fmt.Errorf("invalid transmit method: %d", d.txMethod)
	}

	if len(p) > (bufferSize - offset) {
		return 0, fmt.Errorf("p must be %dbytes or less: got=%d", bufferSize-offset, len(p))
	}
	copy(d.payload[offset:], p)

	// Wait for device is ready
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	for !d.IsReady() {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("transmitting timed out")
		default:
		}
		runtime.Gosched()
	}
	return d.uart.Write(d.payload[:offset+len(p)])
}

// Buffered returns the number of bytes of data received from E220
func (d *Device) Buffered() int {
	return d.uart.Buffered()
}

// ReadByte reads the value received from E220
func (d *Device) ReadByte() (byte, error) {
	data, err := d.uart.ReadByte()
	if err != nil {
		return 0, err
	}
	return data, nil
}

// Read reads the value received from E220
func (d *Device) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	size := d.uart.Buffered()
	for size == 0 {
		runtime.Gosched()
		size = d.uart.Buffered()
	}

	if size > len(p) {
		size = len(p)
	}
	return d.uart.Read(p[:size])
}
