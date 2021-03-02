package dht

import "time"

type managedDevice struct {
	t           device
	lastUpdate  time.Time
	policy      UpdatePolicy
	initialized bool
}

func (m *managedDevice) Measurements() (temperature int16, humidity uint16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, 0, err
	}
	return m.t.Temperature(), m.t.Humidity(), nil
}

func (m *managedDevice) Temperature() (temp int16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	temp = m.t.Temperature()
	return
}

func (m *managedDevice) checkForUpdateOnDataRequest() (err error) {
	// update if necessary
	if m.policy.UpdateAutomatically {
		err = m.ReadMeasurements()
	}
	// ignore error if the data was updated recently
	if err == updateError {
		err = nil
	}
	// add error if the data is not initialized
	if !m.initialized {
		err = uninitializedData
	}
	return err
}

func (m *managedDevice) TemperatureFloat(scale TemperatureScale) (float32, error) {
	err := m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.TemperatureFloat(scale), err
}

func (m *managedDevice) Humidity() (hum uint16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.Humidity(), err
}

func (m *managedDevice) HumidityFloat() (float32, error) {
	err := m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.HumidityFloat(), err
}

func (m *managedDevice) ReadMeasurements() (err error) {
	timestamp := time.Now()
	if !m.initialized || timestamp.Sub(m.lastUpdate) > m.policy.UpdateTime {
		err = m.t.ReadMeasurements()
	} else {
		err = updateError
	}
	if err == nil {
		m.initialized = true
		m.lastUpdate = timestamp
	}
	return
}
