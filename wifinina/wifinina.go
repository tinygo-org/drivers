// Package wifinina implements TCP wireless communication over SPI
// with an attached separate ESP32 board using the Arduino WiFiNINA protocol.
//
// In order to use this driver, the ESP32 must be flashed with specific firmware from Arduino.
// For more information: https://github.com/arduino/nina-fw
package wifinina // import "tinygo.org/x/drivers/wifinina"

import (
	"encoding/binary"
	"encoding/hex"
	"fmt" // used only in debug printouts and is optimized out when debugging is disabled
	"strconv"
	"strings"
	"sync"
	"time"

	"machine"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/net"
)

const (
	MaxSockets  = 4
	MaxNetworks = 10
	MaxAttempts = 10

	MaxLengthSSID   = 32
	MaxLengthWPAKey = 63
	MaxLengthWEPKey = 13

	LengthMacAddress = 6
	LengthIPV4       = 4

	WlFailure = -1
	WlSuccess = 1
)

const (
	FlagCmd   = 0
	FlagReply = 1 << 7
	FlagData  = 0x40

	NinaCmdPos      = 1
	NinaParamLenPos = 2

	dummyData = 0xFF
)

const (
	ProtoModeTCP = iota
	ProtoModeUDP
	ProtoModeTLS
	ProtoModeMul
)

type IPAddress string // TODO: does WiFiNINA support ipv6???

func (addr IPAddress) String() string {
	if len(addr) < 4 {
		return ""
	}
	return strconv.Itoa(int(addr[0])) + "." + strconv.Itoa(int(addr[1])) + "." + strconv.Itoa(int(addr[2])) + "." + strconv.Itoa(int(addr[3]))
}

func ParseIPv4(s string) (IPAddress, error) {
	v := strings.Split(s, ".")
	v0, _ := strconv.Atoi(v[0])
	v1, _ := strconv.Atoi(v[1])
	v2, _ := strconv.Atoi(v[2])
	v3, _ := strconv.Atoi(v[3])
	return IPAddress([]byte{byte(v0), byte(v1), byte(v2), byte(v3)}), nil
}

func (addr IPAddress) AsUint32() uint32 {
	if len(addr) < 4 {
		return 0
	}
	b := []byte(string(addr))
	return binary.BigEndian.Uint32(b[0:4])
}

type MACAddress uint64

func (addr MACAddress) String() string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(addr))
	encoded := hex.EncodeToString(b)
	result := ""
	for i := 2; i < 8; i++ {
		result += encoded[i*2 : i*2+2]
		if i < 7 {
			result += ":"
		}
	}
	return result
}

// Cmd Struct Message */
// ._______________________________________________________________________.
// | START CMD | C/R  | CMD  | N.PARAM | PARAM LEN | PARAM  | .. | END CMD |
// |___________|______|______|_________|___________|________|____|_________|
// |   8 bit   | 1bit | 7bit |  8bit   |   8bit    | nbytes | .. |   8bit  |
// |___________|______|______|_________|___________|________|____|_________|
type command struct {
	cmd       uint8
	reply     bool
	params    []int
	paramData []byte
}

type Device struct {
	SPI   drivers.SPI
	CS    machine.Pin
	ACK   machine.Pin
	GPIO0 machine.Pin
	RESET machine.Pin

	buf   [64]byte
	ssids [10]string

	sock    uint8
	readBuf readBuffer

	proto uint8
	ip    uint32
	port  uint16
	mu    sync.Mutex

	// ResetIsHigh controls if the RESET signal to the processor
	// should be High or Low (the default). Set this to true
	// before calling Configure() for boards such as the Arduino MKR 1010,
	// where the reset signal needs to go high instead of low.
	ResetIsHigh bool
}

// New returns a new Wifinina device.
func New(bus drivers.SPI, csPin, ackPin, gpio0Pin, resetPin machine.Pin) *Device {
	return &Device{
		SPI:   bus,
		CS:    csPin,
		ACK:   ackPin,
		GPIO0: gpio0Pin,
		RESET: resetPin,
	}
}

