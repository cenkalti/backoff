package backoff

import "testing"

func TestFactoryFunc(t *testing.T) {
	backoff := NewExponentialBackOff()

	ff := FactoryFunc(func() BackOff {
		return backoff
	})

	newBackOff := ff.NewBackOff()

	expontential, ok := newBackOff.(*ExponentialBackOff)
	if !ok {
		t.Error("wrong type from NewBackOff")
	}

	if *expontential != *backoff {
		t.Error("backoff was not equal to expected backoff")
	}
}
