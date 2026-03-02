package app_errors

import "time"

type InsufficientAllowanceError struct {
	Message string
}

func (e *InsufficientAllowanceError) Error() string {
	return e.Message
}

type RetryLaterError struct {
	Message    string
	RetryAfter time.Duration
}

func (e *RetryLaterError) Error() string {
	return e.Message
}
