package l3gd20

import "errors"

var (
	ErrBadIdentity = errors.New("got unexpected identity from WHOMAI")
	ErrBadRange    = errors.New("bad range configuration value")
)

// Sensitivity factors
const (
	// range at bits 4-5
	// 00 = 250 dps
	// 01 = 500 dps
	// 10 = 2000 dps
	// 11 = 2000 dps
	rangePos   = 4
	Range_250  = 0b00 << rangePos // 8.75 mdps/digit
	Range_500  = 0b01 << rangePos // 17.5 mdps/digit
	Range_2000 = 0b11 << rangePos // 70 mdps/digit
	rangebits  = Range_250 | Range_500 | Range_2000

	// Sensitivities for degrees
	sensDiv250dps  = 800
	sensDiv500dps  = 400
	sensDiv2000dps = 100
	sens_250       = 7. / sensDiv250dps  // Sensitivity at 250 dps
	sens_500       = 7. / sensDiv500dps  // Sensitivity at 500 dps
	sens_2000      = 7. / sensDiv2000dps // Sensitivity at 500 dp

	// 1e6*Pi/180. = 17453.292519943298 (constant for Degree to micro radians conversion)
	// sensitivities for radians
	sensMul250  = 7 * 1745329 / 100 / sensDiv250dps
	sensMul500  = 7 * 1745329 / 100 / sensDiv500dps
	sensMul2000 = 7 * 1745329 / 100 / sensDiv2000dps
)

type Config struct {
	Range uint8
}

// validate scans config for invalid data and returns non-nil
// error indicating what data must be modified.
func (cfg *Config) validate() error {
	if cfg.Range&^rangebits != 0 && cfg.Range != 1 {
		return ErrBadRange
	}
	return nil
}
