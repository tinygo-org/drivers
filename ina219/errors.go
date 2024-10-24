package ina219

type ErrOverflow struct{}

func (e ErrOverflow) Error() string { return "overflow" }

type ErrNotReady struct{}

func (e ErrNotReady) Error() string { return "not ready" }

type ErrConfigMismatch struct{}

func (e ErrConfigMismatch) Error() string { return "config mismatch" }
