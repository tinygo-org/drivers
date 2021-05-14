package sdcard

import (
	"time"
)

var timeoutTimer [2]timer

type timer struct {
	start   int64
	timeout int64
}

func setTimeout(timerID int, timeout time.Duration) *timer {
	timeoutTimer[timerID].start = time.Now().UnixNano()
	timeoutTimer[timerID].timeout = timeout.Nanoseconds()
	return &timeoutTimer[timerID]
}

func (t timer) expired() bool {
	return time.Now().UnixNano() > (t.start + t.timeout)
}
