package esp8266

// Basic AT commands
const (
	// Test that the device is working.
	Test = ""

	// Restart module
	Restart = "+RST"

	// Version show info about the current software version.
	Version = "+GMR"

	// Enter deep-sleep mode
	Sleep = "+GSLP"

	// Configure echo.
	EchoConfig = "E"

	// EchoConfigOn
	EchoConfigOn = EchoConfig + "1"

	// EchoConfigOff
	EchoConfigOff = EchoConfig + "0"

	// Configure UART
	UARTConfig = "+UART"
)

// WiFi commands.
const (
	// WiFi mode (sta/AP/sta+AP)
	WifiMode = "+CWMODE"

	// Connect to AP
	ConnectAP = "+CWJAP"

	// List available AP's
	ListAP = "+CWLAP"

	// Disconnect from AP
	Disconnect = "+CWQAP"

	// Set softAP configuration
	SoftAPConfig = "+CWSAP"

	// List station IP's connected to softAP
	ListConnectedIP = "+CWLIF"

	// Enable/disable DHCP
	DHCPConfig = "+CWDHCP"

	// Set MAC address of station
	SetStationMACAddress = "+CIPSTAMAC"

	// Set MAC address of softAP
	SetAPMACAddress = "+CIPAPMAC"

	// Set IP address of ESP8266 station
	SetStationIP = "+CIPSTA"

	// Set IP address of ESP8266 softAP
	SetSoftAPIP = "+CIPAP"
)

// TCP/IP commands
const (
	// Get connection status
	TCPStatus = "+CIPSTATUS"

	// Establish TCP connection or register UDP port
	TCPConnect = "+CIPSTART"

	// Send Data
	TCPSend = "+CIPSEND"

	// Close TCP/UDP connection
	TCPClose = "+CIPCLOSE"

	// Get local IP address
	GetLocalIP = "+CIFSR"

	// Set multiple connections mode
	TCPMultiple = "+CIPMUX"

	// Configure as server
	ServerConfig = "+CIPSERVER"

	// Set transmission mode
	TransmissionMode = "+CIPMODE"

	// Set timeout when ESP8266 runs as TCP server
	SetServerTimeout = "+CIPSTO"
)

// Execute sends an AT command to the ESP8266.
func (d Device) Execute(cmd string) error {
	_, err := d.Write([]byte("AT" + cmd + "\r\n"))
	return err
}

// Query sends an AT command to the ESP8266 that returns the
// current value for some configuration parameter.
func (d Device) Query(cmd string) (string, error) {
	_, err := d.Write([]byte("AT" + cmd + "?\r\n"))
	return "", err
}

// Set sends an AT command with params to the ESP8266 for a
// configuration value to be set.
func (d Device) Set(cmd, params string) error {
	_, err := d.Write([]byte("AT" + cmd + "=" + params + "\r\n"))
	return err
}

// Version returns the ESP8266 firmware version info.
func (d Device) Version() []byte {
	d.Execute(Version)
	return d.Response()
}

// Echo sets the ESP8266 echo setting.
func (d Device) Echo(set bool) {
	if set {
		d.Execute(EchoConfigOn)
	} else {
		d.Execute(EchoConfigOff)
	}
	// TODO: check for success
	d.Response()
}

// Reset restarts the ESP8266 firmware. Due to how the baud rate changes,
// this messes up communication with the ESP8266 module. So make sure you know
// what you are doing when you call this.
func (d Device) Reset() {
	d.Execute(Restart)
	d.Response()
}
