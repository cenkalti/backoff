package backoff

import (
	"context"
	"errors"
	"time"
)

// DefaultMaxElapsedTime sets a default limit for the total retry duration.
const DefaultMaxElapsedTime = 15 * time.Minute

// Operation is a function that attempts an operation and may be retried.
type Operation[T any] func() (T, error)

// Notify is a function called on operation error with the error and backoff duration.
type Notify func(error, time.Duration)

// retryOptions holds configuration settings for the retry mechanism.
type retryOptions struct {
	BackOff        BackOff       // Strategy for calculating backoff periods.
	Timer          timer         // Timer to manage retry delays.
	Notify         Notify        // Optional function to notify on each retry error.
	MaxTries       uint          // Maximum number of retry attempts.
	MaxElapsedTime time.Duration // Maximum total time for all retries.
}

type RetryOption func(*retryOptions)

// WithBackOff configures a custom backoff strategy.
func WithBackOff(b BackOff) RetryOption {
	return func(args *retryOptions) {
		args.BackOff = b
	}
}

// withTimer sets a custom timer for managing delays between retries.
func withTimer(t timer) RetryOption {
	return func(args *retryOptions) {
		args.Timer = t
	}
}

// WithNotify sets a notification function to handle retry errors.
func WithNotify(n Notify) RetryOption {
	return func(args *retryOptions) {
		args.Notify = n
	}
}

// WithMaxTries limits the number of all attempts.
func WithMaxTries(n uint) RetryOption {
	return func(args *retryOptions) {
		args.MaxTries = n
	}
}

// WithMaxElapsedTime limits the total duration for retry attempts.
func WithMaxElapsedTime(d time.Duration) RetryOption {
	return func(args *retryOptions) {
		args.MaxElapsedTime = d
	}
}

// Retry attempts the operation until success, a permanent error, or backoff completion.
// It ensures the operation is executed at least once.
//
// Returns the operation result or error if retries are exhausted, context is cancelled, or the operation returns a permanent error.
func Retry[T any](ctx context.Context, operation Operation[T], opts ...RetryOption) (T, error) {
	var res T
	err := retryInternal(ctx, func() (bool, error) {
		var err error

		// Execute the operation.
		res, err = operation()
		if err == nil {
			return true, nil
		}

		return false, err
	}, opts...)

	return res, err
}

// RetryNoResult attempts the operation until success, a permanent error, or backoff completion.
// It ensures the operation is executed at least once.
//
// Returns an error if retries are exhausted, context is cancelled, or the operation returns a permanent error.
func RetryNoResult(ctx context.Context, operation func() error, opts ...RetryOption) error {
	err := retryInternal(ctx, func() (bool, error) {
		// Execute the operation
		err := operation()
		if err == nil {
			return true, nil
		}

		return false, err
	}, opts...)

	return err
}

func retryInternal(ctx context.Context, innerFunc func() (bool, error), opts ...RetryOption) error {
	// Initialize default retry options.
	args := &retryOptions{
		BackOff:        NewExponentialBackOff(),
		Timer:          &defaultTimer{},
		MaxElapsedTime: DefaultMaxElapsedTime,
	}

	// Apply user-provided options to the default settings.
	for _, opt := range opts {
		opt(args)
	}

	defer args.Timer.Stop()

	startedAt := time.Now()
	args.BackOff.Reset()
	for numTries := uint(1); ; numTries++ {
		success, err := innerFunc()
		if success {
			return nil
		}

		// Stop retrying if maximum tries exceeded.
		if args.MaxTries > 0 && numTries >= args.MaxTries {
			return err
		}

		// Handle permanent errors without retrying.
		var permanent *PermanentError
		if errors.As(err, &permanent) {
			return err
		}

		// Stop retrying if context is cancelled.
		if cerr := context.Cause(ctx); cerr != nil {
			return err
		}

		// Calculate next backoff duration.
		next := args.BackOff.NextBackOff()
		if next == Stop {
			return err
		}

		// Reset backoff if RetryAfterError is encountered.
		var retryAfter *RetryAfterError
		if errors.As(err, &retryAfter) {
			next = retryAfter.Duration
			args.BackOff.Reset()
		}

		// Stop retrying if maximum elapsed time exceeded.
		if args.MaxElapsedTime > 0 && time.Since(startedAt)+next > args.MaxElapsedTime {
			return err
		}

		// Notify on error if a notifier function is provided.
		if args.Notify != nil {
			args.Notify(err, next)
		}

		// Wait for the next backoff period or context cancellation.
		args.Timer.Start(next)
		select {
		case <-args.Timer.C():
		case <-ctx.Done():
			return err
		}
	}
}
