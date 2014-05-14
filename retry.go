package backoff

import "time"

// Retry the function f until it does not return error or BackOff stops.
//
// Example:
// 	operation := func() error {
// 		// An operation that may fail
// 	}
//
// 	err := backoff.Retry(operation, backoff.NewExponentialBackoff())
// 	if err != nil {
// 		// handle error
// 	}
//
// 	// operation is successfull
func Retry(f func() error, b BackOff) error {
	var err error
	var next time.Duration

	b.Reset()
	for {
		if err = f(); err == nil {
			return nil
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}

		time.Sleep(next)
	}
}
