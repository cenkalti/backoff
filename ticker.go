package backoff

import (
	"runtime"
	"time"
)

// Ticker holds a channel that delivers `ticks' of a clock at times reported by a BackOff.
//
// Usage:
// 	operation := func() error {
// 		// An operation that may fail
// 	}
//
//	b := backoff.NewExponentialBackOff()
//	ticker := backoff.NewTicker(b)
//
// 	var err error
//	for _ = range ticker.C {
//		if err = operation(); err != nil {
//			log.Println(err, "will retry...")
//			continue
//		}
//
//		break
//	}
//
// 	if err != nil {
// 		// Operation has failed.
// 	}
//
// 	// Operation is successfull.
//
type Ticker struct {
	C    <-chan time.Time
	c    chan time.Time
	b    BackOff
	stop chan struct{}
}

// NewTicker returns a new Ticker containing a channel that will send the time at times
// specified by the BackOff argument. Ticker is guaranteed to tick at least once.
// The channel is closed when Stop method is called or BackOff stops.
func NewTicker(b BackOff) *Ticker {
	c := make(chan time.Time)
	t := &Ticker{
		C:    c,
		c:    c,
		b:    b,
		stop: make(chan struct{}, 1),
	}
	go t.run()
	runtime.SetFinalizer(t, func(x *Ticker) { x.Stop() })
	return t
}

// Stop turns off a ticker. After Stop, no more ticks will be sent.
func (t *Ticker) Stop() {
	select {
	case t.stop <- struct{}{}:
	default:
	}
}

func (t *Ticker) run() {
	var next time.Duration
	var afterC <-chan time.Time

	defer close(t.c)

	first := make(chan time.Time, 1)
	first <- time.Now()

	for {
		var tick time.Time

		select {
		case tick = <-first:
			t.c <- tick
			t.b.Reset()
			next = t.b.NextBackOff()
			if next == Stop {
				return
			}
			afterC = time.After(next)
		case tick = <-afterC:
			t.c <- tick
			next = t.b.NextBackOff()
			if next == Stop {
				return
			}
			afterC = time.After(next)
		case <-t.stop:
			return
		}
	}
}
