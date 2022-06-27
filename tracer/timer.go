package tracer

import (
	"time"
)

// ticker is the same as time.Ticker except that it has jitters.
// A Ticker must be created with newTicker.
type ticker struct {
	tick     *time.Ticker
	duration time.Duration
	jitter   time.Duration
}

// newTicker creates a new Ticker that will send the current time on its channel with the passed jitter.
func newTicker(duration, jitter time.Duration) *ticker {
	t := time.NewTicker(duration - time.Duration(newGlobalRand().Int63n(int64(jitter))))

	jitterTicker := ticker{
		tick:     t,
		duration: duration,
		jitter:   jitter,
	}

	return &jitterTicker
}

// c returns a channel that receives when the ticker fires.
func (j *ticker) c() <-chan time.Time {
	return j.tick.C
}
