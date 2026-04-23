package app_errors

import "time"

// InsufficientAllowanceError indicates that the user does not have enough allowance to perform the requested action
type InsufficientAllowanceError struct {
	Message string
}

func (e *InsufficientAllowanceError) Error() string {
	return e.Message
}

// DownloadFailedError indicates that the download failed
type DownloadFailedError struct {
	Message string
}

func (e *DownloadFailedError) Error() string {
	return e.Message
}

// AuthenticationFailedError indicates that the user's auth key was rejected
type AuthenticationFailedError struct {
	Message string
}

func (e *AuthenticationFailedError) Error() string {
	return e.Message
}

// RetryLaterError indicates that the request was valid but should be retried after a delay
type RetryLaterError struct {
	Message    string
	RetryAfter time.Duration
}

func (e *RetryLaterError) Error() string {
	if e.RetryAfter > 0 {
		return e.Message + " (retry after " + e.RetryAfter.String() + ")"
	}
	return e.Message
}
