package backoff

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"testing"
	"time"
)

type testTimer struct {
	timer *time.Timer
}

func (t *testTimer) Start(duration time.Duration) {
	t.timer = time.NewTimer(0)
}

func (t *testTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

func (t *testTimer) C() <-chan time.Time {
	return t.timer.C
}

func TestRetry(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return nil
		}

		log.Println("error")
		return errors.New("error")
	}

	err := RetryNotifyWithTimer(f, NewExponentialBackOff(), nil, &testTimer{})
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryWithData(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() (int, error) {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return 42, nil
		}

		log.Println("error")
		return 1, errors.New("error")
	}

	res, err := RetryNotifyWithTimerAndData(f, NewExponentialBackOff(), nil, &testTimer{})
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
	if res != 42 {
		t.Errorf("invalid data in response: %d, expected 42", res)
	}
}

func TestRetryContext(t *testing.T) {
	var cancelOn = 3
	var i = 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This function cancels context on "cancelOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		// cancelling the context in the operation function is not a typical
		// use-case, however it allows to get predictable test results.
		if i == cancelOn {
			cancel()
		}

		log.Println("error")
		return fmt.Errorf("error (%d)", i)
	}

	err := RetryNotifyWithTimer(f, WithContext(NewConstantBackOff(time.Millisecond), ctx), nil, &testTimer{})
	if err == nil {
		t.Errorf("error is unexpectedly nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != cancelOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryPermanent(t *testing.T) {
	ensureRetries := func(test string, shouldRetry bool, f func() (int, error), expectRes int) {
		numRetries := -1
		maxRetries := 1

		res, _ := RetryNotifyWithTimerAndData(
			func() (int, error) {
				numRetries++
				if numRetries >= maxRetries {
					return -1, Permanent(errors.New("forced"))
				}
				return f()
			},
			NewExponentialBackOff(),
			nil,
			&testTimer{},
		)

		if shouldRetry && numRetries == 0 {
			t.Errorf("Test: '%s', backoff should have retried", test)
		}

		if !shouldRetry && numRetries > 0 {
			t.Errorf("Test: '%s', backoff should not have retried", test)
		}

		if res != expectRes {
			t.Errorf("Test: '%s', got res %d but expected %d", test, res, expectRes)
		}
	}

	for _, testCase := range []struct {
		name        string
		f           func() (int, error)
		shouldRetry bool
		res         int
	}{
		{
			"nil test",
			func() (int, error) {
				return 1, nil
			},
			false,
			1,
		},
		{
			"io.EOF",
			func() (int, error) {
				return 2, io.EOF
			},
			true,
			-1,
		},
		{
			"Permanent(io.EOF)",
			func() (int, error) {
				return 3, Permanent(io.EOF)
			},
			false,
			3,
		},
		{
			"Wrapped: Permanent(io.EOF)",
			func() (int, error) {
				return 4, fmt.Errorf("Wrapped error: %w", Permanent(io.EOF))
			},
			false,
			4,
		},
	} {
		ensureRetries(testCase.name, testCase.shouldRetry, testCase.f, testCase.res)
	}
}

func TestPermanent(t *testing.T) {
	want := errors.New("foo")
	other := errors.New("bar")
	var err error = Permanent(want)

	got := errors.Unwrap(err)
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if is := errors.Is(err, want); !is {
		t.Errorf("err: %v is not %v", err, want)
	}

	if is := errors.Is(err, other); is {
		t.Errorf("err: %v is %v", err, other)
	}

	wrapped := fmt.Errorf("wrapped: %w", err)
	var permanent *PermanentError
	if !errors.As(wrapped, &permanent) {
		t.Errorf("errors.As(%v, %v)", wrapped, permanent)
	}

	err = Permanent(nil)
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}
}
