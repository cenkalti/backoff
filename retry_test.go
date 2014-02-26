package backoff

import (
	"errors"
	"testing"
)

func TestRetry(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successfull after "successOn" calls.
	f := func() error {
		i++
		t.Logf("function is called %d. time", i)

		if i == successOn {
			t.Log("OK")
			return nil
		}

		t.Log("error")
		return errors.New("error")
	}

	err := Retry(f, NewExponentialBackoff())
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}
