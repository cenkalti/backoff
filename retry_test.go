package backoff

import (
	"errors"
	"log"
	"testing"

	"golang.org/x/net/context"
)

func TestRetry(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return nil
		}

		log.Println("error")
		return errors.New("error")
	}

	err := Retry(f, NewExponentialBackOff())
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryWithContext(t *testing.T) {
	f := func() error {
		t.Error("This function shouldn't be called at all")
		return errors.New("error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RetryNotifyWithContext(ctx, f, NewExponentialBackOff(), nil)
	if err != ctx.Err() {
		t.Errorf("unexpected error: %v", err)
	}
}
