package errors

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type MultiError struct {
	Errors []error
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
	if len(m.Errors) != 0 {
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

func (m MultiError) Error() string {
	if len(m.Errors) == 0 {
		return ""
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	var aggregatedErrors []error

	count := 1
	current := m.Errors[0]
	for i := 1; i < len(m.Errors); i++ {
		if equalErr(m.Errors[i], current) {
			count++
			continue
		}
		aggregatedErrors = append(aggregatedErrors, &errorWithCount{errorCount: count, err: current})
		count = 1
		current = m.Errors[i]
	}
	aggregatedErrors = append(aggregatedErrors, &errorWithCount{errorCount: count, err: current})

	var strBuilder strings.Builder
	for _, err := range aggregatedErrors {
		strBuilder.WriteString(err.Error())
		strBuilder.WriteString("\n")
	}
	return strBuilder.String()
}

func (m *MultiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
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
			return m
		}
		logging.Debugf("error: %v - sleeping %s", err, d)
		time.Sleep(d)
	}
	logging.Debugf("RetryAfter timeout after %d tries", attempts)
	return m
}
