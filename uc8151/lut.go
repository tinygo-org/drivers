package uc8151

// LUTType is the look-up table for the display
type LUTType [42]uint8

type LUTSet struct {
	VCOM LUTType
	WW   LUTType
	BW   LUTType
	WB   LUTType
	BB   LUTType
}

func (lut *LUTType) Clear() {
	for i := range lut {
		lut[i] = 0
	}
}

func (lut *LUTType) SetRow(row int, pat uint8, dur [4]uint8, rep uint8) error {
	index := row * 6
	lut[index] = pat
	lut[index+1] = dur[0]
	lut[index+2] = dur[1]
	lut[index+3] = dur[2]
	lut[index+4] = dur[3]
	lut[index+5] = rep

	return nil
}
