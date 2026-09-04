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

// TestPermanentIsByIdentity guards against the regression described in
// issue #186: PermanentError.Is previously returned true for any
// *PermanentError target, which made errors.Is(Permanent(A), Permanent(B))
// incorrectly report a match. Two distinct wrapped sentinels must remain
// distinguishable.
func TestPermanentIsByIdentity(t *testing.T) {
	sentinelA := errors.New("sentinel A")
	sentinelB := errors.New("sentinel B")

	permA := Permanent(sentinelA)
	permB := Permanent(sentinelB)

	// Two calls to Permanent with the same wrapped error must produce
	// distinct *PermanentError instances, so errors.Is must not equate them.
	permA2 := Permanent(sentinelA)
	if errors.Is(permA, permA2) {
		t.Errorf("errors.Is(permA, permA2): different *PermanentError instances wrapping the same error must not match")
	}

	// Different wrapped errors must never match.
	if errors.Is(permA, permB) {
		t.Errorf("errors.Is(Permanent(A), Permanent(B)): distinct wrapped errors matched (issue #186)")
	}
	if errors.Is(permB, permA) {
		t.Errorf("errors.Is(Permanent(B), Permanent(A)): distinct wrapped errors matched (issue #186)")
	}

	// Identity must still match itself.
	if !errors.Is(permA, permA) {
		t.Errorf("errors.Is(permA, permA): identity must match itself")
	}

	// A *PermanentError target that lives in the wrapped chain still matches
	// via errors.Is's identity check (errors.Is walks Unwrap and reaches permA
	// itself in the chain).
	wrappedA := fmt.Errorf("wrap: %w", permA)
	if !errors.Is(wrappedA, permA) {
		t.Errorf("errors.Is(wrapped(permA), permA): permA must be reachable via Unwrap")
	}

	// But the unrelated permB is not reachable from permA's chain, so it
	// must not match wrapped(permA).
	if errors.Is(wrappedA, permB) {
		t.Errorf("errors.Is(wrapped(permA), permB): permB is not in wrappedA's chain")
	}
}
