//go:build !mimxrt1062 && !stm32f405 && !atsamd51 && !stm32f103xx && !k210 && !stm32f407

package dht // import "tinygo.org/x/drivers/dht"

// This file provides a definition of the counter for boards with frequency lower than 2^8 ticks per millisecond (<64MHz)
type counter uint16
