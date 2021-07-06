/*
Package enc28j60 Based on the enc28j60.c file by Guido Socher.
Original file can be found at https://github.com/muanis/arduino-projects under
libraries/etherShield/*

Device information

This device communicates with its host through SPI
by means of a CS (active low) with a clock speed of up to
20 MHz. The IC takes care of several things, mainly at
the Data Link layer (layer 2) such as the first 8 bytes of
the ethernet packet containing the preamble and start-of-frame,
these are generated for outgoing buffer and removed from the
incoming buffer (section 5.1.1).
It also implements CRC sum and can verify incoming packets
and reject faulty ones. It can also generate them for outgoing
packets.

Layer 2 Overview

The Host (microcontroller) is responsible for writing the desired
destination address into the transmit buffer (5.1.2). Users of the ENC
must generate a MAC address to populate the source address field. The
first three bytes are the OUI and are distributed by IEEE. The Host
is responsible for writing the source address too (5.1.3).

The type field specifies the protocol (ARP | IP) or may be treated
as an application specific field for proprietary networks (5.1.4).

The data field is variable length (0 to 1500 bytes). Larger packets may be dropped by nodes (5.1.5).

The padding field is used for small payloads to meet IEEE requirements. An Ethernet packet
must be no smaller than 60 bytes (64 counting the CRC sum). The host must
generate padding as the IC will not add padding to the packet before transmitting. (5.1.6)

The IC will check CRC of incoming packets. Packets with invalid CRC will automatically be
discarded if ERXFCON.CRCEN is set. Host can also determine if recieved packets are valid
by reading the recieve status vector (see section 7.2). CRC sum field generation by IC
is set by default and so should not have to be filled in by Host (5.1.7)

Network frame diagrams

The following sections will detail Layer 2 and layer 3 protocol frames
which are implmented in this code.

Ethernet frame

Diagram:
	0        6        7         13         19        21
	| 7 bytes  | 1 byte |  6 bytes |  6 bytes | 2 bytes | 46-1500 bytes /.../ | 4 bytes   |
	| Preamble | SFD    | Dst Addr | Src Addr | Type    | PAYLOAD       /.../ | FCS (CRC) |

the enc28j60 takes care of the preamble and the SFD. It is by default configured to take care
of the FCS too.

ARP Payload Frame

Address resolution protocol. See https://www.youtube.com/watch?v=aamG4-tH_m8

Legend:
	HW:    Hardware
	AT:    Address type
	AL:    Address Length
	AoS:   Address of sender
	AoT:   Address of Target
	Proto: Protocol

Diagram:
	0      2          4       5          6         8       14          18       24          28
	| HW AT | Proto AT | HW AL | Proto AL | OP Code | HW AoS | Proto AoS | HW AoT | Proto AoT |
	|  2B   |  2B      |  1B   |  1B      | 2B      |   6B   |    4B     |  6B    |   4B      |
	| ethern| IP       |macaddr|          |ask|reply|        |           |for op=1|           |
	| = 1   |=0x0800   |=6     |=4        | 1 | 2   |       known        |=0      |    known  |

*/
package enc28j60
