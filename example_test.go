package backoff

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

func ExampleRetry() {
	// An operation that may fail.
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

func ExampleRetryContext() {
	// A context
	ctx := context.Background()

	// An operation that may fail.
	operation := func() error {
		return nil // or an error
	}

	b := WithContext(NewExponentialBackOff(), ctx)

	err := Retry(operation, b)
	if err != nil {
		// Handle error.
		return
	}

	// Operation is successful.
}

func ExampleThreadSafe() {
	backoff := NewExponentialBackOff()

	backoff.MaxElapsedTime = time.Millisecond * 500
	backoff.MaxInterval = time.Millisecond * 200

	wg := sync.WaitGroup{}
	wg.Add(2)

	failTimes := 3

	go func() {
		tries := 0

		err := Retry(func() error {
			if tries >= failTimes {
				return nil
			}

			tries++
			return errors.New("FAILED")
		}, backoff)

		if err != nil {
			// Handle error.
		}

		// Operation is successful.
		wg.Done()
	}()

	go func() {
		tries := 0

		err := Retry(func() error {
			if tries >= failTimes {
				return nil
			}

			tries++
			return errors.New("FAILED")
		}, backoff)

		if err != nil {
			// Handle error.
		}

		// Operation is successful.
		wg.Done()
	}()

	wg.Wait()
}

func ExampleTicker() {
	// An operation that may fail.
	operation := func() error {
		return nil // or an error
	}

	ticker := NewTicker(NewExponentialBackOff())

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
