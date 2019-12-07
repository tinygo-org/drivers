package wifinina

import "time"

func wait(duration time.Duration) {
	newTimer(duration).WaitUntilExpired()
}

type timer struct {
	start    int64
	interval int64
}

func newTimer(interval time.Duration) timer {
	return timer{
		start:    time.Now().UnixNano(),
		interval: int64(interval),
	}
}

func (t timer) Expired() bool {
	return time.Now().UnixNano() > (t.start + t.interval)
}

func (t timer) WaitUntilExpired() {
	for !t.Expired() {
	}
}
