package wifinina

type debug uint8

const (
	debugBasic  debug = 1 << iota // show fw version, mac addr, etc
	debugNetdev                   // show netdev entry points
	debugCmd                      // show non-chatty wifinina cmds
	debugDetail                   // show chatty wifinina cmds

	debugOff = 0
	debugAll = debugBasic | debugNetdev | debugCmd | debugDetail
)

func debugging(want debug) bool {
	return (_debug & want) != 0
}