// Configure sets the needed pin settings and performs a reset
// of the WiFi device.
func (d *Device) Configure() {
	net.UseDriver(d)
	pinUseDevice(d)

	d.CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.ACK.Configure(machine.PinConfig{Mode: machine.PinInput})
	d.RESET.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.GPIO0.Configure(machine.PinConfig{Mode: machine.PinOutput})

	d.GPIO0.High()
	d.CS.High()

	d.RESET.Set(d.ResetIsHigh)
	time.Sleep(10 * time.Millisecond)

	d.RESET.Set(!d.ResetIsHigh)
	time.Sleep(750 * time.Millisecond)

	d.GPIO0.Low()
	d.GPIO0.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}

// ----------- client methods (should this be a separate struct?) ------------

func (d *Device) StartClient(hostname string, addr uint32, port uint16, sock uint8, mode uint8) error {
	if _debug {
		fmt.Printf("[StartClient] hostname: %s addr: %02X, port: %d, sock: %d\r\n", hostname, addr, port, sock)
	}
	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return err
	}

	if len(hostname) > 0 {
		d.sendCmd(CmdStartClientTCP, 5)
		d.sendParamStr(hostname, false)
	} else {
		d.sendCmd(CmdStartClientTCP, 4)
	}
	d.sendParam32(addr, false)
	d.sendParam16(port, false)
	d.sendParam8(sock, false)
	d.sendParam8(mode, true)

	if len(hostname) > 0 {
		d.padTo4(17 + len(hostname))
	}

	d.spiChipDeselect()

	_, err := d.waitRspCmd1(CmdStartClientTCP)
	return err
}

func (d *Device) GetSocket() (uint8, error) {
	return d.getUint8(d.req0(CmdGetSocket))
}

func (d *Device) GetClientState(sock uint8) (uint8, error) {
	return d.getUint8(d.reqUint8(CmdGetClientStateTCP, sock))
}

func (d *Device) SendData(buf []byte, sock uint8) (uint16, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return 0, err
	}
	l := d.sendCmd(CmdSendDataTCP, 2)
	l += d.sendParamBuf([]byte{sock}, false)
	l += d.sendParamBuf(buf, true)
	d.addPadding(l)
	d.spiChipDeselect()
	return d.getUint16(d.waitRspCmd1(CmdSendDataTCP))
}

func (d *Device) CheckDataSent(sock uint8) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var lastErr error
	for timeout := 0; timeout < 10; timeout++ {
		sent, err := d.getUint8(d.reqUint8(CmdDataSentTCP, sock))
		if err != nil {
			lastErr = err
		}
		if sent > 0 {
			return true, nil
		}
		time.Sleep(100 * time.Microsecond)
	}
	return false, lastErr
}

func (d *Device) GetDataBuf(sock uint8, buf []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return 0, err
	}
	p := uint16(len(buf))
	l := d.sendCmd(CmdGetDatabufTCP, 2)
	l += d.sendParamBuf([]byte{sock}, false)
	l += d.sendParamBuf([]byte{uint8(p & 0x00FF), uint8((p) >> 8)}, true)
	d.addPadding(l)
	d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return 0, err
	}
	n, err := d.waitRspBuf16(CmdGetDatabufTCP, buf)
	d.spiChipDeselect()
	return int(n), err
}

func (d *Device) StopClient(sock uint8) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _debug {
		println("[StopClient] called StopClient()\r")
	}
	_, err := d.getUint8(d.reqUint8(CmdStopClientTCP, sock))
	return err
}

func (d *Device) StartServer(port uint16, sock uint8, mode uint8) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return err
	}
	l := d.sendCmd(CmdStartServerTCP, 3)
	l += d.sendParam16(port, false)
	l += d.sendParam8(sock, false)
	l += d.sendParam8(mode, true)
	d.addPadding(l)
	d.spiChipDeselect()
	_, err := d.waitRspCmd1(CmdStartClientTCP)
	return err
}

// InsertDataBuf adds data to the buffer used for sending UDP data
func (d *Device) InsertDataBuf(buf []byte, sock uint8) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return false, err
	}
	l := d.sendCmd(CmdInsertDataBuf, 2)
	l += d.sendParamBuf([]byte{sock}, false)
	l += d.sendParamBuf(buf, true)
	d.addPadding(l)
	d.spiChipDeselect()
	n, err := d.getUint8(d.waitRspCmd1(CmdInsertDataBuf))
	return n == 1, err
}

