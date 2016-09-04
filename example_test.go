package backoff

import "log"

func ExampleRetry() {
	// An operation that might fail.
	operation := func() error {
		return nil // or an error
	}

	err := Retry(operation, NewExponentialBackOff())
	if err != nil {
		// Handle error.
		return
	}

	// Operation is successful.
}

func ExampleTicker() {
	// An operation that might fail
	operation := func() error {
		return nil // or an error
	}

	b := NewExponentialBackOff()
	ticker := NewTicker(b)

	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for _ = range ticker.C {
		if err = operation(); err != nil {
			log.Println(err, "will retry...")
			continue
		}

		ticker.Stop()
		break
	}

	if err != nil {
		// Operation has failed.
		return
	}

	// Operation is successful.
	return
}
