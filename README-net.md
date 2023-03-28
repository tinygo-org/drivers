#### Table of Contents

- ["net" Package](#net-package)
- [Using "net" Package](#using-net-package)
- [Using "net/http" Package](#using-nethttp-package)
- [Using "crypto/tls" Package](#using-cryptotls-package)
- [Using Sockets](#using-sockets)
- [Netdev and Netlink](#netdev-and-netlink)
- [Writing a New Netdev Driver](#writing-a-new-netdev-driver)

## "net" Package

TinyGo's "net" package is ported from Go.  The port offers a subset of Go's
"net" package.  The subset maintains Go 1 compatiblity guarantee.  A Go
application that uses "net" will most-likey just work on TinyGo if the usage is
within the subset offered.  (There may be external constraints such as limited
SRAM on embedded environment that may limit full functionality).

Continue below for details on using "net" and "net/http" packages.

See src/net/READMD.md in the TinyGo repo for more details on maintaining
TinyGo's "net" package.

## Using "net" Package

Ideally, TinyGo's "net" package would be Go's "net" package and applications
using "net" would just work, as-is.  TinyGo's net package is a partial port
from Go's net package, replacing OS socket syscalls with netdev socket calls.

Netdev is TinyGo's network device driver model; read more about
[Netdev](#netdev-and-netlink).

There are a few features excluded during the porting process, in particular:

- No IPv6 support
- No DualStack support

Run ```go doc -all ./src/net``` in TinyGo repo to see full listing.

Applications using Go's net package will need a few setup steps to work with
TinyGo's net package.

### Step 1: Create the netdev for your target device.

The available netdev are:

- [wifinina]: ESP32 WiFi co-controller running Arduino WiFiNINA firmware

	targets: pyportal arduino_nano33 nano_rp2040 metro_m4_airlift
		 arduino_mkrwifi1010 matrixportal_m4

- [rtl8720dn]: RealTek WiFi rtl8720dn co-controller

	targets: wioterminal

- [espat]: ESP32/ESP8266 WiFi co-controller running Espressif AT firmware

	targets: TBD

This example configures and creates a wifinina netdev using New().

```go
import "tinygo.org/x/drivers/wifinina"

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	netdev := wifinina.New(&cfg)
	...
}
```

New() registers the netdev with the "net" package using net.useNetdev().

The Config structure is netdev-specific; consult the specific netdev package
for Config details.  In this case, the WiFi credentials are passed, but other
settings are typically passed such as device configuration.

### Step 2: Connect to an IP Network

Before the net package is fully functional, connect the netdev to an underlying
IP network.  For example, a WiFi netdev would connect to a WiFi access point or
become a WiFi access point; either way, once connected, the netdev has a
station IP address and is connected on the IP network.  Similarly, a LTE netdev
would connect to a LTE provider, giving the device an IP address on the LTE
network.

Using the Netlinker interface, Call netdev.NetConnect() to connect the device
to an IP network.  Call netdev.NetDisconnect() to disconnect.  Continuing example:

```go
import (
	"tinygo.org/x/drivers/wifinina"
)

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	netdev := wifinina.New(&cfg)

	netdev.NetConnect()

	// "net" package calls here

	netdev.NetDisconnect()
}
```

Optionally, get notified of IP network connects and disconnects:

```go
	netdev.Notify(func(e drivers.NetlinkEvent) {
		switch e {
		case drivers.NetlinkEventNetUp:
			println("Network UP")
		case drivers.NetlinkEventNetDown:
			println("Network DOWN")
	})
```

Here is a simple example of an http server listening on port :8080, before and
after:

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

	"tinygo.org/x/drivers/wifinina"
)

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	netdev := wifinina.New(&cfg)
	netdev.NetConnect()

	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}
```

## Using "net/http" Package

TinyGo's net/http package is a partial port of Go's net/http package, providing
a subset of the full net/http package.  There are a few features excluded
during the porting process, in particular:

- No HTTP/2 support
- No TLS support for HTTP servers (no https servers)
- HTTP client request can't be reused

HTTP client methods (http.Get, http.Head, http.Post, and http.PostForm) are
functional.  Dial clients support both HTTP and HTTPS URLs.

HTTP server methods and objects are mostly ported, but for HTTP only; HTTPS
servers are not supported.

HTTP request and response handling code is mostly ported, so most the intricacy
of parsing and writing headers is handled as in the full net/http package.

Run ```go doc -all ./src/net/http``` in TinyGo repo to see full listing.

## Using "crypto/tls" Package

TinyGo's TLS support (crypto/tls) relies on hardware offload of the TLS
protocol.  This is different from Go's crypto/tls package which handles the TLS
protocol in software.

TinyGo's TLS support is only available for client applications.  You can
http.Get() to an http:// or https:// address, but you cannot
http.ListenAndServeTLS() an https server.

The offloading hardware has pre-defined TLS certificates built-in.

## Using Sockets

A netdev implements a BSD socket-like interface so an application can make direct
socket calls, bypassing the net package.

Here is a simple TCP application using direct sockets:

```go
package main

import (
	"net"  // only need to parse IP address

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/wifinina"
)

func main() {
	cfg := wifinina.Config{Ssid: "foo", Passphrase: "bar"}
	netdev := wifinina.New(&cfg)

	// ignoring error handling

	netdev.NetConnect()

	sock, _ := netdev.Socket(drivers.AF_INET, drivers.SOCK_STREAM, drivers.IPPROTO_TCP)

        netdev.Connect(sock, "", net.ParseIP("10.0.0.100"), 8080)
	netdev.Send(sock, []bytes("hello"), 0, 0)

	netdev.Close(sock)
}
```

## Netdev and Netlink

Netdev is TinyGo's network device driver model.  Network drivers implement the
netdever interface, providing a common network I/O interface to TinyGo's "net"
package.  The interface is modeled after the BSD socket interface.  net.Conn
implementations (TCPConn, UDPConn, and TLSConn) use the netdev interface for
device I/O access.  For example, net.DialTCP, which returns a net.TCPConn,
calls netdev.Socket() and netdev.Connect():

```go
func DialTCP(network string, laddr, raddr *TCPAddr) (*TCPConn, error) {

        fd, _ := netdev.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)

        netdev.Connect(fd, "", raddr.IP, raddr.Port)

        return &TCPConn{
                fd:    fd,
                laddr: laddr,
                raddr: raddr,
        }, nil
}
```

Network drivers also (optionally) implement the Netlinker interface.  This
interface is not used by TinyGo's "net" package, but rather provides the TinyGo
application direct access to the network device for common settings and control
that fall outside of netdev's socket interface.

## Writing a New Netdev Driver

A new netdev driver will implement the netdever and optionally the Netlinker
interfaces.  See the wifinina or rtl8720dn drivers for examples.

#### Locking

Multiple goroutines may invoke methods on a net.Conn simultaneously, and since
the net package translates net.Conn calls into netdev socket calls, it follows
that multiple goroutines may invoke socket calls, so locking is required to
keep socket calls from stepping on one another.

Don't hold a lock while Time.Sleep()ing waiting for a hardware operation to
finish.  Unlocking while sleeping let's other goroutines make progress.  If the
sleep period is really small, then you can get away with holding the lock.

#### Sockfd

The netdev socket interface uses a socket fd (int) to represent a socket
connection (end-point).  Each net.Conn maps 1:1 to a fd.  The number of fds
available is a hardware limitation.  Wifinina, for example, can hand out 10
fds.

### Testing

The netdev driver should minimally run all of the example/net examples.

TODO: automate testing to catch regressions.
