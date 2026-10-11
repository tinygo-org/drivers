package stusb4500

import (
	"strconv"
	"strings"
)

type PDO struct {
	Number     int
	Voltage    uint32
	Current    uint32
	MaxCurrent uint32
}

var invalidPDO = PDO{}

func (p PDO) Equals(o PDO) bool {
	return (p.Voltage == o.Voltage) &&
		(p.Current == o.Current) &&
		(p.MaxCurrent == o.MaxCurrent)
}

func (p PDO) IsValid() bool {
	return p.Number > 0 && !p.Equals(invalidPDO)
}

func (p PDO) String() string {
	var sb strings.Builder
	sb.WriteString("{ Number: ")
	sb.WriteString(strconv.FormatInt(int64(p.Number), 10))
	sb.WriteString(", Voltage: ")
	sb.WriteString(strconv.FormatUint(uint64(p.Voltage), 10))
	sb.WriteString(" mV, Current: ")
	sb.WriteString(strconv.FormatUint(uint64(p.Current), 10))
	sb.WriteString(" mA, MaxCurrent: ")
	sb.WriteString(strconv.FormatUint(uint64(p.MaxCurrent), 10))
	sb.WriteString(" mA }")
	return sb.String()
}