// SendUDPData sends the data previously added to the UDP buffer
func (d *Device) SendUDPData(sock uint8) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return false, err
	}
	l := d.sendCmd(CmdSendDataUDP, 1)
	l += d.sendParam8(sock, true)
	d.addPadding(l)
	d.spiChipDeselect()
	n, err := d.getUint8(d.waitRspCmd1(CmdSendDataUDP))
	return n == 1, err
}

// ---------- /client methods (should this be a separate struct?) ------------

/*
	static bool startServer(uint16_t port, uint8_t sock);
	static uint8_t getServerState(uint8_t sock);
	static bool getData(uint8_t connId, uint8_t *data, bool peek, bool* connClose);
	static int getDataBuf(uint8_t connId, uint8_t *buf, uint16_t bufSize);
	static bool sendData(uint8_t sock, const uint8_t *data, uint16_t len);
	static bool sendDataUdp(uint8_t sock, const char* host, uint16_t port, const uint8_t *data, uint16_t len);
	static uint16_t availData(uint8_t connId);


	static bool ping(const char *host);
	static void reset();

	static void getRemoteIpAddress(IPAddress& ip);
	static uint16_t getRemotePort();
*/

func (d *Device) Disconnect() error {
	_, err := d.req1(CmdDisconnect)
	return err
}

func (d *Device) GetFwVersion() (string, error) {
	return d.getString(d.req0(CmdGetFwVersion))
}

func (d *Device) GetConnectionStatus() (ConnectionStatus, error) {
	status, err := d.getUint8(d.req0(CmdGetConnStatus))
	return ConnectionStatus(status), err
}

func (d *Device) GetCurrentEncryptionType() (EncryptionType, error) {
	enctype, err := d.getUint8(d.req1(CmdGetCurrEncrType))
	return EncryptionType(enctype), err
}

func (d *Device) GetCurrentBSSID() (MACAddress, error) {
	return d.getMACAddress(d.req1(CmdGetCurrBSSID))
}

func (d *Device) GetCurrentRSSI() (int32, error) {
	return d.getInt32(d.req1(CmdGetCurrRSSI))
}

func (d *Device) GetCurrentSSID() (string, error) {
	return d.getString(d.req1(CmdGetCurrSSID))
}

func (d *Device) GetMACAddress() (MACAddress, error) {
	return d.getMACAddress(d.req1(CmdGetMACAddr))
}

func (d *Device) GetIP() (ip, subnet, gateway IPAddress, err error) {
	sl := make([]string, 3)
	if l, err := d.reqRspStr1(CmdGetIPAddr, dummyData, sl); err != nil {
		return "", "", "", err
	} else if l != 3 {
		return "", "", "", ErrUnexpectedLength
	}
	return IPAddress(sl[0]), IPAddress(sl[1]), IPAddress(sl[2]), err
}

func (d *Device) GetHostByName(hostname string) (IPAddress, error) {
	ok, err := d.getUint8(d.reqStr(CmdReqHostByName, hostname))
	if err != nil {
		return "", err
	}
	if ok != 1 {
		return "", ErrUnknownHost
	}
	ip, err := d.getString(d.req0(CmdGetHostByName))
	return IPAddress(ip), err
}

func (d *Device) GetNetworkBSSID(idx int) (MACAddress, error) {
	if idx < 0 || idx >= MaxNetworks {
		return 0, nil
	}
	return d.getMACAddress(d.reqUint8(CmdGetIdxBSSID, uint8(idx)))
}

func (d *Device) GetNetworkChannel(idx int) (uint8, error) {
	if idx < 0 || idx >= MaxNetworks {
		return 0, nil
	}
	return d.getUint8(d.reqUint8(CmdGetIdxChannel, uint8(idx)))
}

func (d *Device) GetNetworkEncrType(idx int) (EncryptionType, error) {
	if idx < 0 || idx >= MaxNetworks {
		return 0, nil
	}
	enctype, err := d.getUint8(d.reqUint8(CmdGetIdxEncrType, uint8(idx)))
	return EncryptionType(enctype), err
}

func (d *Device) GetNetworkRSSI(idx int) (int32, error) {
	if idx < 0 || idx >= MaxNetworks {
		return 0, nil
	}
	return d.getInt32(d.reqUint8(CmdGetIdxRSSI, uint8(idx)))
}

func (d *Device) GetNetworkSSID(idx int) string {
	if idx < 0 || idx >= MaxNetworks {
		return ""
	}
	return d.ssids[idx]
}

