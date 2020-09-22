package errors

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type MultiError struct {
	errors []*errorWithCount
}

type errorWithCount struct {
	errorCount int
	err        error
}

func (err *errorWithCount) Error() string {
	if err.errorCount == 1 {
		return err.err.Error()
	}
	return fmt.Sprintf("%s (x%d)", err.err.Error(), err.errorCount)
}

func (m *MultiError) ErrorOrNil() error {
	if len(m.errors) != 0 {
		return m
	}

	return nil
}

func equalErr(err1, err2 error) bool {
	if reflect.TypeOf(err1) != reflect.TypeOf(err2) {
		return false
	}
	return err1.Error() == err2.Error()
}

func (m *MultiError) Error() string {
	if len(m.errors) == 0 {
		return ""
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(m.errors[0].Error())
	for _, err := range m.errors[1:] {
		strBuilder.WriteString("\n")
		strBuilder.WriteString(err.Error())
	}
	return strBuilder.String()
}

func (m *MultiError) Collect(err error) {
	var lastErr *errorWithCount
	if len(m.errors) == 0 {
		/* Dummy 'errorWithCount' which will not be equal to any real error */
		lastErr = &errorWithCount{}
	} else {
		lastErr = m.errors[len(m.errors)-1]
	}

	if equalErr(err, lastErr.err) {
		lastErr.errorCount++
	} else {
		m.errors = append(m.errors, &errorWithCount{errorCount: 1, err: err})
	}
}

// RetriableError is an error that can be tried again
type RetriableError struct {
	Err error
}

func (r RetriableError) Error() string {
	return "Temporary error: " + r.Err.Error()
}

// RetryAfter retries a number of attempts, after a delay
func RetryAfter(attempts int, callback func() error, d time.Duration) error {
	m := MultiError{}
	for i := 0; i < attempts; i++ {
		if i > 0 {
			logging.Debugf("retry loop: attempt %d out of %d", i, attempts)
		}
		err := callback()
		if err == nil {
			return nil
		}
		m.Collect(err)
		if _, ok := err.(*RetriableError); !ok {
			logging.Debugf("non-retriable error: %v", err)
			return &m
		}
		logging.Debugf("error: %v - sleeping %s", err, d)
		time.Sleep(d)
	}
	logging.Debugf("RetryAfter timeout after %d tries", attempts)
	return &m
}
