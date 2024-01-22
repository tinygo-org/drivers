### Table of Contents

- [Netlinker](#netlinker)

## Netlinker

TinyGo's network device driver model comprises two Go interfaces: Netdever and
Netlinker.  This README covers Netlinker.

The Netlinker interface describes an L2 network interface.  A netlink is a
concrete implementation of a Netlinker.  See [Netdev](../netdev/) for 
the L4/L3 network interface.

A netlink can:

- Connect/disconnect device to/from network
- Notify of network events (e.g. link UP/DOWN)
- Send and receive Ethernet packets
- Get/set device's hardware address (MAC address)