func (d *Device) GetReasonCode() (uint8, error) {
	return d.getUint8(d.req0(CmdGetReasonCode))
}

// GetTime is the time as a Unix timestamp
func (d *Device) GetTime() (uint32, error) {
	return d.getUint32(d.req0(CmdGetTime))
}

func (d *Device) GetTemperature() (float32, error) {
	return d.getFloat32(d.req0(CmdGetTemperature))
}

func (d *Device) Ping(ip IPAddress, ttl uint8) int16 {
	return 0
}

func (d *Device) SetDebug(on bool) error {
	var v uint8
	if on {
		v = 1
	}
	_, err := d.reqUint8(CmdSetDebug, v)
	return err
}

func (d *Device) SetNetwork(ssid string) error {
	_, err := d.reqStr(CmdSetNet, ssid)
	return err
}

func (d *Device) SetPassphrase(ssid string, passphrase string) error {
	_, err := d.reqStr2(CmdSetPassphrase, ssid, passphrase)
	return err
}

func (d *Device) SetKey(ssid string, index uint8, key string) error {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}

	d.sendCmd(CmdSetKey, 3)
	d.sendParamStr(ssid, false)
	d.sendParam8(index, false)
	d.sendParamStr(key, true)

	d.padTo4(8 + len(ssid) + len(key))

	_, err := d.waitRspCmd1(CmdSetKey)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) SetNetworkForAP(ssid string) error {
	_, err := d.reqStr(CmdSetAPNet, ssid)
	return err
}

func (d *Device) SetPassphraseForAP(ssid string, passphrase string) error {
	_, err := d.reqStr2(CmdSetAPPassphrase, ssid, passphrase)
	return err
}

func (d *Device) SetIP(which uint8, ip uint32, gw uint32, subnet uint32) error {
	return ErrNotImplemented
}

func (d *Device) SetDNS(which uint8, dns1 uint32, dns2 uint32) error {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}

	d.sendCmd(CmdSetDNSConfig, 3)
	d.sendParam8(which, false)
	d.sendParam32(dns1, false)
	d.sendParam32(dns2, true)

	_, err := d.waitRspCmd1(CmdSetDNSConfig)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) SetHostname(hostname string) error {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}

	d.sendCmd(CmdSetHostname, 3)
	d.sendParamStr(hostname, true)

	d.padTo4(5 + len(hostname))

	_, err := d.waitRspCmd1(CmdSetHostname)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) SetPowerMode(mode uint8) error {
	_, err := d.reqUint8(CmdSetPowerMode, mode)
	return err
}

func (d *Device) ScanNetworks() (uint8, error) {
	return d.reqRspStr0(CmdScanNetworks, d.ssids[:])
}

func (d *Device) StartScanNetworks() (uint8, error) {
	return d.getUint8(d.req0(CmdStartScanNetworks))
}

func (d *Device) PinMode(pin uint8, mode uint8) error {
	_, err := d.req2Uint8(CmdSetPinMode, pin, mode)
	return err
}

func (d *Device) DigitalWrite(pin uint8, value uint8) error {
	_, err := d.req2Uint8(CmdSetDigitalWrite, pin, value)
	return err
}

func (d *Device) AnalogWrite(pin uint8, value uint8) error {
	_, err := d.req2Uint8(CmdSetAnalogWrite, pin, value)
	return err
}

// ------------- End of public device interface ----------------------------

func (d *Device) getString(l uint8, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return string(d.buf[0:l]), err
}

func (d *Device) getUint8(l uint8, err error) (uint8, error) {
	if err != nil {
		return 0, err
	}
	if l != 1 {
		if _debug {
			println("expected length 1, was actually", l, "\r")
		}
		return 0, ErrUnexpectedLength
	}
	return d.buf[0], err
}

func (d *Device) getUint16(l uint8, err error) (uint16, error) {
	if err != nil {
		return 0, err
	}
	if l != 2 {
		if _debug {
			println("expected length 2, was actually", l, "\r")
		}
		return 0, ErrUnexpectedLength
	}
	return binary.BigEndian.Uint16(d.buf[0:2]), err
}

func (d *Device) getUint32(l uint8, err error) (uint32, error) {
	if err != nil {
		return 0, err
	}
	if l != 4 {
		return 0, ErrUnexpectedLength
	}
	return binary.LittleEndian.Uint32(d.buf[0:4]), err
}

