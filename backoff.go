// Package backoff implements backoff algorithms for retrying operations.
//
// Also has a Retry() helper for retrying operations that may fail.
package backoff

import "time"

// Back-off policy when retrying an operation.
type BackOff interface {
	// Gets the number of milliseconds to wait before retrying the operation or {@link #STOP} to
	// indicate that no retries should be made.
	//
	// Example usage:
	//
	// 	long backOffMillis = backoff.NextBackOffMillis();
	// 	if (backOffMillis == Backoff.Stop) {
	// 		// do not retry operation
	// 	} else {
	// 		// sleep for backOffMillis milliseconds and retry operation
	// 	}
	//
	NextBackOff() time.Duration

	// Reset to initial state.
	Reset()
}

// Indicates that no more retries should be made for use in NextBackOffMillis().
const Stop time.Duration = time.Duration(-1)

// ZeroBackOff is a fixed back-off policy whose back-off time is always zero,
// meaning that the operation is retried immediately without waiting.
type ZeroBackOff struct{}

func (b *ZeroBackOff) Reset() {}

func (b *ZeroBackOff) NextBackOff() time.Duration { return 0 }

// StopBackOff is a fixed back-off policy that always returns backoff.Stop for
// NextBackOffMillis(), meaning that the operation should not be retried.
type StopBackOff struct{}

func (b *StopBackOff) Reset() {}

func (b *StopBackOff) NextBackOff() time.Duration { return Stop }
