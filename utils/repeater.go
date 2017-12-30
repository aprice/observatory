package utils

import "time"

// Repeater repeats calls to a channel, allowing for changes in repeat interval
// while running without disturbing cadence.
type Repeater struct {
	Interval time.Duration
	C        SentinelChannel
	timer    *time.Timer
	lastPop  time.Time
}

// StartRepeater starts a new repeater with the given interval. The first call
// to the channel will be after interval.
func StartRepeater(interval time.Duration) *Repeater {
	r := Repeater{
		Interval: interval,
		C:        make(SentinelChannel),
		lastPop:  time.Now(),
	}
	r.timer = time.AfterFunc(interval, r.fire)
	return &r
}

// UpdateInterval updates the interval this Repeater repeats at. The next call
// to the channel will be newInterval from the last time it fired.
func (r *Repeater) UpdateInterval(newInterval time.Duration) {
	if r.Interval == newInterval {
		return
	}
	r.Interval = newInterval
	passed := time.Now().Sub(r.lastPop)
	if newInterval <= passed {
		r.fire()
	} else {
		r.timer.Reset(newInterval - passed)
	}
}

// Stop the repeater and close the channel. The Repeater cannot be reused.
func (r *Repeater) Stop() {
	r.timer.Stop()
	close(r.C)
}

func (r *Repeater) fire() {
	r.timer.Stop()
	r.lastPop = time.Now()
	r.timer = time.AfterFunc(r.Interval, r.fire)
	r.C <- Nothing
}
