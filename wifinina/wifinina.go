// Package wifinina implements TCP wireless communication over SPI
// with an attached separate ESP32 board using the Arduino WiFiNINA protocol.
//
// In order to use this driver, the ESP32 must be flashed with specific firmware from Arduino.
// For more information: https://github.com/arduino/nina-fw
//
package wifinina // import "tinygo.org/x/drivers/wifinina"

import (
	"encoding/binary"
	"fmt"
	"time"

	"tinygo.org/x/drivers/net"
)

const _debug = false

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

	StatusNoShield       ConnectionStatus = 255
	StatusIdle           ConnectionStatus = 0
	StatusNoSSIDAvail    ConnectionStatus = 1
	StatusScanCompleted  ConnectionStatus = 2
	StatusConnected      ConnectionStatus = 3
	StatusConnectFailed  ConnectionStatus = 4
	StatusConnectionLost ConnectionStatus = 5
	StatusDisconnected   ConnectionStatus = 6

	EncTypeTKIP EncryptionType = 2
	EncTypeCCMP EncryptionType = 4
	EncTypeWEP  EncryptionType = 5
	EncTypeNone EncryptionType = 7
	EncTypeAuto EncryptionType = 8

	TCPStateClosed      = 0
	TCPStateListen      = 1
	TCPStateSynSent     = 2
	TCPStateSynRcvd     = 3
	TCPStateEstablished = 4
	TCPStateFinWait1    = 5
	TCPStateFinWait2    = 6
	TCPStateCloseWait   = 7
	TCPStateClosing     = 8
	TCPStateLastACK     = 9
	TCPStateTimeWait    = 10
	/*
		// Default state value for Wifi state field
		#define NA_STATE -1
	*/

	FlagCmd   = 0
	FlagReply = 1 << 7
	FlagData  = 0x40

	NinaCmdPos      = 1
	NinaParamLenPos = 2

	CmdStart = 0xE0
	CmdEnd   = 0xEE
	CmdErr   = 0xEF

	//dummyData = 0xFF

	CmdSetNet          = 0x10
	CmdSetPassphrase   = 0x11
	CmdSetKey          = 0x12
	CmdSetIPConfig     = 0x14
	CmdSetDNSConfig    = 0x15
	CmdSetHostname     = 0x16
	CmdSetPowerMode    = 0x17
	CmdSetAPNet        = 0x18
	CmdSetAPPassphrase = 0x19
	CmdSetDebug        = 0x1A
	CmdGetTemperature  = 0x1B
	CmdGetReasonCode   = 0x1F
	//	TEST_CMD	        = 0x13

	CmdGetConnStatus     = 0x20
	CmdGetIPAddr         = 0x21
	CmdGetMACAddr        = 0x22
	CmdGetCurrSSID       = 0x23
	CmdGetCurrBSSID      = 0x24
	CmdGetCurrRSSI       = 0x25
	CmdGetCurrEncrType   = 0x26
	CmdScanNetworks      = 0x27
	CmdStartServerTCP    = 0x28
	CmdGetStateTCP       = 0x29
	CmdDataSentTCP       = 0x2A
	CmdAvailDataTCP      = 0x2B
	CmdGetDataTCP        = 0x2C
	CmdStartClientTCP    = 0x2D
	CmdStopClientTCP     = 0x2E
	CmdGetClientStateTCP = 0x2F
	CmdDisconnect        = 0x30
	CmdGetIdxRSSI        = 0x32
	CmdGetIdxEncrType    = 0x33
	CmdReqHostByName     = 0x34
	CmdGetHostByName     = 0x35
	CmdStartScanNetworks = 0x36
	CmdGetFwVersion      = 0x37
	CmdSendDataUDP       = 0x39
	CmdGetRemoteData     = 0x3A
	CmdGetTime           = 0x3B
	CmdGetIdxBSSID       = 0x3C
	CmdGetIdxChannel     = 0x3D
	CmdPing              = 0x3E
	CmdGetSocket         = 0x3F
	//	GET_IDX_SSID_CMD	= 0x31,
	//	GET_TEST_CMD		= 0x38

	// All command with DATA_FLAG 0x40 send a 16bit Len
	CmdSendDataTCP   = 0x44
	CmdGetDatabufTCP = 0x45
	CmdInsertDataBuf = 0x46

	// regular format commands
	CmdSetPinMode      = 0x50
	CmdSetDigitalWrite = 0x51
	CmdSetAnalogWrite  = 0x52

	ErrTimeoutSlaveReady  Error = 0x01
	ErrTimeoutSlaveSelect Error = 0x02
	ErrCheckStartCmd      Error = 0x03
	ErrWaitRsp            Error = 0x04
	ErrUnexpectedLength   Error = 0xE0
	ErrNoParamsReturned   Error = 0xE1
	ErrIncorrectSentinel  Error = 0xE2
	ErrCmdErrorReceived   Error = 0xEF
	ErrNotImplemented     Error = 0xF0
	ErrUnknownHost        Error = 0xF1
	ErrSocketAlreadySet   Error = 0xF2
	ErrConnectionTimeout  Error = 0xF3
	ErrNoData             Error = 0xF4
	ErrDataNotWritten     Error = 0xF5
	ErrCheckDataError     Error = 0xF6
	ErrBufferTooSmall     Error = 0xF7
	ErrNoSocketAvail      Error = 0xFF

	NoSocketAvail uint8 = 0xFF
)

