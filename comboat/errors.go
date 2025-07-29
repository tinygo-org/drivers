package comboat

import (
	"bytes"
	"strconv"
)

var errStrings = map[int]string{

	// System framework related error codes

	0:  "success",
	1:  "The command is not supported (the combo framework contains the command but the current platform has not transplanted or adapted to support it)",
	2:  "The command parameters contain unsupported operations (the current platform only supports some operations for this command)",
	3:  "The instruction format is incorrect (this refers to the wrong number of parameters, for example, two parameters are required, but only one parameter is entered)",
	4:  "Parameter error (the content of the parameter is wrong, for example, a number between 0 and 9 is required, but 10 or xyz is passed in, which is a parameter error)",
	5:  "Parameter length error (command length exceeds the maximum supported length)",
	31: "The current command has not ended and needs to report the status asynchronously. This value is used by the state machine to determine the use of the command and no message is returned.",
	32: "Unknown error (or unhandled error type)",

	// Common error codes

	33: "malloc error",
	34: "Failed to read buf",
	35: "Failed to write buf",
	36: "Configuration error (configuration error loaded from memory, for example, we set port -1 for OTA upgrade, and check port error when executing AT+OTA, then configuration error will be reported)",
	37: "Failed to create task",
	38: "Flash read and write failure",
	39: "Serial port configuration error, unsupported baud rate",
	40: "Serial port configuration error, unsupported data bits",
	41: "Serial port configuration error, unsupported stop bit",
	42: "Serial port configuration error, unsupported parity bit",
	43: "Serial port configuration error, unsupported flow control",
	44: "Serial port configuration failed",
	45: "Wrong username/password",
	46: "Low power mode error or unsupported low power mode",
	47: "Uninitialized configuration data error (including io mapping data)",
	63: "General error code (without other information)",

	// Wi-Fi related error codes

	64: "Wi-Fi not initialized or initialization failed",
	65: "Wi-Fi mode error (unable to connect to Wi-Fi in single AP mode)",
	66: "Wi-Fi connection failed",
	67: "Wi-Fi connection successful, error in obtaining IP (DHCP)",
	68: "Failed to obtain encryption method",
	69: "The specified AP was not found.",
	70: "Wi-Fi scan start failed",
	71: "Wi-Fi scan timeout",
	72: "Failed to enable AP hotspot",
	73: "Failed to obtain the Wi-Fi information of the router or the AP information that you enabled yourself",
	74: "The network card (STA/AP) is not running",
	75: "Wi-Fi country code error (unsupported Wi-Fi country code)",
	76: "The current network configuration mode is wrong.",
	95: "Wi-Fi connection unknown error",

	// Socket related error codes

	96:  "Failed to create socket",
	97:  "Socket connection failed",
	98:  "DNS Failure",
	99:  "The socket status is wrong (for example, TCP is not connected yet)",
	100: "Socket type error",
	101: "Socket send failed",
	102: "Socket receive failed",
	103: "Socket monitoring thread creation failed",
	104: "Socket bind error",
	105: "The current connection cannot be transparently linked (wrong socket type or number)",
	106: "PING test failed (all packets lost)",
	107: "Wi-Fi country code error (unsupported Wi-Fi country code)",
	108: "SSL Config Error",
	109: "SSL verification error (usually caused by unsupported SSL encryption type or certificate error)",
	127: "Unknown socket error",
}

func getErrStr(errLine []byte) (errStr string) {
	errStr = "Can't parse ERROR response"
	tokens := bytes.Split(errLine, []byte(":"))
	if len(tokens) > 1 {
		errCode, err := strconv.Atoi(string(tokens[1]))
		if err == nil {
			errStr = errStrings[errCode]
		}
	}
	return
}
