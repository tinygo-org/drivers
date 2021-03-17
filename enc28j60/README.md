# Golang implementation of ENC28J60 IC
Ethernet controller with ARP.
## Writing this library

When porting an arduino library, it might be of use to know some of the similarities between two languages.
| C++       | Go |
|----       |-----|
| `uint8_t` | `byte` or `uint8` |
| `*uint8_t`| `[]byte` or `[]uint8` |

