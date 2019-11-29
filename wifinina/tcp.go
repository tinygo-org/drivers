package wifinina

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bgould/tinygo-model-m/timer"
)

func (d *Device) NewDriver() *Driver {
	return &Driver{d, NoSocketAvail}
}

type Driver struct {
	dev  *Device
	sock uint8
}

func (drv *Driver) GetDNS(domain string) (string, error) {
	ipAddr, err := drv.dev.GetHostByName(domain)
	return ipAddr.String(), err
}

func (drv *Driver) ConnectTCPSocket(addr, portStr string) error {

	//fmt.Println("[ConnectTCPSocket] called ConnectTCPSocket\r")

	// convert port to uint16
	p64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return fmt.Errorf("could not convert port to uint16: %s", err.Error())
	}
	port := uint16(p64)

	// look up the hostname if necessary; if an IP address was specified, the
	// same will be returned.  Otherwise, an IPv4 for the hostname is returned.
	//fmt.Println("[ConnectTCPSocket] Getting host name\r")
	ipAddr, err := drv.dev.GetHostByName(addr)
	if err != nil {
		return err
	}
	//fmt.Printf("[ConnectTCPSocket] Attempting to connect to %s\r\n", ipAddr)
	ip := ipAddr.AsUint32()

	// check to see if socket is already set; if so, stop it
	if drv.sock != NoSocketAvail {
		if err := drv.stop(); err != nil {
			return err
		}
	}

	// get a socket from the device
	if drv.sock, err = drv.dev.GetSocket(); err != nil {
		return err
	}
	//fmt.Printf("[ConnectTCPSocket] TCP socket set to: %d", drv.sock)

	// attempt to start the client
	//fmt.Println("[ConnectTCPSocket] starting client to IP", ipAddr, port, "\r")
	if err := drv.dev.StartClient(ip, port, drv.sock, ProtoModeTCP); err != nil {
		return err
	}

	// FIXME: this 4 second timeout is simply mimicking the Arduino driver
	for t := timer.New(4 * time.Second); !t.Expired(); {
		connected, err := drv.connected()
		if err != nil {
			//fmt.Println("[ConnectTCPSocket] Returning error", err.Error(), "\r")
			return err
		}
		if connected {
			//fmt.Println("[ConnectTCPSocket] Connected!\r")
			return nil
		}
		timer.Wait(1 * time.Millisecond)
	}

	//fmt.Println("[ConnectTCPSocket] Connection Timeout\r")
	return ErrConnectionTimeout
}

func (drv *Driver) ConnectUDPSocket(addr, sport, lport string) error {
	return ErrNotImplemented
}

func (drv *Driver) DisconnectSocket() error {
	return drv.stop()
}

func (drv *Driver) StartSocketSend(size int) error {
	// not needed for WiFiNINA???
	return nil
	// return ErrNotImplemented
}

func (drv *Driver) Write(b []byte) (n int, err error) {
	if drv.sock == NoSocketAvail {
		return 0, ErrNoSocketAvail
	}
	if len(b) == 0 {
		return 0, ErrNoData
	}
	written, err := drv.dev.SendData(b, drv.sock)
	if err != nil {
		return 0, err
	}
	if written == 0 {
		return 0, ErrDataNotWritten
	}
	if sent, _ := drv.dev.CheckDataSent(drv.sock); !sent {
		return 0, ErrCheckDataError
	}
	return len(b), nil
}

func (drv *Driver) ReadSocket(b []byte) (n int, err error) {
	return 0, ErrNotImplemented
}

func (drv *Driver) available() int {
	if drv.sock != NoSocketAvail {

	}
	return 0
}

func (drv *Driver) connected() (bool, error) {
	if drv.sock == NoSocketAvail {
		return false, nil
	}
	if drv.available() > 0 {
		return true, nil
	}
	s, err := drv.status()
	if err != nil {
		return false, err
	}
	isConnected := !(s == TCPStateListen || s == TCPStateClosed ||
		s == TCPStateFinWait1 || s == TCPStateFinWait2 || s == TCPStateTimeWait ||
		s == TCPStateSynSent || s == TCPStateSynRcvd || s == TCPStateCloseWait)
	if !isConnected {
		// close socket buffer?
		// WiFiSocketBuffer.close(_sock);
		// _sock = 255;
	}
	return isConnected, nil
}

func (drv *Driver) status() (uint8, error) {
	if drv.sock == NoSocketAvail {
		return TCPStateClosed, nil
	}
	return drv.dev.GetClientState(drv.sock)
}

func (drv *Driver) stop() error {
	//println("[stop] entering stop()\r")
	if drv.sock == NoSocketAvail {
		return nil
	}
	drv.dev.StopClient(drv.sock)
	for t := timer.New(5 * time.Second); !t.Expired(); {
		st, _ := drv.status()
		if st == TCPStateClosed {
			//println("[stop] TCPStateClosed\r")
			break
		}
		//time.Sleep(1 * time.Millisecond)
	}
	drv.sock = NoSocketAvail
	return nil
}
