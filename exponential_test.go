package backoff

import (
	"math"
	"testing"
	"time"
)

func TestBackOff(t *testing.T) {
	var (
		testInitialInterval     = 500 * time.Millisecond
		testRandomizationFactor = 0.1
		testMultiplier          = 2.0
		testMaxInterval         = 5 * time.Second
	)

	exp := NewExponentialBackOff()
	exp.InitialInterval = testInitialInterval
	exp.RandomizationFactor = testRandomizationFactor
	exp.Multiplier = testMultiplier
	exp.MaxInterval = testMaxInterval
	exp.Reset()

	var expectedResults = []time.Duration{500, 1000, 2000, 4000, 5000, 5000, 5000, 5000, 5000, 5000}
	for i, d := range expectedResults {
		expectedResults[i] = d * time.Millisecond
	}

	for _, expected := range expectedResults {
		assertEquals(t, expected, exp.currentInterval)
		// Assert that the next backoff falls in the expected range.
		var minInterval = expected - time.Duration(testRandomizationFactor*float64(expected))
		var maxInterval = expected + time.Duration(testRandomizationFactor*float64(expected))
		var actualInterval = exp.NextBackOff()
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

func TestBackOffOverflow(t *testing.T) {
	var (
		testInitialInterval time.Duration = math.MaxInt64 / 2
		testMaxInterval     time.Duration = math.MaxInt64
		testMultiplier                    = 2.1
	)

	exp := NewExponentialBackOff()
	exp.InitialInterval = testInitialInterval
	exp.Multiplier = testMultiplier
	exp.MaxInterval = testMaxInterval
	exp.Reset()

	exp.NextBackOff()
	// Assert that when an overflow is possible, the current interval time.Duration is set to the max interval time.Duration.
	assertEquals(t, testMaxInterval, exp.currentInterval)
}

func TestBackOffRandomizedOverflow(t *testing.T) {
	exp := NewExponentialBackOff()
	exp.InitialInterval = math.MaxInt64
	exp.MaxInterval = math.MaxInt64
	exp.RandomizationFactor = 0.5
	exp.Reset()

	for range 100 {
		if next := exp.NextBackOff(); next < 0 {
			t.Fatalf("NextBackOff() = %d, want non-negative", next)
		}
	}
}

func assertEquals(t *testing.T, expected, value time.Duration) {
	if expected != value {
		t.Errorf("got: %d, expected: %d", value, expected)
	}
}
