package backoff

import (
	"testing"
	"time"
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

func TestStopBackOff(t *testing.T) {
	backoff := &StopBackOff{}

	backoffCpy := backoff.NewBackOff()
	_, ok := backoffCpy.(*StopBackOff)
	if !ok {
		t.Error("wrong type from NewBackOff")
	}
}

func TestZeroBackOff(t *testing.T) {
	backoff := &ZeroBackOff{}

	backoffCpy := backoff.NewBackOff()
	_, ok := backoffCpy.(*ZeroBackOff)
	if !ok {
		t.Error("wrong type from NewBackOff")
	}
}

func TestConstantBackOff(t *testing.T) {
	backoff := NewConstantBackOff(time.Second)
	if backoff.NextBackOff() != time.Second {
		t.Error("invalid interval")
	}

	backoffCpy := backoff.NewBackOff()
	constant, ok := backoffCpy.(*ConstantBackOff)
	if !ok {
		t.Error("wrong type from NewBackOff")
	}

	if constant == backoff {
		t.Error("returned backoff is the same as original")
	}

	assertEquals(t, backoff.Interval, constant.Interval)
}
