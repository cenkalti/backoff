// +build go1.7

package backoff

import (
	"context"
	"errors"
	"testing"
)

func TestRetryWithCanceledContext(t *testing.T) {
	f := func() error {
		t.Error("This function shouldn't be called at all")
		return errors.New("error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RetryNotifyWithContext(ctx, f, NewExponentialBackOff(), nil)
	if err != ctx.Err() {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRetryWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := false
	f := func() error {
		if called {
			t.Error("This function shouldn't be called more than once")
		} else {
			cancel()
			called = true
		}
		return errors.New("error")
	}

	err := RetryNotifyWithContext(ctx, f, NewExponentialBackOff(), nil)
	if err != ctx.Err() {
		t.Errorf("unexpected error: %v", err)
	}
}