const (
	ProtoModeTCP = iota
	ProtoModeUDP
	ProtoModeTLS
	ProtoModeMul
)

type ConnectionStatus uint8

func (c ConnectionStatus) String() string {
	switch c {
	case StatusIdle:
		return "Idle"
	case StatusNoSSIDAvail:
		return "No SSID Available"
	case StatusScanCompleted:
		return "Scan Completed"
	case StatusConnected:
		return "Connected"
	case StatusConnectFailed:
		return "Connect Failed"
	case StatusConnectionLost:
		return "Connection Lost"
	case StatusDisconnected:
		return "Disconnected"
	case StatusNoShield:
		return "No Shield"
	default:
		return "Unknown"
	}
}

type EncryptionType uint8

func (e EncryptionType) String() string {
	switch e {
	case EncTypeTKIP:
		return "TKIP"
	case EncTypeCCMP:
		return "WPA2"
	case EncTypeWEP:
		return "WEP"
	case EncTypeNone:
		return "None"
	case EncTypeAuto:
		return "Auto"
	default:
		return "Unknown"
	}
}

type IPAddress string // TODO: does WiFiNINA support ipv6???

func (addr IPAddress) String() string {
	if len(addr) < 4 {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])
}

func ParseIPv4(s string) (IPAddress, error) {
	var v0, v1, v2, v3 uint8
	if _, err := fmt.Sscanf(s, "%d.%d.%d.%d", &v0, &v1, &v2, &v3); err != nil {
		return "", err
	}
	return IPAddress([]byte{v0, v1, v2, v3}), nil
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
	return fmt.Sprintf("%016X", uint64(addr))
}

type Error uint8

func (err Error) Error() string {
	return fmt.Sprintf("wifinina error: 0x%02X", uint8(err))
}

type Device struct {
	Transport Transport
	buf       [64]byte
	ssids     [10]string
	cmdbuf    Buffer
}

type Transport interface {
	Configure()

	SetGPIO0(level bool)
	SetReset(level bool)
	SetCS(level bool)

	GetACK(level bool, timeout time.Duration) bool

	Transfer(b byte) (byte, error)
	Tx(w []byte, r []byte) error
}

func (d *Device) Configure() {
	net.UseDriver(d.NewDriver())
	d.Transport.Configure()

	d.Transport.SetGPIO0(true)
	d.Transport.SetCS(true)

	d.Transport.SetReset(false)
	time.Sleep(1 * time.Millisecond)

	d.Transport.SetReset(true)
	time.Sleep(1 * time.Millisecond)

	d.Transport.SetGPIO0(false)
}

// ----------- client methods (should this be a separate struct?) ------------

