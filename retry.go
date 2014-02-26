package backoff

import "time"

// Retry the function f until it does not return error or BackOff stops.
//
// Example:
// 	operation := func() error {
// 		// An operation that may fail
// 	}
//
// 	err := Retry(operation, NewExponentialBackoff())
// 	if err != nil {
// 		// handle error
// 	}
//
// 	// operation is successfull
func Retry(f func() error, b BackOff) error {
	err := f()
	if err == nil {
		return nil
	}

	b.Reset()
	for {
		next := b.NextBackOff()
		if next == Stop {
			return err
		}

		time.Sleep(time.Duration(next))
		err = f()
		if err != nil {
			continue
		}

		return nil
	}
}