func (d *Device) getInt32(l uint8, err error) (int32, error) {
	i, err := d.getUint32(l, err)
	return int32(i), err
}

func (d *Device) getFloat32(l uint8, err error) (float32, error) {
	i, err := d.getUint32(l, err)
	return float32(i), err
}

func (d *Device) getMACAddress(l uint8, err error) (MACAddress, error) {
	if err != nil {
		return 0, err
	}
	if l != 6 {
		return 0, ErrUnexpectedLength
	}
	return MACAddress(binary.LittleEndian.Uint64(d.buf[0:8]) & 0xFFFFFFFFFFFF), err
}

// req0 sends a command to the device with no request parameters
func (d *Device) req0(cmd CommandType) (l uint8, err error) {
	if err := d.sendCmd0(cmd); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// req1 sends a command to the device with a single dummy parameters of 0xFF
func (d *Device) req1(cmd CommandType) (l uint8, err error) {
	return d.reqUint8(cmd, dummyData)
}

// reqUint8 sends a command to the device with a single uint8 parameter
func (d *Device) reqUint8(cmd CommandType, data uint8) (l uint8, err error) {
	if err := d.sendCmdPadded1(cmd, data); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// req2Uint8 sends a command to the device with two uint8 parameters
func (d *Device) req2Uint8(cmd CommandType, p1, p2 uint8) (l uint8, err error) {
	if err := d.sendCmdPadded2(cmd, p1, p2); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// reqStr sends a command to the device with a single string parameter
func (d *Device) reqStr(cmd CommandType, p1 string) (uint8, error) {
	if err := d.sendCmdStr(cmd, p1); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// reqStr sends a command to the device with 2 string parameters
func (d *Device) reqStr2(cmd CommandType, p1 string, p2 string) (uint8, error) {
	if err := d.sendCmdStr2(cmd, p1, p2); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// reqStrRsp0 sends a command passing a string slice for the response
func (d *Device) reqRspStr0(cmd CommandType, sl []string) (l uint8, err error) {
	if err := d.sendCmd0(cmd); err != nil {
		return 0, err
	}
	defer d.spiChipDeselect()
	if err = d.waitForChipSelect(); err != nil {
		return
	}
	return d.waitRspStr(cmd, sl)
}

// reqStrRsp1 sends a command with a uint8 param and a string slice for the response
func (d *Device) reqRspStr1(cmd CommandType, data uint8, sl []string) (uint8, error) {
	if err := d.sendCmdPadded1(cmd, data); err != nil {
		return 0, err
	}
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return 0, err
	}
	return d.waitRspStr(cmd, sl)
}

func (d *Device) sendCmd0(cmd CommandType) error {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}
	d.sendCmd(cmd, 0)
	return nil
}

func (d *Device) sendCmdPadded1(cmd CommandType, data uint8) error {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}
	d.sendCmd(cmd, 1)
	d.sendParam8(data, true)
	d.SPI.Transfer(dummyData)
	d.SPI.Transfer(dummyData)
	return nil
}

func (d *Device) sendCmdPadded2(cmd CommandType, data1, data2 uint8) error {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}
	l := d.sendCmd(cmd, 1)
	l += d.sendParam8(data1, false)
	l += d.sendParam8(data2, true)
	d.SPI.Transfer(dummyData)
	return nil
}

func (d *Device) sendCmdStr(cmd CommandType, p1 string) (err error) {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}
	l := d.sendCmd(cmd, 1)
	l += d.sendParamStr(p1, true)
	d.padTo4(5 + len(p1))
	return nil
}

func (d *Device) sendCmdStr2(cmd CommandType, p1 string, p2 string) (err error) {
	defer d.spiChipDeselect()
	if err := d.waitForChipSelect(); err != nil {
		return err
	}
	d.sendCmd(cmd, 2)
	d.sendParamStr(p1, false)
	d.sendParamStr(p2, true)
	d.padTo4(6 + len(p1) + len(p2))
	return nil
}

func (d *Device) waitRspCmd1(cmd CommandType) (l uint8, err error) {
	defer d.spiChipDeselect()
	if err = d.waitForChipSelect(); err != nil {
		return
	}
	return d.waitRspCmd(cmd, 1)
}

