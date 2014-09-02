package backoff

import (
	"time"

	"testing"
)

func TestNextBackOffMillis(t *testing.T) {
	subtestNextBackOff(t, 0, new(ZeroBackOff))
	subtestNextBackOff(t, Stop, new(StopBackOff))
}

func subtestNextBackOff(t *testing.T, expectedValue time.Duration, backOffPolicy BackOff) {
	for i := 0; i < 10; i++ {
		next := backOffPolicy.NextBackOff()
		if next != expectedValue {
			t.Errorf("got: %d expected: %d", next, expectedValue)
		}
	}
}

func TestConstantBackoff(t *testing.T) {
	backoff := NewConstantBackOff()
	if backoff.NextBackOff() != DefaultInitialInterval {
		t.Error("Should have been default interval")
	}
}
