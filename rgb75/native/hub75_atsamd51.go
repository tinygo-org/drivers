// +build sam,atsamd51

// Hardware interface implementation of GPIO-driven HUB75 RGB LED matrix panels
// for SAMD51 targets.
package native

import (
	"device/sam"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

// Timer/counter peripherals configuration
const (
	TC_ROW      = 4           // timer peripheral for row data transmission (TC4)
	TC_ROW_IRQ  = sam.IRQ_TC4 // interrupt ID (TC4)
	TC_ROW_FREQ = 48000000    // timer peripheral clock frequency (48 MHz)
)

// interruptTimer holds the fields necessary to initialize and control a SAMD51
// timer/counter (TC) peripheral.
type interruptTimer struct {
	id int
	ok bool
	bc *volatile.Register32
	cm uint32
	tc *sam.TC_COUNT32_Type
}

var rowTimer interruptTimer

// timer holds a reference, IRQ, bus clock (with mask), and GCLK ID for each
// SAMD51 general-purpose timer/counter (TC) peripheral in 32-bit counter mode.
//
// For HUB75, we use the 32-bit counter mode (TC_COUNT32_Type) instead of the
// 8-bit (TC_COUNT8_Type) or 16-bit (TC_COUNT16_Type) modes; the 32-bit mode is
// actually implemented as two 16-bit counters, where TC[n] 32-bit would use
// both TC[n] and TC[n+1] (e.g., TC2 32-bit requires both TC2 and TC3).
func init() {
	timer := []interruptTimer{
		{
			id: sam.PCHCTRL_GCLK_TC0,
			ok: false,
			bc: &sam.MCLK.APBAMASK,
			cm: sam.MCLK_APBAMASK_TC0_,
			tc: sam.TC0_COUNT32,
		},
		{
			id: sam.PCHCTRL_GCLK_TC1,
			ok: false,
			bc: &sam.MCLK.APBAMASK,
			cm: sam.MCLK_APBAMASK_TC1_,
			tc: sam.TC1_COUNT32,
		},
		{
			id: sam.PCHCTRL_GCLK_TC2,
			ok: false,
			bc: &sam.MCLK.APBBMASK,
			cm: sam.MCLK_APBBMASK_TC2_,
			tc: sam.TC2_COUNT32,
		},
		{
			id: sam.PCHCTRL_GCLK_TC3,
			ok: false,
			bc: &sam.MCLK.APBBMASK,
			cm: sam.MCLK_APBBMASK_TC3_,
			tc: sam.TC3_COUNT32,
		},
		{
			id: sam.PCHCTRL_GCLK_TC4,
			ok: false,
			bc: &sam.MCLK.APBCMASK,
			cm: sam.MCLK_APBCMASK_TC4_,
			tc: sam.TC4_COUNT32,
		},
		{
			id: sam.PCHCTRL_GCLK_TC5,
			ok: false,
			bc: &sam.MCLK.APBCMASK,
			cm: sam.MCLK_APBCMASK_TC5_,
			tc: sam.TC5_COUNT32,
		},
	}

	rowTimer = timer[TC_ROW]
}

// SetPins configures the pre-computed GPIO port bitmasks for the HUB75 data and
// control pins.
func (hub *hub75) SetPins(rgb [6]machine.Pin, clk machine.Pin, addr ...machine.Pin) {

	hub.maskR1 = uint32(1 << (rgb[0] & 0x1F))
	hub.maskG1 = uint32(1 << (rgb[1] & 0x1F))
	hub.maskB1 = uint32(1 << (rgb[2] & 0x1F))
	hub.maskR2 = uint32(1 << (rgb[3] & 0x1F))
	hub.maskG2 = uint32(1 << (rgb[4] & 0x1F))
	hub.maskB2 = uint32(1 << (rgb[5] & 0x1F))
	hub.maskRGB = hub.maskR1 | hub.maskG1 | hub.maskB1 |
		hub.maskR2 | hub.maskG2 | hub.maskB2
	hub.groupRGB = uint32(rgb[0]) >> 5

	if clk != machine.NoPin {
		hub.maskCLK = uint32(1 << (clk & 0x1F))
		hub.maskRGB |= hub.maskCLK
	}

	if len(addr) > 0 {
		hub.maskA = uint32(1 << (addr[0] & 0x1F))
		hub.maskAddr |= hub.maskA
		hub.groupAddr = uint32(addr[0]) >> 5
	}
	if len(addr) > 1 {
		hub.maskB = uint32(1 << (addr[1] & 0x1F))
		hub.maskAddr |= hub.maskB
	}
	if len(addr) > 2 {
		hub.maskC = uint32(1 << (addr[2] & 0x1F))
		hub.maskAddr |= hub.maskC
	}
	if len(addr) > 3 {
		hub.maskD = uint32(1 << (addr[3] & 0x1F))
		hub.maskAddr |= hub.maskD
	}
	if len(addr) > 4 {
		hub.maskE = uint32(1 << (addr[4] & 0x1F))
		hub.maskAddr |= hub.maskE
	}
}

// SetRgb sets/clears each of the 6 RGB data pins.
func (hub *hub75) SetRgb(r1, g1, b1, r2, g2, b2 bool) {
	var data uint32
	if r1 {
		data |= hub.maskR1
	}
	if g1 {
		data |= hub.maskG1
	}
	if b1 {
		data |= hub.maskB1
	}
	if r2 {
		data |= hub.maskR2
	}
	if g2 {
		data |= hub.maskG2
	}
	if b2 {
		data |= hub.maskB2
	}
	hub.SetRgbMask(data)
}

// SetRgbMask sets/clears each of the 6 RGB data pins from the given bitmask.
func (hub *hub75) SetRgbMask(data uint32) {
	// replace the first 6 bits in PORT register OUT with the corresponding bits
	// from the given data, updating R1,G1,B1,R2,G2,B2 all simultaneously
	sam.PORT.GROUP[hub.groupRGB].OUT.ReplaceBits(data, hub.maskRGB, 0)
}

// ClkRgb sets CLK, sets/clears each of the 6 RGB data pins, then clears CLK.
// Note that this method is only permitted when RGB data pins and CLK pin are
// all on the same GPIO port.
func (hub *hub75) ClkRgb(r1, g1, b1, r2, g2, b2 bool) {
	var data uint32
	if r1 {
		data |= hub.maskR1
	}
	if g1 {
		data |= hub.maskG1
	}
	if b1 {
		data |= hub.maskB1
	}
	if r2 {
		data |= hub.maskR2
	}
	if g2 {
		data |= hub.maskG2
	}
	if b2 {
		data |= hub.maskB2
	}
	hub.ClkRgbMask(data)
}

// ClkRgbMask sets CLK, sets/clears each of the 6 RGB data pins from the given
// bitmask, then clears CLK.
// Note that this method is only permitted when RGB data pins and CLK pin are
// all on the same GPIO port.
func (hub *hub75) ClkRgbMask(data uint32) {
	data &= hub.maskRGB
	sam.PORT.GROUP[hub.groupRGB].OUTTGL.Set(hub.maskCLK | data)
	sam.PORT.GROUP[hub.groupRGB].OUTTGL.Set(hub.maskCLK)
	sam.PORT.GROUP[hub.groupRGB].OUTTGL.Set(data)
}

// SetRow sets the active pair of data rows with the given index.
func (hub *hub75) SetRow(row int) {
	var addr uint32
	if 0 != row&(1<<0) { // 0x01
		addr |= hub.maskA
	}
	if 0 != row&(1<<1) { // 0x02
		addr |= hub.maskB
	}
	if 0 != row&(1<<2) { // 0x04
		addr |= hub.maskC
	}
	if 0 != row&(1<<3) { // 0x08
		addr |= hub.maskD
	}
	if 0 != row&(1<<4) { // 0x10
		addr |= hub.maskE
	}
	sam.PORT.GROUP[hub.groupAddr].OUT.ReplaceBits(addr, hub.maskAddr, 0)
}

// GetPinGroupAlignment returns true if and only if all given Pins are on the
// same GPIO port, and returns the minimum size of the group to which all pins
// belong (8, 16, or 32 if true, otherwise 0). Returns (true, 0) if no Pins
// are provided.
func (hub *hub75) GetPinGroupAlignment(pin ...machine.Pin) (samePort bool, alignment uint8) {

	// return a unique condition if no pins provided
	if len(pin) == 0 {
		return true, 0
	}

	var (
		bitMask  uint32 // position of each pin
		byteMask uint8  // byte-wise grouping of pins
		group    uint8  // common GPIO PORT group
	)

	for i, p := range pin {
		// see: `(Pin).getPinGrouping` in package "machine"
		grp, pos := uint8(p)>>5, uint8(p)&0x1f
		if i == 0 {
			group = grp
		} else {
			if group != grp {
				return false, 0 // error: all pins not on same GPIO PORT
			}
		}
		bitMask |= 1 << pos
	}

	// if we've reached here, all pins are on the same GPIO PORT. now we need to
	// decide if they are all within an 8-, 16-, or 32-bit group.

	if bitMask&0x000000FF != 0 {
		byteMask |= 1 << 0 // 0x1
	}
	if bitMask&0x0000FF00 != 0 {
		byteMask |= 1 << 1 // 0x2
	}
	if bitMask&0x00FF0000 != 0 {
		byteMask |= 1 << 2 // 0x4
	}
	if bitMask&0xFF000000 != 0 {
		byteMask |= 1 << 3 // 0x8
	}

	// the above can be written as the following loop, but I don't think it's as
	// clear to the reader what we're doing.
	//for i, m := 0, uint32(0xFF); i < 4; i, m = i+1, m<<8 {
	//  if 0 != bitMask&m {
	//    byteMask |= 1 << i
	//  }
	//}

	// we have determined which group(s) of bytes to which all of the pins belong,
	// now we just have to count the number of adjacent groups there are.

	switch byteMask {
	case 0x1, 0x2, 0x4, 0x8:
		// all pins are in the same byte (0b0001, 0b0010, 0b0100, 0b1000)
		return true, 8 // 8-bit alignment

	case 0x3, 0x6, 0xC:
		// all pins are in the same word (0b0011, 0b0110, 0b1100)
		return true, 16 // 16-bit alignment

	default:
		// otherwise, the pins are spread out all across the register
		return true, 32 // 32-bit alignment
	}
}

// InitTimer is used to initialize a timer service that fires an interrupt at
// regular frequency, which, for HUB75, is used to signal row data transmission.
// The timer does not begin raising interrupts until ResumeTimer is called.
func (hub *hub75) InitTimer(handle func()) {

	// set the actual interrupt handler for row data transmission
	hub.handleRow = handle

	// disable clock source
	sam.GCLK.PCHCTRL[rowTimer.id].ClearBits(sam.GCLK_PCHCTRL_CHEN)
	for sam.GCLK.PCHCTRL[rowTimer.id].HasBits(sam.GCLK_PCHCTRL_CHEN) {
	} // wait for it to disable

	// run timer off of GCLK1
	sam.GCLK.PCHCTRL[rowTimer.id].ReplaceBits(
		sam.GCLK_PCHCTRL_GEN_GCLK1,
		sam.GCLK_PCHCTRL_GEN_Msk,
		sam.GCLK_PCHCTRL_GEN_Pos)

	// enable clock source
	sam.GCLK.PCHCTRL[rowTimer.id].SetBits(sam.GCLK_PCHCTRL_CHEN)
	for !sam.GCLK.PCHCTRL[rowTimer.id].HasBits(sam.GCLK_PCHCTRL_CHEN) {
	} // wait for it to enable

	// disable timer before configuring
	rowTimer.tc.CTRLA.ClearBits(sam.TC_COUNT32_CTRLA_ENABLE)
	for rowTimer.tc.SYNCBUSY.HasBits(sam.TC_COUNT32_SYNCBUSY_ENABLE) {
	} // wait for it to disable

	// enable the TC bus clock
	rowTimer.bc.SetBits(rowTimer.cm)

	// use 32-bit counter mode, DIV1 prescalar (1:1)
	mode, pdiv :=
		uint32(sam.TC_COUNT32_CTRLA_MODE_COUNT32<<
			sam.TC_COUNT32_CTRLA_MODE_Pos)&
			sam.TC_COUNT32_CTRLA_MODE_Msk,
		uint32(sam.TC_COUNT32_CTRLA_PRESCALER_DIV1<<
			sam.TC_COUNT32_CTRLA_PRESCALER_Pos)&
			sam.TC_COUNT32_CTRLA_PRESCALER_Msk
	rowTimer.tc.CTRLA.SetBits(mode | pdiv)

	// use match frequency (MFRQ) mode
	rowTimer.tc.WAVE.Set((sam.TC_COUNT32_WAVE_WAVEGEN_MFRQ <<
		sam.TC_COUNT32_WAVE_WAVEGEN_Pos) &
		sam.TC_COUNT32_WAVE_WAVEGEN_Msk)

	// use up-counter
	rowTimer.tc.CTRLBCLR.Set(sam.TC_COUNT32_CTRLBCLR_DIR)
	for rowTimer.tc.SYNCBUSY.HasBits(sam.TC_COUNT32_SYNCBUSY_CTRLB) {
	}

	// interrupt on overflow
	rowTimer.tc.INTENSET.Set(sam.TC_COUNT32_INTENSET_OVF)
	rowTimer.tc.INTFLAG.Set(sam.TC_COUNT32_INTFLAG_OVF) // clear all pending

	// install interrupt handler
	in := interrupt.New(TC_ROW_IRQ, handleTimer)
	//in.SetPriority(0) // use highest priority
	in.Enable()
}

// ResumeTimer resumes the timer service, with given current value, that signals
// row data transmission for HUB75 by raising interrupts with given periodicity.
func (hub *hub75) ResumeTimer(value, period uint32) {

	// don't do anything if we already resumed
	if rowTimer.ok {
		return
	}

	// reset the current counter value
	rowTimer.tc.COUNT.Set(value)
	for rowTimer.tc.SYNCBUSY.HasBits(sam.TC_COUNT32_SYNCBUSY_COUNT) {
	} // wait for counter sync

	// set the given period. note that if period is less than current counter
	// value, the counter will continue counting up until it has overflowed the
	// 32-bit storage, and then has to count back up to overflow period before it
	// will raise the next interrupt.
	rowTimer.tc.CC[0].Set(period)
	for rowTimer.tc.SYNCBUSY.HasBits(sam.TC_COUNT32_SYNCBUSY_CC0) {
	} // wait for period sync

	// enable the counter
	rowTimer.tc.CTRLA.Set(sam.TC_COUNT32_CTRLA_ENABLE)
	for rowTimer.tc.SYNCBUSY.HasBits(sam.TC_COUNT32_SYNCBUSY_ENABLE) {
	} // wait for it to enable

	rowTimer.ok = true // timer has resumed
}

// PauseTimer pauses the timer service that signals row data transmission for
// HUB75 and returns the current value of the timer.
func (hub *hub75) PauseTimer() uint32 {

	// don't do anything if we are already paused
	if !rowTimer.ok {
		return 0
	}

	// request a synchronized read
	rowTimer.tc.CTRLBSET.Set((sam.TC_COUNT32_CTRLBSET_CMD_READSYNC <<
		sam.TC_COUNT32_CTRLBSET_CMD_Pos) &
		sam.TC_COUNT32_CTRLBSET_CMD_Msk)
	for rowTimer.tc.SYNCBUSY.HasBits(sam.TC_COUNT32_SYNCBUSY_CTRLB) {
	} // wait for command sync
	val := rowTimer.tc.COUNT.Get() // now safe to read

	// disable the counter
	rowTimer.tc.CTRLA.ClearBits(sam.TC_COUNT32_CTRLA_ENABLE)
	for rowTimer.tc.SYNCBUSY.HasBits(sam.TC_COUNT32_SYNCBUSY_STATUS) {
	} // wait for it to disable

	rowTimer.ok = false // timer is now paused

	return val
}

// handleTimer is a wrapper interrupt service routine (ISR) that simply calls
// the user-provided handler for row data transmission on each timer interrupt.
//
// The interrupt.New method requires constants for both IRQ and ISR, thus we
// can't use the argument to InitTimer as handler and must use this indirection
// as a constant handler.
func handleTimer(interrupt.Interrupt) {
	rowTimer.tc.INTFLAG.Set(sam.TC_COUNT32_INTFLAG_OVF) // clear interrupt
	HUB75.handleRow()
}