func (d *Device) sendCmd(cmd CommandType, numParam uint8) (l int) {
	if _debug {
		fmt.Printf(
			"sendCmd: %s %s(%02X) %02X",
			CmdStart, cmd, uint8(cmd) & ^(uint8(FlagReply)), numParam)
	}
	l = 3
	d.SPI.Transfer(byte(CmdStart))
	d.SPI.Transfer(uint8(cmd) & ^(uint8(FlagReply)))
	d.SPI.Transfer(numParam)
	if numParam == 0 {
		d.SPI.Transfer(byte(CmdEnd))
		l += 1
		if _debug {
			fmt.Printf(" %s", CmdEnd)
		}
	}
	if _debug {
		fmt.Printf(" (%d)\r\n", l)
	}
	return
}

func (d *Device) sendParamLen16(p uint16) (l int) {
	d.SPI.Transfer(uint8(p >> 8))
	d.SPI.Transfer(uint8(p & 0xFF))
	if _debug {
		fmt.Printf(" %02X %02X", uint8(p>>8), uint8(p&0xFF))
	}
	return 2
}

func (d *Device) sendParamBuf(p []byte, isLastParam bool) (l int) {
	if _debug {
		println("sendParamBuf:")
	}
	l += d.sendParamLen16(uint16(len(p)))
	for _, b := range p {
		if _debug {
			fmt.Printf(" %02X", b)
		}
		d.SPI.Transfer(b)
		l += 1
	}
	if isLastParam {
		if _debug {
			fmt.Printf(" %s", CmdEnd)
		}
		d.SPI.Transfer(byte(CmdEnd))
		l += 1
	}
	if _debug {
		fmt.Printf(" (%d) \r\n", l)
	}
	return
}

func (d *Device) sendParamStr(p string, isLastParam bool) (l int) {
	l = len(p)
	d.SPI.Transfer(uint8(l))
	if l > 0 {
		d.SPI.Tx([]byte(p), nil)
	}
	if isLastParam {
		d.SPI.Transfer(byte(CmdEnd))
		l += 1
	}
	return
}

func (d *Device) sendParam8(p uint8, isLastParam bool) (l int) {
	if _debug {
		println("sendParam8:", p, "lastParam:", isLastParam, "\r")
	}
	l = 2
	d.SPI.Transfer(1)
	d.SPI.Transfer(p)
	if isLastParam {
		d.SPI.Transfer(byte(CmdEnd))
		l += 1
	}
	return
}

func (d *Device) sendParam16(p uint16, isLastParam bool) (l int) {
	l = 3
	d.SPI.Transfer(2)
	d.SPI.Transfer(uint8(p >> 8))
	d.SPI.Transfer(uint8(p & 0xFF))
	if isLastParam {
		d.SPI.Transfer(byte(CmdEnd))
		l += 1
	}
	return
}

func (d *Device) sendParam32(p uint32, isLastParam bool) (l int) {
	l = 5
	d.SPI.Transfer(4)
	d.SPI.Transfer(uint8(p >> 24))
	d.SPI.Transfer(uint8(p >> 16))
	d.SPI.Transfer(uint8(p >> 8))
	d.SPI.Transfer(uint8(p & 0xFF))
	if isLastParam {
		d.SPI.Transfer(byte(CmdEnd))
		l += 1
	}
	return
}

func (d *Device) checkStartCmd() (bool, error) {
	check, err := d.waitSpiChar(byte(CmdStart))
	if err != nil {
		return false, err
	}
	if !check {
		return false, ErrCheckStartCmd
	}
	return true, nil
}

func (d *Device) waitForChipSelect() (err error) {
	if err = d.waitForChipReady(); err == nil {
		err = d.spiChipSelect()
	}
	return
}

func (d *Device) waitForChipReady() error {
	if _debug {
		println("waitForChipReady()\r")
	}
	start := time.Now()
	for time.Since(start) < 10*time.Second {
		if !d.ACK.Get() {
			return nil
		}
		time.Sleep(1 * time.Millisecond)
	}
	return ErrTimeoutChipReady
}

func (d *Device) spiChipSelect() error {
	if _debug {
		println("spiChipSelect()\r")
	}
	d.CS.Low()
	start := time.Now()
	for time.Since(start) < 5*time.Millisecond {
		if d.ACK.Get() {
			return nil
		}
		time.Sleep(100 * time.Microsecond)
	}
	return ErrTimeoutChipSelect
}

func (d *Device) spiChipDeselect() {
	if _debug {
		println("spiChipDeselect\r")
	}
	d.CS.High()
}

