package backoff

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
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

func ExampleRetryContext() { // nolint: govet
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

func ExampleTicker() {
	// An operation that may fail.
	operation := func() error {
		return nil // or an error
	}

	ticker := NewTicker(NewExponentialBackOff())

	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
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
}

func ExampleRetryWithData() {
	b := NewConstantBackOff(time.Microsecond * 100)
	WithMaxRetries(b, 1000)

	// loop through to retry until success or max retries
	data, err := RetryWithData(func() (string, error) {
		n := rand.Intn(100)
		fmt.Println(n)
		if n == 99 {
			return "bingo", nil
		} else {
			return "", errors.New("not bingo")
		}
	}, b)

	if err != nil {
		panic(err)
	}

	fmt.Println(data)
}
