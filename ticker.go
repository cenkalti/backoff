package backoff

import (
	"runtime"
	"time"
)

type Ticker struct {
	C    <-chan time.Time
	c    chan time.Time
	b    BackOff
	stop chan struct{}
}

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
