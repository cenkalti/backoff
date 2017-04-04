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
// It is the caller's responsibility to reset b after Retry returns.
//
// If o returns a *FatalError, the operation is not retried, and the
// wrapped error is returned.
//
// Retry sleeps the goroutine for the duration returned by BackOff after a
// failed operation returns.
func Retry(o Operation, b BackOff) error { return RetryNotify(o, b, nil) }

// RetryNotify calls notify function with the error and wait duration
// for each failed attempt before sleep.
func RetryNotify(operation Operation, b BackOff, notify Notify) error {
	var err error
	var next time.Duration

	cb := ensureContext(b)

	b.Reset()
	for {
		if err = operation(); err == nil {
			return nil
		}

		if fatal, ok := err.(*FatalError); ok {
			return fatal.Err
		}

		//For backwards-compatibility
		if permanent, ok := err.(*PermanentError); ok {
			return permanent.Err
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}

		if notify != nil {
			notify(err, next)
		}

		t := time.NewTimer(next)

		select {
		case <-cb.Context().Done():
			t.Stop()
			return err
		case <-t.C:
		}
	}
}

// FatalError signals that the operation should not be retried because the outcome won't change.
type FatalError struct {
	Err error
}

func (e *FatalError) Error() string {
	return e.Err.Error()
}

// Fatal wraps the given err in a *FatalError.
func Fatal(err error) *FatalError {
	return &FatalError{
		Err: err,
	}
}

// Deprecated: Use FatalError instead
// PermanentError signals that the operation should not be retried.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

// Deprecated: Use FatalError instead
// Permanent wraps the given err in a *PermanentError.
func Permanent(err error) *PermanentError {
	return &PermanentError{
		Err: err,
	}
}
