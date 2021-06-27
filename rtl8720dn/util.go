package rtl8720dn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

type httpHeader []byte

func (h httpHeader) ContentLength() int {
	contentLength := -1
	idx := bytes.Index(h, []byte("Content-Length: "))
	if 0 <= idx {
		_, err := fmt.Sscanf(string(h[idx+16:]), "%d", &contentLength)
		if err != nil {
			return -1
		}
	}
	return contentLength
}

// TODO: IPAddress implementation should be moved under drivers/net
// The same implementation exists in wifinina.
type IPAddress []byte

func (addr IPAddress) String() string {
	if len(addr) < 4 {
		return ""
	}
	return strconv.Itoa(int(addr[0])) + "." + strconv.Itoa(int(addr[1])) + "." + strconv.Itoa(int(addr[2])) + "." + strconv.Itoa(int(addr[3]))
}

func ParseIPv4(s string) (IPAddress, error) {
	v := strings.Split(s, ".")
	v0, _ := strconv.Atoi(v[0])
	v1, _ := strconv.Atoi(v[1])
	v2, _ := strconv.Atoi(v[2])
	v3, _ := strconv.Atoi(v[3])
	return IPAddress([]byte{byte(v0), byte(v1), byte(v2), byte(v3)}), nil
}

func (addr IPAddress) AsUint32() uint32 {
	if len(addr) < 4 {
		return 0
	}
	b := []byte(string(addr))
	return binary.BigEndian.Uint32(b[0:4])
}
