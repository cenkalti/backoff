package backoff

import (
	"math"
	"testing"
)

func TestBackOff(t *testing.T) {
	var testInitialInterval int64 = 500
	var testRandomizationFactor float64 = 0.1
	var testMultiplier float64 = 2.0
	var testMaxInterval int64 = 5000
	var testMaxElapsedTime int64 = 900000

	var exp = NewExponentialBackoff()
	exp.InitialIntervalMillis = testInitialInterval
	exp.RandomizationFactor = testRandomizationFactor
	exp.Multiplier = testMultiplier
	exp.MaxIntervalMillis = testMaxInterval
	exp.MaxElapsedTimeMillis = testMaxElapsedTime
	exp.Reset()

	var expectedResults = []int64{500, 1000, 2000, 4000, 5000, 5000, 5000, 5000, 5000, 5000}
	for _, expected := range expectedResults {
		assertEquals(t, expected, exp.currentIntervalMillis)
		// Assert that the next back off falls in the expected range.
		var minInterval int64 = expected - int64(testRandomizationFactor*float64(expected))
		var maxInterval int64 = expected + int64(testRandomizationFactor*float64(expected))
		var actualInterval int64 = exp.NextBackOffMillis()
		if !(minInterval <= actualInterval && actualInterval <= maxInterval) {
			t.Error("error")
		}
	}
}

func TestGetRandomizedInterval(t *testing.T) {
	// 33% chance of being 1.
	assertEquals(t, 1, getRandomValueFromInterval(0.5, 0, 2))
	assertEquals(t, 1, getRandomValueFromInterval(0.5, 0.33, 2))
	// 33% chance of being 2.
	assertEquals(t, 2, getRandomValueFromInterval(0.5, 0.34, 2))
	assertEquals(t, 2, getRandomValueFromInterval(0.5, 0.66, 2))
	// 33% chance of being 3.
	assertEquals(t, 3, getRandomValueFromInterval(0.5, 0.67, 2))
	assertEquals(t, 3, getRandomValueFromInterval(0.5, 0.99, 2))
}

type MyNanoClock struct {
	i            int64
	startSeconds int64
}

func (c *MyNanoClock) NanoTime() int64 {
	t := (c.startSeconds + c.i) * 1e9
	c.i++
	return t
}

func TestGetElapsedTimeMillis(t *testing.T) {
	var exp = NewExponentialBackoff()
	exp.NanoTimer = &MyNanoClock{}
	exp.Reset()

	var elapsedTimeMillis int64 = exp.GetElapsedTimeMillis()
	if elapsedTimeMillis != 1000 {
		t.Errorf("elapsedTimeMillis=%d", elapsedTimeMillis)
	}
}

func TestMaxElapsedTime(t *testing.T) {
	var exp = NewExponentialBackoff()
	exp.NanoTimer = &MyNanoClock{startSeconds: 10000}
	if exp.NextBackOffMillis() != Stop {
		t.Error("error2")
	}
	// Change the currentElapsedTimeMillis to be 0 ensuring that the elapsed time will be greater
	// than the max elapsed time.
	exp.startTimeNanos = 0
	assertEquals(t, Stop, exp.NextBackOffMillis())
}

func TestBackOffOverflow(t *testing.T) {
	var testInitialInterval int64 = math.MaxInt64 / 2
	var testMultiplier float64 = 2.1
	var testMaxInterval int64 = math.MaxInt64
	var exp = NewExponentialBackoff()
	exp.InitialIntervalMillis = testInitialInterval
	exp.Multiplier = testMultiplier
	exp.MaxIntervalMillis = testMaxInterval
	exp.Reset()

	exp.NextBackOffMillis()
	// Assert that when an overflow is possible the current varerval   int64    is set to the max varerval   int64   .
	assertEquals(t, testMaxInterval, exp.currentIntervalMillis)
}

func assertEquals(t *testing.T, expected, value int64) {
	if expected != value {
		t.Errorf("got: %d, expected: %d", value, expected)
	}
}
