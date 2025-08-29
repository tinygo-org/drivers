package w5500

type debug uint8

const (
	debugNetdev debug = 1 << iota // show netdev entry points
	debugDetail                   // show chatty w5500 cmds

	debugOff = 0
	debugAll = debugNetdev | debugDetail
)

func debugging(want debug) bool {
	return (_debug & want) != 0
}
