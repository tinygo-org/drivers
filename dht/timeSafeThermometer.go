package dht

import "time"

type Device interface {
	DummyDevice
	Configure(policy UpdatePolicy)
}

type managedDevice struct {
	t          device
	lastUpdate time.Time
	policy     UpdatePolicy
}

func (m *managedDevice) Measurements() (temperature int16, humidity uint16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, 0, err
	}
	return m.t.Measurements()
}

func (m *managedDevice) Temperature() (temp int16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	temp, err = m.t.Temperature()
	return
}

func (m *managedDevice) checkForUpdateOnDataRequest() (err error) {
	// update if necessary
	if m.policy.UpdateAutomatically {
		err = m.ReadMeasurements()
	}
	// ignore error if the data was updated recently
	// interface comparison does not work in tinygo. Therefore need to cast to explicit type
	if code, ok := err.(ErrorCode); ok && code == UpdateError {
		err = nil
	}
	// add error if the data is not initialized
	if !m.t.initialized {
		err = UninitializedDataError
	}
	return err
}

func (m *managedDevice) TemperatureFloat(scale TemperatureScale) (float32, error) {
	err := m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.TemperatureFloat(scale)
}

func (m *managedDevice) Humidity() (hum uint16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.Humidity()
}

func (m *managedDevice) HumidityFloat() (float32, error) {
	err := m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.HumidityFloat()
}

func (m *managedDevice) ReadMeasurements() (err error) {
	timestamp := time.Now()
	if !m.t.initialized || timestamp.Sub(m.lastUpdate) > m.policy.UpdateTime {
		err = m.t.ReadMeasurements()
	} else {
		err = UpdateError
	}
	if err == nil {
		m.lastUpdate = timestamp
	}
	return
}
func (m *managedDevice) Configure(policy UpdatePolicy) {
	if policy.UpdateAutomatically && policy.UpdateTime < time.Second*2 {
		policy.UpdateTime = time.Second * 2
	}
	m.policy = policy
}

func New(pin machine.Pin, deviceType DeviceType) Device {
	return &managedDevice{
		t: device{
			pin:          pin,
			measurements: deviceType,
			initialized:  false,
		},
		lastUpdate: time.Time{},
		policy: UpdatePolicy{
			UpdateTime:          time.Second * 2,
			UpdateAutomatically: true,
		},
	}
}

func NewWithPolicy(pin machine.Pin, deviceType DeviceType, updatePolicy UpdatePolicy) Device {
	result := &managedDevice{
		t: device{
			pin:          pin,
			measurements: deviceType,
			initialized:  false,
		},
		lastUpdate: time.Time{},
	}
	result.Configure(updatePolicy)
	return result
}
