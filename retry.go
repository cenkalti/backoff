package backoff

import "time"

// An Operation is executing by Retry() or RetryNotify().
// The operation will be retried using a backoff policy if it returns an error.
type Operation func() error

// Notify is a notify-on-error function. It receives an operation error and
// backoff delay if the operation failed (with an error).
//
// NOTE that if the backoff policy stated to stop retrying,
// the notify function isn't called.
type Notify func(error, time.Duration)

// Retry the operation o until it does not return error or BackOff stops.
// o is guaranteed to be run at least once.
//
// If o returns a *PermanentError, the operation is not retried, and the
// wrapped error is returned.
//
// Retry sleeps the goroutine for the duration returned by BackOff after a
// failed operation returns.
func Retry(o Operation, b BackOff) error {
	return RetryNotify(o, b, nil)
}

// RetryNotify calls notify function with the error and wait duration
// for each failed attempt before sleep.
func RetryNotify(operation Operation, b BackOff, notify Notify) error {
	return RetryNotifyWithTimer(operation, b, notify, nil)
}

// RetryNotifyWithTimer calls notify function with the error and wait duration using the given Timer
// for each failed attempt before sleep.
func RetryNotifyWithTimer(operation Operation, b BackOff, notify Notify, t Timer) error {
	var err error
	var next time.Duration
	if t == nil {
		t = &DefaultTimer{}
	}

	defer func() {
		t.Stop()
	}()

	ctx := getContext(b)

	b.Reset()
	for {
		if err = operation(); err == nil {
			return nil
		}

		if permanent, ok := err.(*PermanentError); ok {
			return permanent.Err
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}

		if notify != nil {
			notify(err, next)
		}

		t.Start(next)

		select {
		case <-ctx.Done():
			return err
		case <-t.C():
		}
	}
}

type Timer interface {
	Start(duration time.Duration)
	Stop()
	C() <-chan time.Time
}

// DefaultTimer implements Timer interface using time.Timer
type DefaultTimer struct {
	timer *time.Timer
}

// C returns the timers channel which receives the current time when the timer fires.
func (t *DefaultTimer) C() <-chan time.Time {
	return t.timer.C
}

// Start starts the timer to fire after the given duration
func (t *DefaultTimer) Start(duration time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(duration)
	} else {
		t.timer.Reset(duration)
	}
}

// Stop is called when the timer is not used anymore and resources may be freed.
func (t *DefaultTimer) Stop() {
	t.timer.Stop()
}

// PermanentError signals that the operation should not be retried.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// Permanent wraps the given err in a *PermanentError.
func Permanent(err error) *PermanentError {
	return &PermanentError{
		Err: err,
	}
}
