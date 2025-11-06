package agent

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ErrorType int

const (
	ErrorTypeTemporary ErrorType = iota
	ErrorTypeCritical
	ErrorTypeRetryable
)

func (e ErrorType) String() string {
	switch e {
	case ErrorTypeTemporary:
		return "temporary"
	case ErrorTypeCritical:
		return "critical"
	case ErrorTypeRetryable:
		return "retryable"
	default:
		return "unknown"
	}
}

type ActionError struct {
	Type    ErrorType
	Action  string
	Message string
	Err     error
}

func (e *ActionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Action, e.Message)
}

func (e *ActionError) Unwrap() error {
	return e.Err
}

func classifyError(action string, err error) *ActionError {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	msg := err.Error()

	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "ECONNREFUSED") ||
		strings.Contains(errStr, "ETIMEDOUT") {
		return &ActionError{
			Type:    ErrorTypeRetryable,
			Action:  action,
			Message: msg,
			Err:     err,
		}
	}

	if strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "selector") ||
		strings.Contains(errStr, "element") {
		return &ActionError{
			Type:    ErrorTypeTemporary,
			Action:  action,
			Message: msg,
			Err:     err,
		}
	}

	return &ActionError{
		Type:    ErrorTypeCritical,
		Action:  action,
		Message: msg,
		Err:     err,
	}
}

func retryAction(ctx context.Context, maxRetries int, delay time.Duration, fn func() error) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		actionErr := classifyError("", err)
		if actionErr.Type == ErrorTypeCritical {
			return err
		}
	}

	return fmt.Errorf("после %d попыток: %w", maxRetries, lastErr)
}

func isCriticalError(err error) bool {
	actionErr := classifyError("", err)
	return actionErr.Type == ErrorTypeCritical
}