func (d *Device) StartClient(addr uint32, port uint16, sock uint8, mode uint8) error {
	if _debug {
		println("[StartClient] called StartClient()\r")
		fmt.Printf("[StartClient] addr: % 02X, port: %d, sock: %d\r\n", addr, port, sock)
	}
	d.cmdbuf.StartCmd(CmdStartClientTCP)
	d.cmdbuf.AddUint32(addr)
	d.cmdbuf.AddUint16(port)
	d.cmdbuf.AddByte(sock)
	d.cmdbuf.AddByte(mode)
	if err := d.txcmd(); err != nil {
		return err
	}
	r, err := d.waitRspCmd1(CmdStartClientTCP)
	_ = r
	//println("r:", r)
	return err
}

func (d *Device) GetSocket() (uint8, error) {
	return d.getUint8(d.req0(CmdGetSocket))
}

func (d *Device) GetClientState(sock uint8) (uint8, error) {
	return d.getUint8(d.reqUint8(CmdGetClientStateTCP, sock))
}

func (d *Device) SendData(buf []byte, sock uint8) (uint16, error) {
	d.cmdbuf.StartCmd(CmdSendDataTCP)
	d.cmdbuf.AddByte(sock)
	d.cmdbuf.AddData(buf)
	if err := d.txcmd(); err != nil {
		return 0, err
	}
	return d.getUint16(d.waitRspCmd1(CmdSendDataTCP))
}

// InsertDataBuf adds data to the buffer used for sending UDP data
func (d *Device) InsertDataBuf(buf []byte, sock uint8) (bool, error) {
	d.cmdbuf.StartCmd(CmdInsertDataBuf)
	d.cmdbuf.AddByte(sock)
	d.cmdbuf.AddData(buf)
	if err := d.txcmd(); err != nil {
		return false, err
	}
	n, err := d.getUint8(d.waitRspCmd1(CmdInsertDataBuf))
	return n == 1, err
}

// SendUDPData sends the data previously added to the UDP buffer
func (d *Device) SendUDPData(sock uint8) (bool, error) {
	d.cmdbuf.StartCmd(CmdSendDataUDP)
	d.cmdbuf.AddByte(sock)
	if err := d.txcmd(); err != nil {
		return false, err
	}
	n, err := d.getUint8(d.waitRspCmd1(CmdSendDataUDP))
	//println("n:", n)
	return n == 1, err
}

func (d *Device) CheckDataSent(sock uint8) (bool, error) {
	var lastErr error
	for timeout := 0; timeout < 10; timeout++ {
		sent, err := d.getUint8(d.reqUint8(CmdDataSentTCP, sock))
		if err != nil {
			lastErr = err
		}
		if sent > 0 {
			return true, nil
		}
		wait(100 * time.Microsecond)
	}
	return false, lastErr
}

func (d *Device) GetDataBuf(sock uint8, buf []byte) (int, error) {
	p := uint16(len(buf))
	d.cmdbuf.StartCmd(CmdGetDatabufTCP)
	d.cmdbuf.AddByte(sock)
	//d.cmdbuf.AddUint16(p)
	d.cmdbuf.AddData([]byte{uint8(p & 0x00FF), uint8((p) >> 8)}) // TODO: is this the right byte order?
	if err := d.txcmd(); err != nil {
		return 0, err
	}
	n, err := d.waitRspBuf16(CmdGetDatabufTCP, buf)
	return int(n), err
}

func (d *Device) StopClient(sock uint8) error {
	if _debug {
		println("[StopClient] called StopClient()\r")
	}
	_, err := d.getUint8(d.reqUint8(CmdStopClientTCP, sock))
	return err
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
	return d.getInt32(d.req1(CmdGetCurrBSSID))
}

func (d *Device) GetCurrentSSID() (string, error) {
	return d.getString(d.req1(CmdGetCurrSSID))
}

func (d *Device) GetMACAddress() (MACAddress, error) {
	return d.getMACAddress(d.req1(CmdGetMACAddr))
}

