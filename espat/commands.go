package espat

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

	// Connect to an access point.
	ConnectAP = "+CWJAP"

	// List available AP's
	ListAP = "+CWLAP"

	// Disconnect from the current AP
	Disconnect = "+CWQAP"

	// Set softAP configuration. This also activates the ESP8266/ESP32 to act as an access point.
	// The settings will not be saved in flash memory, so they will be forgotten on next reset.
	SoftAPConfigCurrent = "+CWSAP_CUR"

	// Set softAP configuration as saved in flash. This also activates the ESP8266/ESP32 to act as an
	// access point. The settings will be saved in flash memory, so they will be used on next reset.
	SoftAPConfigFlash = "+CWSAP_DEF"

	// List station IP's connected to softAP
	ListConnectedIP = "+CWLIF"

	// Enable/disable DHCP
	DHCPConfig = "+CWDHCP"

	// Set MAC address of station
	SetStationMACAddress = "+CIPSTAMAC"

	// Set MAC address of softAP
	SetAPMACAddress = "+CIPAPMAC"

	// Set IP address of ESP8266/ESP32 station
	SetStationIP = "+CIPSTA"

	// Set IP address of ESP8266/ESP32 when acting as access point.
	// The IP address will not be saved in flash memory, so it will be forgotten on next reset.
	SetSoftAPIPCurrent = "+CIPAP_CUR"

	// Set IP address of ESP8266/ESP32 when acting as access point.
	// The IP address will be saved in flash memory, so they will be used on next reset.
	SetSoftAPIPFlash = "+CIPAP_DEF"
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

	// Set timeout when ESP8266/ESP32 runs as TCP server
	SetServerTimeout = "+CIPSTO"
)
