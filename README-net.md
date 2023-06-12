### Table of Contents

- ["net" Package](#net-package)
- [Using "net" Package](#using-net-package)
- [Using "net/http" Package](#using-nethttp-package)
- [Using "crypto/tls" Package](#using-cryptotls-package)
- [Using Sockets](#using-sockets)

## "net" Package

TinyGo's "net" package is ported from Go.  The port offers a subset of Go's
"net" package.  The subset maintains Go 1 compatiblity guarantee.  A Go
application that uses "net" will most-likey just work on TinyGo if the usage is
within the subset offered.  (There may be external constraints such as limited
SRAM on some targets that may limit full "net" functionality).

Continue below for details on using "net" and "net/http" packages.

See src/net/READMD.md in the TinyGo repo for more details on maintaining
TinyGo's "net" package.

## Using "net" Package

Ideally, TinyGo's "net" package would be Go's "net" package and applications
using "net" would just work, as-is.  TinyGo's net package is a partial port of
Go's net package, so some things may not work because they have not been
ported.

There are a few features excluded during the porting process, in particular:

- No IPv6 support
- No DualStack support

Run ```go doc -all ./src/net``` in TinyGo repo to see full listing of what has
been ported.  Here is a list of things known to work.  You can find examples
of these at [examples/net](examples/net/).

### What is Known to Work

(These are all IPv4 only).

- TCP client and server
- UDP client
- TLS client
- HTTP client and server
- HTTPS client
- NTP client (UDP)
- MQTT client (paho & natiu)
- WebSocket client and server

Multiple sockets can be opened in a single app.  For example, the app could run
as an http server listen on port :80 and also use NTP to get the current time
or send something over MQTT.  There is a practical limit to the number of
active sockets per app, around 8 or 10, so don't go crazy.

Applications using Go's net package will need a few setup steps to work with
TinyGo's net package.  The steps are required before using "net".

### Step 1: Probe to Load Network Driver

Call Probe() to load the correct network driver for your target.  Probe()
allows the app to work on multiple targets.

```go
package main

import (
	"tinygo.org/x/drivers/netlink/probe"
)

func main() {

	// load network driver for target
	link, dev := probe.Probe()

	...	
}
```

Probe() will load the driver with default configuration for the target.  For
custom configuration, the app can open code Probe() for the target
requirements.

Probe() returns a [Netlinker](netlink/README.md) and a
[Netdever](netdev/README.md), interfaces implemented by the network driver.
Next, we'll use the Netlinker interface to connect the target to an IP network.

### Step 2: Connect to an IP Network

Before the net package is fully functional, we need to connect the target to an
IP network.

```go
package main

import (
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

func main() {

	// load network driver for target
	link, _ := probe.Probe()

	// Connect target to IP network
	link.NetConnect(&netlink.ConnectParams{
		Ssid:       "my SSID",
		Passphrase: "my passphrase",
	})

	// OK to use "net" from here on
	...	
}
```

Optionally, get notified of IP network connects and disconnects:

```go
	link.Notify(func(e netlink.Event) {
		switch e {
		case netlink.EventNetUp:   println("Network UP")
		case netlink.EventNetDown: println("Network DOWN")
	})
```

Here is an example of an http server listening on port :8080:

```go
package main

import (
	"fmt"
	"net/http"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func main() {

	// load network driver for target
	link, _ := probe.Probe()

	// Connect target to IP network
	link.NetConnect(&netlink.ConnectParams{
		Ssid:       "my SSID",
		Passphrase: "my passphrase",
	})

	// Serve it up
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
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
http.Get() to an https:// address, but you cannot http.ListenAndServeTLS() an
https server.

The offloading hardware has pre-defined TLS certificates built-in.

## Using Sockets

The Netdever interface is a BSD socket-like interface so an application can make direct
socket calls, bypassing the "net" package for the lowest overhead.

Here is a simple TCP client application using direct sockets:

```go
package main

import (
	"net"  // only need to parse IP address

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

func main() {

	// load network driver for target
	link, dev := probe.Probe()

	// Connect target to IP network
	link.NetConnect(&netlink.ConnectParams{
		Ssid:       "my SSID",
		Passphrase: "my passphrase",
	})

	// omit error handling

	sock, _ := dev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TCP)

        dev.Connect(sock, "", net.ParseIP("10.0.0.100"), 8080)
	dev.Send(sock, []bytes("hello"), 0, 0)

	dev.Close(sock)
	link.NetDisconnect()
}
```
