package irremote // import "tinygo.org/x/drivers/irremote"

import (
	"errors"
	"machine"
	"sync"
	"time"

	"tinygo.org/x/drivers/irremote/irprotocol"
)

// ErrNoPWM is returned when the SenderConfig has no PWM.
var ErrNoPWM = errors.New("irremote: no PWM")

// PWM is the interface necessary for driving the IR carrier.
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
}

// SenderConfig is used to configure the SenderDevice.
type SenderConfig struct {
	Pin machine.Pin
	PWM PWM
	// DutyCycle is the carrier duty in percent. Values outside 1 to 100 give 33.
	DutyCycle uint8
}

// SenderDevice is the device for sending IR commands.
type SenderDevice struct {
	pwm     PWM
	pin     machine.Pin
	dutyPct uint8
	channel uint8
	duty    uint32
	mu      sync.Mutex
	stop    chan struct{}
}

// NewSender returns a new IR sender device.
func NewSender(config SenderConfig) SenderDevice {
	return SenderDevice{pwm: config.PWM, pin: config.Pin, dutyPct: config.DutyCycle}
}

// Configure sets the PWM to the 38kHz carrier and leaves it off.
func (ir *SenderDevice) Configure() error {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if ir.pwm == nil {
		return ErrNoPWM
	}
	err := ir.pwm.Configure(machine.PWMConfig{Period: uint64(time.Second / irprotocol.NECCarrierHz)})
	if err != nil {
		return err
	}
	ch, err := ir.pwm.Channel(ir.pin)
	if err != nil {
		return err
	}
	ir.channel = ch
	pct := ir.dutyPct
	if pct < 1 || pct > 100 {
		pct = 33
	}
	// Multiply first, so a small PWM top does not truncate the duty to zero.
	ir.duty = uint32(uint64(ir.pwm.Top()) * uint64(pct) / 100)
	ir.pwm.Set(ir.channel, 0)
	return nil
}

// SendNEC sends one NEC frame. With autoRepeat set, repeat frames go out
// until StopNECRepeats is called.
func (ir *SenderDevice) SendNEC(address uint16, command byte, autoRepeat bool) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if ir.stop != nil {
		close(ir.stop)
		ir.stop = nil
	}
	start := time.Now()
	ir.sendNECRawCode(irprotocol.MakeRawNECData(address, command))
	if !autoRepeat {
		return
	}
	stop := make(chan struct{})
	ir.stop = stop
	go func(next time.Time) {
		for {
			time.Sleep(time.Until(next))
			ir.mu.Lock()
			if ir.stop != stop {
				ir.mu.Unlock()
				return
			}
			next = time.Now().Add(irprotocol.NECRepeatPeriod)
			ir.sendNECRepeat()
			ir.mu.Unlock()
		}
	}(start.Add(irprotocol.NECRepeatPeriod))
}

// StopNECRepeats ends the repeat frames started by SendNEC.
func (ir *SenderDevice) StopNECRepeats() {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if ir.stop != nil {
		close(ir.stop)
		ir.stop = nil
	}
}

// SendNECRawCode sends one NEC frame from a 32 bit code, least significant bit
// first, and returns the time it took.
func (ir *SenderDevice) SendNECRawCode(code uint32) time.Duration {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if ir.stop != nil {
		close(ir.stop)
		ir.stop = nil
	}
	return ir.sendNECRawCode(code)
}

// SendNECRepeat sends one repeat frame and returns the time it took.
func (ir *SenderDevice) SendNECRepeat() time.Duration {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	return ir.sendNECRepeat()
}

func (ir *SenderDevice) sendNECRawCode(code uint32) time.Duration {
	start := time.Now()
	ir.mark(irprotocol.NECLeadMark)
	ir.space(irprotocol.NECLeadSpace)
	for i := 0; i < 32; i++ {
		ir.mark(irprotocol.NECBitMark)
		if code&(1<<uint(i)) != 0 {
			ir.space(irprotocol.NECOneSpace)
		} else {
			ir.space(irprotocol.NECZeroSpace)
		}
	}
	ir.mark(irprotocol.NECTrailMark)
	// End on a space, so the next lead mark starts from a rising edge.
	ir.space(irprotocol.NECZeroSpace)
	return time.Since(start)
}

func (ir *SenderDevice) sendNECRepeat() time.Duration {
	start := time.Now()
	ir.mark(irprotocol.NECLeadMark)
	ir.space(irprotocol.NECRepeatSpace)
	ir.mark(irprotocol.NECTrailMark)
	ir.space(irprotocol.NECZeroSpace)
	return time.Since(start)
}

func (ir *SenderDevice) mark(d time.Duration) {
	ir.pwm.Set(ir.channel, ir.duty)
	wait(d)
	ir.pwm.Set(ir.channel, 0)
}

func (ir *SenderDevice) space(d time.Duration) {
	wait(d)
}

func wait(d time.Duration) {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
	}
}
