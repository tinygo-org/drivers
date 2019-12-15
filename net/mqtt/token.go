package mqtt

import "time"

type mqtttoken struct {
	err error
}

func (t *mqtttoken) Wait() bool {
	return true
}

func (t *mqtttoken) WaitTimeout(time.Duration) bool {
	return true
}

func (t *mqtttoken) Error() error {
	return t.err
}
