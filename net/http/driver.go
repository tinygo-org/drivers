package http

type DeviceDriver interface {
	ListenAndServe(addr string, handler Handler) error
}

var ActiveDevice DeviceDriver

func UseDriver(driver DeviceDriver) {
	// TODO: rethink and refactor this
	if ActiveDevice != nil {
		panic("net.ActiveDevice is already set")
	}
	ActiveDevice = driver
}