func (d *Device) waitSpiChar(wait byte) (bool, error) {
	var timeout = 1000
	var read byte
	for first := true; first || (timeout > 0 && read != wait); timeout-- {
		first = false
		d.readParam(&read)
		if read == byte(CmdErr) {
			return false, ErrCmdErrorReceived
		}
	}
	if _debug && read != wait {
		fmt.Printf("read: %02X, wait: %02X\r\n", read, wait)
	}
	return read == wait, nil
}

func (d *Device) waitRspCmd(cmd CommandType, np uint8) (l uint8, err error) {
	if _debug {
		println("waitRspCmd")
	}
	var check bool
	var data byte
	if check, err = d.checkStartCmd(); !check {
		return
	}
	if check = d.readAndCheckByte(byte(cmd)|FlagReply, &data); !check {
		return
	}
	if check = d.readAndCheckByte(np, &data); check {
		d.readParam(&l)
		for i := uint8(0); i < l; i++ {
			d.readParam(&d.buf[i])
		}
	}
	if !d.readAndCheckByte(byte(CmdEnd), &data) {
		err = ErrIncorrectSentinel
	}
	return
}

func (d *Device) waitRspBuf16(cmd CommandType, buf []byte) (l uint16, err error) {
	if _debug {
		println("waitRspBuf16")
	}
	var check bool
	var data byte
	if check, err = d.checkStartCmd(); !check {
		return
	}
	if check = d.readAndCheckByte(byte(cmd)|FlagReply, &data); !check {
		return
	}
	if check = d.readAndCheckByte(1, &data); check {
		l, _ = d.readParamLen16()
		for i := uint16(0); i < l; i++ {
			d.readParam(&buf[i])
		}
	}
	if !d.readAndCheckByte(byte(CmdEnd), &data) {
		err = ErrIncorrectSentinel
	}
	return
}

func (d *Device) waitRspStr(cmd CommandType, sl []string) (numRead uint8, err error) {
	if _debug {
		println("waitRspStr")
	}
	var check bool
	var data byte
	if check, err = d.checkStartCmd(); !check {
		return
	}
	if check = d.readAndCheckByte(byte(cmd)|FlagReply, &data); !check {
		return
	}
	numRead, _ = d.SPI.Transfer(dummyData)
	if numRead == 0 {
		return 0, ErrNoParamsReturned
	}
	maxNumRead := uint8(len(sl))
	for j, l := uint8(0), uint8(0); j < numRead; j++ {
		d.readParam(&l)
		for i := uint8(0); i < l; i++ {
			d.readParam(&d.buf[i])
		}
		if j < maxNumRead {
			sl[j] = string(d.buf[0:l])
			if _debug {
				fmt.Printf("str %d (%d) - %08X\r\n", j, l, []byte(sl[j]))
			}
		}
	}
	for j := numRead; j < maxNumRead; j++ {
		if _debug {
			println("str", j, "\"\"\r")
		}
		sl[j] = ""
	}
	if !d.readAndCheckByte(byte(CmdEnd), &data) {
		err = ErrIncorrectSentinel
	}
	if numRead > maxNumRead {
		numRead = maxNumRead
	}
	return
}

func (d *Device) readAndCheckByte(check byte, read *byte) bool {
	d.readParam(read)
	return (*read == check)
}

// readParamLen16 reads 2 bytes from the SPI bus (MSB first), returning uint16
func (d *Device) readParamLen16() (v uint16, err error) {
	if b, err := d.SPI.Transfer(0xFF); err == nil {
		v |= uint16(b) << 8
		if b, err = d.SPI.Transfer(0xFF); err == nil {
			v |= uint16(b)
		}
	}
	return
}

func (d *Device) readParam(b *byte) (err error) {
	*b, err = d.SPI.Transfer(0xFF)
	return
}

func (d *Device) addPadding(l int) {
	if _debug {
		println("addPadding", l, "\r")
	}
	for i := (4 - (l % 4)) & 3; i > 0; i-- {
		if _debug {
			println("padding\r")
		}
		d.SPI.Transfer(dummyData)
	}
}

func (d *Device) padTo4(l int) {
	if _debug {
		println("padTo4", l, "\r")
	}

	for l%4 != 0 {
		d.SPI.Transfer(dummyData)
		l++
	}
}
