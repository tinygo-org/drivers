# Netdev

#### Table of Contents

- [Overview](#overview)
- [Porting Applications from Go "net"](#porting-applications-from-go-net)
- [Using Raw Sockets](#using-raw-sockets)
- [Writing a New Driver](#writing-a-new-driver)
 
## Overview

Netdev is TinyGo's network device driver model.  

Let's see where netdev fits in the network stack.  The diagram below shows the traditional full OS stack vs. different possible embedded stacks for TinyGo.

![Netdev models](netdev_models.jpg)

In the traditional full OS stack, the driver that sits above hardware (the "nic") and below TCP/IP is the network driver, the netdev.  The netdev provides a raw packet interface to the OS.

For TinyGo netdev, the netdev includes TCP/IP and provides a socket(2) interface to TinyGo's "net" package.  Applications are written to use the net.Conn interfaces.  "net" translates net.Conn functions (Dial, Listen, Read, Write) into netdev socket(2) calls.  The netdev translates those socket(2) calls into hardware access, ultimately.  Let's consider the three use cases:

#### Firware Offload Model

Here we are fortunate that hardware includes firmware with a TCP/IP implmentation, and the firmware manages the TCP/IP connection state.  The netdev driver translates socket(2) calls to the firmware's TCP/IP calls.  Usually, minimal work is required since the firmware is likely to use lwip, which has an socket(2) API.

The Wifinina (ESP32) and RTL8720dn netdev drivers are examples of the firmware offload model.

#### Full Stack Model

Here the netdev includes the TCP/IP stack, maybe some port of lwip/uip to Go?

#### "Bring-Your-Own-net.Comm" Model

Here the netdev is the entire stack, accessing hardware on the bottom and serving up net.Conn connections above to applications.

## Porting Applications from Go "net"

Ideally, TinyGo's "net" package would just be Go's "net" package and applications using "net" would just work, as-is.  Unfortunately, Go's "net" can't fully be ported to TinyGo, so TinyGo's "net" is a subset of Go's.

To view TinyGo's "net" package exports, use ```go doc ./net```, ```go doc ./net/http```, etc.  For the most part, Go's "net" documentation applies to TinyGo's "net".  There are a few features excluded during the porting process, in particular:

- No IPv6 support
- No HTTP/2 support
- HTTP client request can't be reused
- No TLS support for HTTP servers (no https servers)
- No DualStack support

Applications using Go's "net" package will need a few setup modifications to work with TinyGo's "net" package.

### Step 1: Create the netdev for your target device.

The available netdev are:

- wifinina:  SPI to ESP32 WiFi co-controller running Arduino WiFiNINA firmware
- rtl8720dn: UART to RealTek WiFi rtl8720dn co-controller
- espat:     UART to ESP32/ESP8266 WiFi co-controller running Espressif AT firmware

This example configures and creates a wifinina netdev using New().

```
import "tinygo.org/x/drivers/wifinina"

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	dev := wifinina.New(&cfg)
	...
}
```

The Config structure is netdev-specific; consult the netdev package for Config details.  In this case, the WiFi credentials are passed.

### Step 2: Hook the netdev into the "net" package

Tell the "net" package to use the netdev.  Continuing with the wifinina example:

```
import "tinygo.org/x/drivers/netdev"
import "tinygo.org/x/drivers/wifinina"

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	dev := wifinina.New(&cfg)
	netdev.Use(dev)
	...
}
```

### Step 3: Connect to an IP Network

Before the "net" package is fully functional, connect the netdev to an underlying IP network.  For example, a WiFi netdev would connect to a WiFi access point or become a WiFi access point; either way, once connected, the netdev has a station IP address and is connected on the IP network.

Call dev.NetConnect() to connect the device to an IP network.  Call dev.NetDisconnect() to disconnect.  Continuing example:

```
import "tinygo.org/x/drivers/netdev"
import "tinygo.org/x/drivers/wifinina"

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	dev := wifinina.New(&cfg)
	netdev.Use(dev)

	dev.NetConnect()
        
	// "net" package calls here
	
	dev.NetDisconnect()
}
```

Here is a simple http server listening on port :8080, before and after porting from Go "net/http":

#### Before
```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}
```

#### After
```go
package main

import (
	"fmt"
	"net/http"
    
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/wifinina"
)

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	dev := wifinina.New(&cfg)
	netdev.Use(dev)
	dev.NetConnect()

	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}
```

## Using Raw Sockets

A netdev implements the Socketer interface so an application can make raw socket calls, bypassing the "net" package.

Here is a simple TCP application using raw sockets:

```go
package main

import (
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/wifinina"
)

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	dev := wifinina.New(&cfg)
	netdev.Use(dev)
	dev.NetConnect()

	sock, _ := dev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TCP)

	sockAddr := netdev.NewSockAddr("", netdev.Port(8080), netdev.ParseIP("10.0.0.100")
        dev.Connect(sock, sockAddr)

	dev.Send(sock, []bytes("hello"), 0, 0)

	dev.Close(sock)
}
```

## Writing a New Driver

:bulb: A reference netdev driver is the Wifinina driver (netdev/wifinina).

Netdev drivers implement the net.Netdever interface, which includes the net.Socketer interface.  The Socketer interface is modeled after BSD socket(2).  TinyGo's "net" package translates net.Conn calls into netdev Socketer calls.  For example, DialTCP calls netdev.Socket() and netdev.Connect():

```go
func DialTCP(network string, laddr, raddr *TCPAddr) (*TCPConn, error) {

        fd, _ := netdev.Socket(AF_INET, SOCK_STREAM, IPPROTO_TCP)

        addr := NewSockAddr("", uint16(raddr.Port), raddr.IP)
        
        netdev.Connect(fd, addr)

        return &TCPConn{
                fd:    fd,
                laddr: laddr,
                raddr: raddr,
        }, nil
}
```

### net.Socketer Interface

```go
type Socketer interface {
        Socket(family AddressFamily, sockType SockType, protocol Protocol) (Sockfd, error)
        Bind(sockfd Sockfd, myaddr SockAddr) error
        Connect(sockfd Sockfd, servaddr SockAddr) error
        Listen(sockfd Sockfd, backlog int) error
        Accept(sockfd Sockfd, peer SockAddr) (Sockfd, error)
        Send(sockfd Sockfd, buf []byte, flags SockFlags, timeout time.Duration) (int, error)
        Recv(sockfd Sockfd, buf []byte, flags SockFlags, timeout time.Duration) (int, error)
        Close(sockfd Sockfd) error
        SetSockOpt(sockfd Sockfd, level SockOptLevel, opt SockOpt, value any) error
}
```

Socketer interface is intended to mimic a subset of BSD socket(2).  They've been Go-ified, but should otherwise maintain the semantics of the original socket(2) calls.  Send and Recv add a timeout to put a limit on blocking operations.  Recv in paricular is blocking and will block until data arrives on the socket or EOF.  The timeout is calculated from net.Conn's SetDeadline(), typically.

#### Locking

Multiple goroutines may invoke methods on a net.Conn simultaneously, and the "net" package translates net.Conn calls into Socketer calls.  It follows that multiple goroutines may invoke Socketer calls, so locking is required to keep Socketer calls from stepping on one another.

Don't hold a lock while Time.Sleep()'ing waiting for a hardware operation to finish.  Unlocking while sleeping let's other goroutines to run.  If the sleep period is really small, then you can get away with holding the lock sometimes.

#### Sockfd

The Socketer interface uses a socket fd to represent a socket connection end-point.  Each net.Conn maps 1:1 to a fd.  The number of fds available is a netdev hardware limitation.  Wifinina, for example, can hand out 10 socket fds.

### Packaging

1. Create a new directory in netdev/foo to hold the driver files.

2. Add a initialization file netdev/netdev_foo.go to compile and load the driver based on target build tags.

### Testing

The netdev driver should minimally pass all of the example/net examples.

TODO: automate testing to catch regressions.  