func (d *Device) GetIP() (ip, subnet, gateway IPAddress, err error) {
	sl := make([]string, 3)
	if l, err := d.reqRspStr1(CmdGetIPAddr, 0xFF, sl); err != nil {
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

func (d *Device) GetTime() (string, error) {
	return d.getString(d.req0(CmdGetTime))
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
	return ErrNotImplemented
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
	return ErrNotImplemented
}

func (d *Device) SetHostname(hostname string) error {
	return ErrNotImplemented
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
	return MACAddress(binary.LittleEndian.Uint64(d.buf[0:8]) >> 16), err
}

// --------- end of methods for getting buffered response data --------------

// req0 sends a command to the device with no request parameters
func (d *Device) req0(cmd uint8) (l uint8, err error) {
	if err := d.sendCmdNoParams(cmd); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// req1 sends a command to the device with a single dummy parameters of 0xFF
func (d *Device) req1(cmd uint8) (l uint8, err error) {
	return d.reqUint8(cmd, 0xFF)
}

// reqUint8 sends a command to the device with a single uint8 parameter
func (d *Device) reqUint8(cmd uint8, data uint8) (l uint8, err error) {
	if err := d.sendCmdWithByteParam(cmd, data); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// req2Uint8 sends a command to the device with two uint8 parameters
func (d *Device) req2Uint8(cmd, p1, p2 uint8) (l uint8, err error) {
	if err := d.sendCmdWith2ByteParams(cmd, p1, p2); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// reqStr sends a command to the device with a single string parameter
func (d *Device) reqStr(cmd uint8, p1 string) (uint8, error) {
	if err := d.sendCmdWithStringParam(cmd, p1); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// reqStr sends a command to the device with 2 string parameters
func (d *Device) reqStr2(cmd uint8, p1 string, p2 string) (uint8, error) {
	if err := d.sendCmdWith2StringParams(cmd, p1, p2); err != nil {
		return 0, err
	}
	return d.waitRspCmd1(cmd)
}

// reqStrRsp0 sends a command passing a string slice for the response
func (d *Device) reqRspStr0(cmd uint8, sl []string) (l uint8, err error) {
	if err := d.sendCmdNoParams(cmd); err != nil {
		return 0, err
	}
	return d.waitRspStr(cmd, sl)
}

// reqStrRsp1 sends a command with a uint8 param and a string slice for the response
func (d *Device) reqRspStr1(cmd uint8, data uint8, sl []string) (uint8, error) {
	if err := d.sendCmdWithByteParam(cmd, data); err != nil {
		return 0, err
	}
	return d.waitRspStr(cmd, sl)
}

// --------- end of methods for sending command "requests" --------------

func (d *Device) sendCmdNoParams(cmd uint8) error {
	d.cmdbuf.StartCmd(cmd)
	return d.txcmd()
}

func (d *Device) sendCmdWithByteParam(cmd uint8, data uint8) error {
	d.cmdbuf.StartCmd(cmd)
	d.cmdbuf.AddByte(data)
	return d.txcmd()
}

func (d *Device) sendCmdWith2ByteParams(cmd, data1, data2 uint8) error {
	d.cmdbuf.StartCmd(cmd)
	d.cmdbuf.AddByte(data1)
	d.cmdbuf.AddByte(data2)
	return d.txcmd()
}

func (d *Device) sendCmdWithStringParam(cmd uint8, p1 string) (err error) {
	d.cmdbuf.StartCmd(cmd)
	d.cmdbuf.AddString(p1)
	return d.txcmd()
}

func (d *Device) sendCmdWith2StringParams(cmd uint8, p1 string, p2 string) (err error) {
	d.cmdbuf.StartCmd(cmd)
	d.cmdbuf.AddString(p1)
	d.cmdbuf.AddString(p2)
	return d.txcmd()
}

// --------- end of methods for sending commands --------------

func (d *Device) waitRspCmd1(cmd uint8) (l uint8, err error) {
	return d.waitRspCmd(cmd, 1)
}

func (d *Device) txcmd() error {
	d.cmdbuf.EndCmd()
	defer d.Transport.SetCS(true)
	//if len(debug) > 0 && debug[0] {
	//	PrintBuffer(&d.cmdbuf, os.Stdout)
	//}
	if err := d._waitForSlaveSelect(); err != nil {
		return err
	}
	return d.Transport.Tx(d.cmdbuf.Bytes(), nil)
}

func (d *Device) _waitForSlaveSelect() (err error) {
	if _debug {
		println("wifinina: wait for slave ready\r")
	}
	if ok := d.Transport.GetACK(false, 10*time.Second); !ok {
		return ErrTimeoutSlaveReady
	}
	if _debug {
		println("spiSlaveSelect()\r")
	}
	d.Transport.SetCS(false)
	if ok := d.Transport.GetACK(true, 5*time.Millisecond); !ok {
		return ErrTimeoutSlaveSelect
	}
	return
}

func (d *Device) waitRspCmd(cmd uint8, np uint8) (l uint8, err error) {
	if _debug {
		println("waitRspCmd")
	}
	defer d.Transport.SetCS(true)
	if err = d._waitForSlaveSelect(); err != nil {
		return
	}
	var check bool
	var data byte
	if check, err = d.checkStartCmd(); !check {
		return
	}
	if check = d.readAndCheckByte(cmd|FlagReply, &data); !check {
		return
	}
	if check = d.readAndCheckByte(np, &data); check {
		d.readParam(&l)
		for i := uint8(0); i < l; i++ {
			d.readParam(&d.buf[i])
		}
	}
	if !d.readAndCheckByte(CmdEnd, &data) {
		err = ErrIncorrectSentinel
	}
	return
}

func (d *Device) waitRspBuf16(cmd uint8, buf []byte) (l uint16, err error) {
	if _debug {
		println("waitRspBuf16")
	}
	defer d.Transport.SetCS(true)
	if err = d._waitForSlaveSelect(); err != nil {
		return
	}
	var check bool
	var data byte
	if check, err = d.checkStartCmd(); !check {
		return
	}
	if check = d.readAndCheckByte(cmd|FlagReply, &data); !check {
		return
	}
	if check = d.readAndCheckByte(1, &data); check {
		l, _ = d.readParamLen16()
		for i := uint16(0); i < l; i++ {
			d.readParam(&buf[i])
		}
	}
	if !d.readAndCheckByte(CmdEnd, &data) {
		err = ErrIncorrectSentinel
	}
	return
}

func (d *Device) waitRspStr(cmd uint8, sl []string) (numRead uint8, err error) {
	if _debug {
		println("waitRspStr")
	}
	defer d.Transport.SetCS(true)
	if err = d._waitForSlaveSelect(); err != nil {
		return
	}
	var check bool
	var data byte
	if check, err = d.checkStartCmd(); !check {
		return
	}
	if check = d.readAndCheckByte(cmd|FlagReply, &data); !check {
		return
	}
	numRead, _ = d.Transport.Transfer(0xFF)
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
	if !d.readAndCheckByte(CmdEnd, &data) {
		err = ErrIncorrectSentinel
	}
	if numRead > maxNumRead {
		numRead = maxNumRead
	}
	return
}

func (d *Device) checkStartCmd() (bool, error) {
	var read byte
	for now := time.Now(); time.Since(now) < 5*time.Millisecond; {
		d.readParam(&read)
		if read == CmdErr {
			return false, ErrCmdErrorReceived
		}
		if read == CmdStart {
			return true, nil
		}
	}
	return false, ErrCheckStartCmd
}

func (d *Device) readAndCheckByte(check byte, read *byte) bool {
	d.readParam(read)
	return (*read == check)
}

// readParamLen16 reads 2 bytes from the SPI bus (MSB first), returning uint16
func (d *Device) readParamLen16() (v uint16, err error) {
	if b, err := d.Transport.Transfer(0xFF); err == nil {
		v |= uint16(b << 8)
		if b, err = d.Transport.Transfer(0xFF); err == nil {
			v |= uint16(b)
		}
	}
	return
}

func (d *Device) readParam(b *byte) (err error) {
	*b, err = d.Transport.Transfer(0xFF)
	return
}

func wait(d time.Duration) {
	for now := time.Now(); time.Since(now) < d; {
	}
}
