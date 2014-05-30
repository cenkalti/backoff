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
//	for t = range ticker.C {
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
// specified by the BackOff argument. BackOff is reset when the Ticker is created.
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

func (t *Ticker) Stop() {
	select {
	case t.stop <- struct{}{}:
	default:
	}
}

func (t *Ticker) run() {
	select {
	case t.c <- time.Now():
	case <-t.stop:
		return
	}

	t.b.Reset()
	for {
		next := t.b.NextBackOff()
		if next == Stop {
			t.Stop()
			return
		}

		select {
		case tick := <-time.After(next):
			t.c <- tick
		case <-t.stop:
			return
		}
	}
}
