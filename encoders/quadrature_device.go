package encoders

type QuadratureDevice struct {
	cfg  QuadratureConfig
	impl quadratureImpl
}

type QuadratureConfig struct {
	Precision int
}

type quadratureImpl interface {
	configure(cfg QuadratureConfig) error
	readValue() int
	writeValue(int)
}

func (enc *QuadratureDevice) Configure(cfg QuadratureConfig) error {
	if cfg.Precision < 1 {
		cfg.Precision = 4
	}
	enc.cfg = cfg
	return enc.impl.configure(cfg)
}

// Position returns the stored int value for the encoder
func (enc *QuadratureDevice) Position() int {
	return enc.impl.readValue() / enc.cfg.Precision
}

// SetPosition overwrites the currently stored value with the specified int value
func (enc *QuadratureDevice) SetPosition(v int) {
	enc.impl.writeValue(v * enc.cfg.Precision)
}
