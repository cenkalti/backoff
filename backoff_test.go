package backoff

import (
	"testing"
)

func TestNextBackOffMillis(t *testing.T) {
	subtestNextBackOffMillis(t, 0, new(ZeroBackOff))
	subtestNextBackOffMillis(t, Stop, new(StopBackOff))
}

func subtestNextBackOffMillis(t *testing.T, expectedValue int64, backOffPolicy BackOff) {
	for i := 0; i < 10; i++ {
		next := backOffPolicy.NextBackOffMillis()
		if next != expectedValue {
			t.Errorf("got: %d expected: %d", next, expectedValue)
		}
	}
}
