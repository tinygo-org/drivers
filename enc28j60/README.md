# Golang implementation of ENC28J60 IC
Ethernet controller with ARP.

*This is a work in progress!*

See [pkg.go.dev](https://pkg.go.dev/tinygo.org/x/drivers) for more information
## Writing this library

When porting an arduino library, it might be of use to know some of the similarities between two languages.
| C++       | Go |
|----       |-----|
| `uint8_t` | `byte` or `uint8` |
| `*uint8_t`| `[]byte` or `[]uint8` |

