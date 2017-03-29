package backoff

import (
	"time"
)

var (
	_ BackOff = (*JobBackOff)(nil)
)

const (
	DefaultMinJobInterval = 30 * time.Second
)

// JobBackOff is an exponential backoff implementation for long running jobs.
// In long running jobs, an operation() that fails after a long Duration should not increments the backoff period.
// If operation() takes more than MinJobInterval, Reset() is called in NextBackOff().
type JobBackOff struct {
	*ExponentialBackOff
	MinJobInterval time.Duration
}

// NewJobBackOff creates an instance of JobBackOff using default values.
func NewJobBackOff(backOff *ExponentialBackOff) *JobBackOff {
	backOff.MaxElapsedTime = 0
	return &JobBackOff{
		ExponentialBackOff: backOff,
		MinJobInterval:     DefaultMinJobInterval,
	}
}

// NextBackOff calculates the next backoff interval.
func (b *JobBackOff) NextBackOff() time.Duration {
	if b.GetElapsedTime() >= b.MinJobInterval {
		b.Reset()
	}
	return b.ExponentialBackOff.NextBackOff()
}
