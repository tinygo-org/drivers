package stusb4500

import (
	"machine"
	"runtime"
	"sync"
)

// monitorLock is a mutex used to synchronize access to the control flags of a
// Device monitor.
type monitorLock struct {
	sync.Mutex
	active bool
}

func (mon *monitorLock) isActive() bool {
	var active bool
	mon.Lock()
	active = mon.active
	mon.Unlock()
	return active
}

func (mon *monitorLock) setActive(active bool) {
	mon.Lock()
	mon.active = active
	mon.Unlock()
}

// Monitor provides an event loop for maintaining connection to the STUSB4500,
// and calling user callback functions for important state changes.
//
// Since the STUSB4500 may occassionally lose power (such as when a USB Type-C
// cable or PD supply source is disconnected), this will safely re-establish
// a connection and re-initialize the device on subsequent cable attachments.
//
// Note this method will not normally return to the caller. Therefore, this
// method is designed to be run in a goroutine, and it will periodically yield
// to any other goroutines needing processor time (via `runtime.Gosched()`).
//
// Call this Device receiver's `StopMonitor` method to terminate this monitor
// and stop processing all USB PD events.
//
// This routine is implemented entirely with public exported methods from
// package `stusb4500`. Meaning, the user may view this method as a convenience
// function, or as a template for manually managing USB PD connections from
// their own driver package.
func (d *Device) Monitor() error {
	// ensure the monitor has not yet been started
	if d.monitor.isActive() {
		return ErrMonitorStarted
	}
	// verify the STUSB4500 RST pin is connected to GPIO
	if d.config.ResetPin == machine.NoPin {
		return ErrMonitorResetUndef
	}
	// verify the STUSB4500 ALRT pin is connected to GPIO
	if d.config.AlertPin == machine.NoPin {
		return ErrMonitorAlertUndef
	}
	// verify the STUSB4500 ATCH pin is connected to GPIO
	if d.config.AttachPin == machine.NoPin {
		return ErrMonitorAttachUndef
	}
	d.monitor.setActive(true)
	for d.monitor.isActive() {
		// clear alerts, unmask interrupts, initialize registers, reset USB state
		// machines, and verify we can communicate with the expected device.
		if err := d.Initialize(); nil != err {
			if nil != d.config.OnInitFail {
				d.config.OnInitFail(err)
			}
		}
		// check I2C communication is working
		if d.Connected() {
			if nil != d.config.OnConnect {
				d.config.OnConnect()
			}
			// enter the primary state machine run loop
			for d.monitor.isActive() {
				if err := d.Update(); nil != err {
					if nil != d.config.OnError {
						d.config.OnError(err)
					}
					break
				}
				// yield to other goroutines after processing each iteration
				runtime.Gosched()
			}
		} else {
			if nil != d.config.OnConnectFail {
				d.config.OnConnectFail(nil)
			}
		}
		// shouldn't reach here unless something failed. reset the device.
		for d.monitor.isActive() {
			if err := d.Reset(false); nil != err {
				if nil != d.config.OnResetFail {
					d.config.OnResetFail(err)
				}
			} else {
				break
			}
			// yield to other goroutines after each reset attempt
			runtime.Gosched()
		}
		// yield to other goroutines after each closed connection
		runtime.Gosched()
	}
	return nil
}

// StopMonitor signals the monitor event loop to terminate and stop all USB PD
// processing and message handling.
func (d *Device) StopMonitor() error {
	if !d.monitor.isActive() {
		return ErrMonitorNotStarted
	}
	d.monitor.setActive(false)
	return nil
}
