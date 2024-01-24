### Table of Contents

- [Netdever](#netdever)
- [Netdev Driver](#netdev-driver)
- [Netdev Driver Notes](#netdev-driver-notes)

## Netdever

TinyGo's network device driver model comprises two Go interfaces: Netdever and
Netlinker.  This README covers Netdever.

The Netdever interface describes an L4/L3 network interface modeled after the
BSD sockets.  A netdev is a concrete implementation of a Netdever.  See
[Netlinker](../netlink/) for the L2 network interface.

A netdev can:

- Send and receive L4/L3 packets
- Resolve DNS lookups
- Get/set the device's IP address

TinyGo network drivers implement the Netdever interface, providing a BSD
sockets interface to TinyGo's "net" package.  net.Conn implementations
(TCPConn, UDPConn, and TLSConn) use the netdev socket.  For example,
net.DialTCP, which returns a net.TCPConn, calls netdev.Socket() and
netdev.Connect():

```go
func DialTCP(network string, laddr, raddr *TCPAddr) (*TCPConn, error) {

        fd, _ := netdev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TCP)

        netdev.Connect(fd, "", raddr.IP, raddr.Port)

        return &TCPConn{
                fd:    fd,
                laddr: laddr,
                raddr: raddr,
        }, nil
}
```

## Setting Netdev

Before the app can use TinyGo's "net" package, the app must set the netdev
using UseNetdev().  This binds the "net" package to the netdev driver.  For
example, setting the wifinina driver as the netdev:

```
	nina := wifinina.New(&cfg)
	netdev.UseNetdev(nina)
```

## Netdev Driver Notes

See the wifinina and rtl8720dn for examples of netdev drivers.  Here are some
notes for netdev drivers.

#### Locking

Multiple goroutines may invoke methods on a net.Conn simultaneously, and since
the net package translates net.Conn calls into netdev socket calls, it follows
that multiple goroutines may invoke socket calls, so locking is required to
keep socket calls from stepping on one another.

Don't hold a lock while Time.Sleep()ing waiting for a long hardware operation to
finish.  Unlocking while sleeping let's other goroutines make progress.

#### Sockfd

The netdev BSD socket interface uses a socket fd (int) to represent a socket
connection (end-point).  Each fd maps 1:1 to a net.Conn maps.  The number of fds
available is a hardware limitation.  Wifinina, for example, can hand out 10
fds, representing 10 active sockets.

#### Testing

The netdev driver should minimally run all of the example/net examples.
