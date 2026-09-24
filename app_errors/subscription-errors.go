package app_errors

import "time"

// How long to wait before retrying, by kind of failure.
//
// Each delay belongs to the error that uses it. The error already describes what went wrong, and how
// long to wait follows from that, so keeping both together leaves one place to change.
const (
	// AllowanceRetryDelay applies when an allowance is exhausted. Allowances change when a month rolls
	// over or a plan changes, so checking every few minutes is often enough.
	AllowanceRetryDelay = 5 * time.Minute

	// CredentialsRejectedRetryDelay applies when the credentials are refused. Someone has to replace
	// the key or unlock the account.
	CredentialsRejectedRetryDelay = 30 * time.Minute

	// StreamAccessDeniedRetryDelay applies when the credentials do not cover a stream. Someone has to
	// grant access or correct the configuration.
	StreamAccessDeniedRetryDelay = 30 * time.Minute

	// StreamNotFoundRetryDelay applies when no stream has this id. Someone has to correct the id, or
	// create the stream.
	StreamNotFoundRetryDelay = 30 * time.Minute

	// StreamUnavailableRetryDelay applies when a stream has been suspended or removed. Its owner has to
	// resolve whatever caused that.
	StreamUnavailableRetryDelay = 30 * time.Minute

	// AuthenticationFailedRetryDelay applies when a request could not be authenticated.
	AuthenticationFailedRetryDelay = 30 * time.Minute

	// ConnectionRetryDelay applies when the connection drops. Connections usually come back quickly.
	ConnectionRetryDelay = 10 * time.Second

	// DownloadFailedRetryDelay applies when a payload download fails part way through.
	DownloadFailedRetryDelay = 60 * time.Second

	// RetryLaterRetryDelay applies when the service asked us to wait but named no period.
	RetryLaterRetryDelay = 60 * time.Second

	// DefaultRetryDelay applies to a failure that carries no delay of its own, such as an error from
	// outside this package.
	DefaultRetryDelay = 60 * time.Second
)

// RetryableError is a failure that carries its own retry delay and a description for the user.
//
// Every subscription and publication failure implements it. The retry loop reads the delay from the
// error, and the status command shows the description against the affected stream.
type RetryableError interface {
	error

	// RetryAfter is how long to wait before trying again.
	RetryAfter() time.Duration

	// Summary describes the state in a few words, for the status command.
	Summary() string
}

// InsufficientAllowanceError indicates that the user does not have enough allowance to perform the requested action
type InsufficientAllowanceError struct {
	Message string

	// Advised is the delay the service asked for. Zero when it gave no advice, in which case the
	// default for this kind of failure applies.
	Advised time.Duration
}

func (e *InsufficientAllowanceError) Error() string {
	return e.Message
}

func (e *InsufficientAllowanceError) RetryAfter() time.Duration {
	if e.Advised > 0 {
		return e.Advised
	}

	return AllowanceRetryDelay
}

func (e *InsufficientAllowanceError) Summary() string {
	return "waiting on allowance"
}

// ConnectionLostError indicates that the connection to the server was lost
type ConnectionLostError struct {
	Message string
	Advised time.Duration
}

func (e *ConnectionLostError) Error() string {
	return e.Message
}

func (e *ConnectionLostError) RetryAfter() time.Duration {
	if e.Advised > 0 {
		return e.Advised
	}

	return ConnectionRetryDelay
}

func (e *ConnectionLostError) Summary() string {
	return "reconnecting"
}

// DownloadFailedError indicates that the download failed
type DownloadFailedError struct {
	Message string
}

func (e *DownloadFailedError) Error() string {
	return e.Message
}

func (e *DownloadFailedError) RetryAfter() time.Duration {
	return DownloadFailedRetryDelay
}

func (e *DownloadFailedError) Summary() string {
	return "a download failed"
}

// CredentialsRejectedError indicates the credentials were refused. The key may be wrong or revoked, or
// the account may be locked. It applies to every stream, so one call receiving it means all of them
// would.
//
// The sync command stops on this failure at startup, because no change to the configuration can make
// the credentials work.
type CredentialsRejectedError struct {
	Message string
	Advised time.Duration
}

func (e *CredentialsRejectedError) Error() string {
	return e.Message
}

func (e *CredentialsRejectedError) RetryAfter() time.Duration {
	if e.Advised > 0 {
		return e.Advised
	}

	return CredentialsRejectedRetryDelay
}

func (e *CredentialsRejectedError) Summary() string {
	return "key rejected or account locked"
}

// StreamAccessDeniedError indicates the credentials were accepted but do not cover this stream, which
// usually means it belongs to another account. It applies to one stream, so the rest of the sync
// continues.
type StreamAccessDeniedError struct {
	Message string
	Advised time.Duration
}

func (e *StreamAccessDeniedError) Error() string {
	return e.Message
}

func (e *StreamAccessDeniedError) RetryAfter() time.Duration {
	if e.Advised > 0 {
		return e.Advised
	}

	return StreamAccessDeniedRetryDelay
}

func (e *StreamAccessDeniedError) Summary() string {
	return "no access to this stream"
}

// AuthenticationFailedError indicates that the user's auth key was rejected
type AuthenticationFailedError struct {
	Message string
	Advised time.Duration
}

func (e *AuthenticationFailedError) Error() string {
	return e.Message
}

func (e *AuthenticationFailedError) RetryAfter() time.Duration {
	if e.Advised > 0 {
		return e.Advised
	}

	return AuthenticationFailedRetryDelay
}

func (e *AuthenticationFailedError) Summary() string {
	return "key rejected"
}

// StreamNotFoundError indicates there is no stream with this id, which usually means a typo in the
// configuration. It has its own type so the status command can ask someone to correct the id, which is
// the only thing that will resolve it.
type StreamNotFoundError struct {
	Message string
	Advised time.Duration
}

func (e *StreamNotFoundError) Error() string {
	return e.Message
}

func (e *StreamNotFoundError) RetryAfter() time.Duration {
	if e.Advised > 0 {
		return e.Advised
	}

	return StreamNotFoundRetryDelay
}

func (e *StreamNotFoundError) Summary() string {
	return "no such stream - check the id"
}

// StreamUnavailableError indicates the stream exists but the service will not serve it, because it has
// been suspended or removed. The stream may become available again once its owner resolves the
// problem.
type StreamUnavailableError struct {
	Message string
	Advised time.Duration
}

func (e *StreamUnavailableError) Error() string {
	return e.Message
}

func (e *StreamUnavailableError) RetryAfter() time.Duration {
	if e.Advised > 0 {
		return e.Advised
	}

	return StreamUnavailableRetryDelay
}

func (e *StreamUnavailableError) Summary() string {
	return "stream unavailable"
}

// RetryLaterError indicates that the request was valid but should be retried after a delay
type RetryLaterError struct {
	Message            string
	RetryAfterDuration time.Duration
}

func (e *RetryLaterError) Error() string {
	if e.RetryAfterDuration > 0 {
		return e.Message + " (retry after " + e.RetryAfterDuration.String() + ")"
	}
	return e.Message
}

func (e *RetryLaterError) RetryAfter() time.Duration {
	if e.RetryAfterDuration > 0 {
		return e.RetryAfterDuration
	}

	return RetryLaterRetryDelay
}

func (e *RetryLaterError) Summary() string {
	return "waiting as asked"
}
