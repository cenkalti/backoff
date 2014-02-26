package backoff

import "time"

// Retry takes a function and a BackOff implementation and retries the function
// until it does not return error or BackOff stops.
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
