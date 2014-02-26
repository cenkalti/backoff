package backoff

type BackOff interface {
	NextBackOffMillis() int64
	Reset()
}

const Stop int64 = -1

type ZeroBackOff struct{}

func (b *ZeroBackOff) Reset() {}

func (b *ZeroBackOff) NextBackOffMillis() int64 { return 0 }

type StopBackOff struct{}

func (b *StopBackOff) Reset() {}

func (b *StopBackOff) NextBackOffMillis() int64 { return Stop }
