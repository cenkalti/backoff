package backoff

import (
	"math/rand"
	"time"
)

/*
ExponentialBackoff is an implementation of BackOff that increases the back off
period for each retry attempt using a randomization function that grows exponentially.

NextBackOffMillis() is calculated using the following formula:

	randomized_interval =
	    retry_interval * (random value in range [1 - randomization_factor, 1 + randomization_factor])

In other words NextBackOffMillis() will range between the randomization factor
percentage below and above the retry interval. For example, using 2 seconds as the base retry
interval and 0.5 as the randomization factor, the actual back off period used in the next retry
attempt will be between 1 and 3 seconds.

Note: max_interval caps the retry_interval and not the randomized_interval.

If the time elapsed since an ExponentialBackOff instance is created goes past the
max_elapsed_time then the method NextBackOffMillis() starts returning backoff.Stop.
The elapsed time can be reset by calling Reset().

Example: The default retry_interval is .5 seconds, default randomization_factor is 0.5, default
multiplier is 1.5 and the default max_interval is 1 minute. For 10 tries the sequence will be
(values in seconds) and assuming we go over the max_elapsed_time on the 10th try:

	request#     retry_interval     randomized_interval

	1             0.5                [0.25,   0.75]
	2             0.75               [0.375,  1.125]
	3             1.125              [0.562,  1.687]
	4             1.687              [0.8435, 2.53]
	5             2.53               [1.265,  3.795]
	6             3.795              [1.897,  5.692]
	7             5.692              [2.846,  8.538]
	8             8.538              [4.269, 12.807]
	9            12.807              [6.403, 19.210]
	10           19.210              backoff.Stop

Implementation is not thread-safe.
*/
type ExponentialBackoff struct {
	InitialIntervalMillis int64
	RandomizationFactor   float64
	Multiplier            float64
	MaxIntervalMillis     int64
	MaxElapsedTimeMillis  int64
	NanoTimer             NanoTimer
	currentIntervalMillis int64
	startTimeNanos        int64
}

type NanoTimer interface {
	NanoTime() int64
}

// Default values for ExponentialBackoff.
const (
	DefaultInitialIntervalMillis int64   = 500
	DefaultRandomizationFactor   float64 = 0.5
	DefaultMultiplier            float64 = 1.5
	DefaultMaxIntervalMillis     int64   = 60000
	DefaultMaxElapsedTimeMillis  int64   = 900000
)

// NewExponentialBackoff creates an instance of ExponentialBackoff using default values.
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialIntervalMillis: DefaultInitialIntervalMillis,
		RandomizationFactor:   DefaultRandomizationFactor,
		Multiplier:            DefaultMultiplier,
		MaxIntervalMillis:     DefaultMaxIntervalMillis,
		MaxElapsedTimeMillis:  DefaultMaxElapsedTimeMillis,
		NanoTimer:             systemTimer{},
	}
}

type systemTimer struct{}

func (t systemTimer) NanoTime() int64 {
	return time.Now().UnixNano()
}

// Reset the interval back to the initial retry interval and restarts the timer.
func (b *ExponentialBackoff) Reset() {
	b.currentIntervalMillis = b.InitialIntervalMillis
	b.startTimeNanos = b.NanoTimer.NanoTime()
}

// NextBackOffMillis calculates the next back off interval using the formula:
// 	randomized_interval = retry_interval +/- (randomization_factor * retry_interval)
func (b *ExponentialBackoff) NextBackOffMillis() int64 {
	// Make sure we have not gone over the maximum elapsed time.
	if b.GetElapsedTimeMillis() > b.MaxElapsedTimeMillis {
		return Stop
	}
	var randomizedInterval int64 = getRandomValueFromInterval(b.RandomizationFactor, rand.Float64(), b.currentIntervalMillis)
	b.incrementCurrentInterval()
	return randomizedInterval
}

// GetElapsedTimeMillis returns the elapsed time in milliseconds since an
// ExponentialBackOff instance is created and is reset when Reset() is called.
//
// The elapsed time is computed using time.Now().UnixNano().
func (b *ExponentialBackoff) GetElapsedTimeMillis() int64 {
	return (b.NanoTimer.NanoTime() - b.startTimeNanos) / 1000000
}

// Increments the current interval by multiplying it with the multiplier.
func (b *ExponentialBackoff) incrementCurrentInterval() {
	// Check for overflow, if overflow is detected set the current interval to the max interval.
	if float64(b.currentIntervalMillis) >= float64(b.MaxIntervalMillis)/b.Multiplier {
		b.currentIntervalMillis = b.MaxIntervalMillis
	} else {
		b.currentIntervalMillis = int64(float64(b.currentIntervalMillis) * b.Multiplier)
	}
}

// Returns a random value from the interval:
// 	[randomizationFactor * currentInterval, randomizationFactor * currentInterval].
func getRandomValueFromInterval(randomizationFactor, random float64, currentIntervalMillis int64) int64 {
	var delta float64 = randomizationFactor * float64(currentIntervalMillis)
	var minInterval float64 = float64(currentIntervalMillis) - delta
	var maxInterval float64 = float64(currentIntervalMillis) + delta
	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	var randomValue int64 = int64(minInterval + (random * (maxInterval - minInterval + 1)))
	return randomValue
}
