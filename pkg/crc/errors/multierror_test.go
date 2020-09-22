package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryAfter(t *testing.T) {
	calls := 0
	ret := RetryAfter(10, func() error {
		calls++
		return nil
	}, 0)
	assert.NoError(t, ret)
	assert.Equal(t, 1, calls)
}

func TestRetryAfterFailure(t *testing.T) {
	calls := 0
	ret := RetryAfter(10, func() error {
		calls++
		return errors.New("failed")
	}, 0)
	assert.EqualError(t, ret, "failed")
	assert.Equal(t, 1, calls)
}

func TestRetryAfterMaxAttempts(t *testing.T) {
	calls := 0
	ret := RetryAfter(3, func() error {
		calls++
		return &RetriableError{Err: errors.New("failed")}
	}, 0)
	assert.EqualError(t, ret, "Temporary error: failed (x3)")
	assert.Equal(t, 3, calls)
}

func TestRetryAfterSuccessAfterFailures(t *testing.T) {
	calls := 0
	ret := RetryAfter(5, func() error {
		calls++
		if calls < 3 {
			return &RetriableError{Err: errors.New("failed")}
		}
		return nil
	}, 0)
	assert.NoError(t, ret)
	assert.Equal(t, 3, calls)
}

func TestMultiErrorString(t *testing.T) {
	var multiError MultiError
	for i := 0; i < 4; i++ {
		multiError.Collect(errors.New("Temporary Error: No Pending CSR"))
	}
	assert.Equal(t, "Temporary Error: No Pending CSR (x4)", multiError.Error())

	multiError = MultiError{}
	multiError.Collect(errors.New("No Pending CSR"))
	multiError.Collect(errors.New("No Pending CSR"))
	multiError.Collect(errors.New("Connection refused"))
	multiError.Collect(errors.New("Connection refused"))
	multiError.Collect(errors.New("No Pending CSR"))
	assert.Equal(t, "No Pending CSR (x2)\nConnection refused (x2)\nNo Pending CSR", multiError.Error())
}
