package rtl8720dn

type debug uint8

const (
	debugBasic  debug = 1 << iota // show fw version, mac addr, etc
	debugNetdev                   // show netdev entry points
	debugRpc                      // show rtl8720dn cmds

	debugOff = 0
	debugAll = debugBasic | debugNetdev | debugRpc
)

func debugging(want debug) bool {
	return (_debug & want) != 0
}
