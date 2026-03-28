package backoff_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/cenkalti/backoff/v5"
)

func ExampleRetry() {
	// Define an operation function that returns a value and an error.
	// The value can be any type.
	// We'll pass this operation to Retry function.
	operation := func() (string, error) {
		// An example request that may fail.
		resp, err := http.Get("http://httpbin.org/get")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		// If we are being rate limited, return a RetryAfter to specify how long to wait.
		// This will also reset the backoff policy.
		if resp.StatusCode == 429 {
			seconds, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
			if err == nil {
				return "", backoff.RetryAfter(int(seconds))
			}
		}

		// In case of non-retriable error, return Permanent error to stop retrying.
		// For this HTTP example, client errors are non-retriable.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return "", backoff.Permanent(errors.New("bad request"))
		}

		// Return successful response.
		return "hello", nil
	}

	result, err := backoff.Retry(context.TODO(), operation, backoff.WithBackOff(backoff.NewExponentialBackOff()))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Operation is successful after retries.

	fmt.Println(result)
	// Output: hello
}

func ExampleTicker() {
	// An operation that may fail.
	operation := func() (string, error) {
		return "hello", nil
	}

	ticker := backoff.NewTicker(backoff.NewExponentialBackOff())
	defer ticker.Stop()

	var result string
	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if result, err = operation(); err != nil {
			log.Println(err, "will retry...")
			continue
		}

		break
	}

	if err != nil {
		// Operation has failed.
		fmt.Println("Error:", err)
		return
	}

	// Operation is successful after retries.

	fmt.Println(result)
	// Output: hello
}
