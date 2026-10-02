package subscribe

import (
	"errors"
	"hypermass-cli/app_errors"
	"net/http"
	"testing"
	"time"
)

func responseWith(status int, retryAfter string) *http.Response {
	response := &http.Response{StatusCode: status, Header: http.Header{}}

	if retryAfter != "" {
		response.Header.Set("Retry-After", retryAfter)
	}

	return response
}

// Every refusal must produce an error the caller can back off on. Returning nil would leave the caller
// with no reason to retry, and exiting would stop the other subscriptions in the same sync.
func TestSubscriptionRefusalAlwaysReturnsAnError(t *testing.T) {
	for _, status := range []int{400, 401, 402, 403, 404, 418, 429, 500, 503} {
		if subscriptionRefusal(responseWith(status, "")) == nil {
			t.Errorf("status %d produced no error", status)
		}
	}
}

// The description and the delay come from the same error, so the user is told this is an allowance
// problem and the wait is the one the service asked for.
func TestPaymentRequiredKeepsItsReasonAndTakesTheServersDelay(t *testing.T) {
	err := subscriptionRefusal(responseWith(http.StatusPaymentRequired, "21600"))

	var allowance *app_errors.InsufficientAllowanceError
	if !errors.As(err, &allowance) {
		t.Fatalf("expected an InsufficientAllowanceError, got %T", err)
	}

	if allowance.RetryAfter() != 6*time.Hour {
		t.Errorf("expected 6h, got %v", allowance.RetryAfter())
	}
}

// Without the header, the error supplies its own default delay.
func TestPaymentRequiredFallsBackToItsOwnDelay(t *testing.T) {
	err := subscriptionRefusal(responseWith(http.StatusPaymentRequired, ""))

	var allowance *app_errors.InsufficientAllowanceError
	if !errors.As(err, &allowance) {
		t.Fatalf("expected an InsufficientAllowanceError, got %T", err)
	}

	if allowance.RetryAfter() != app_errors.AllowanceRetryDelay {
		t.Errorf("expected the allowance default, got %v", allowance.RetryAfter())
	}
}

// Every refusal must supply both a retry delay and a description for the status command.
func TestEveryRefusalIsRetryableAndDescribes(t *testing.T) {
	for _, status := range []int{400, 401, 402, 403, 404, 410, 418, 429, 500, 503} {
		err := subscriptionRefusal(responseWith(status, ""))

		var retryable app_errors.RetryableError
		if !errors.As(err, &retryable) {
			t.Errorf("status %d: %T does not carry retry advice", status, err)
			continue
		}

		if retryable.RetryAfter() <= 0 {
			t.Errorf("status %d: no retry delay", status)
		}

		if retryable.Summary() == "" {
			t.Errorf("status %d: nothing to show the user", status)
		}
	}
}

// Asking too often clears by waiting, where the other refusals need something to change first.
func TestTooManyRequestsIsAWait(t *testing.T) {
	err := subscriptionRefusal(responseWith(http.StatusTooManyRequests, "120"))

	var retryLater *app_errors.RetryLaterError
	if !errors.As(err, &retryLater) {
		t.Fatalf("expected a RetryLaterError, got %T", err)
	}

	if retryLater.RetryAfter() != 2*time.Minute {
		t.Errorf("expected 2m, got %v", retryLater.RetryAfter())
	}
}

// A typo in the configuration needs correcting, where a feed out of reach needs waiting out. The
// status command shows them differently, so they must arrive as different errors.
func TestMissingStreamIsItsOwnState(t *testing.T) {
	err := subscriptionRefusal(responseWith(http.StatusNotFound, ""))

	var notFound *app_errors.StreamNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected a StreamNotFoundError, got %T", err)
	}
}

// Refused credentials stop the command, while a stream this key cannot reach leaves the rest of the
// sync running. The two need separate error types to produce those separate responses.
func TestRejectedKeyIsToldApartFromADeniedStream(t *testing.T) {
	var credentialsRejected *app_errors.CredentialsRejectedError
	if err := subscriptionRefusal(responseWith(http.StatusUnauthorized, "")); !errors.As(err, &credentialsRejected) {
		t.Errorf("401: expected a CredentialsRejectedError, got %T", err)
	}

	var accessDenied *app_errors.StreamAccessDeniedError
	if err := subscriptionRefusal(responseWith(http.StatusForbidden, "")); !errors.As(err, &accessDenied) {
		t.Errorf("403: expected a StreamAccessDeniedError, got %T", err)
	}
}

// An unrecognised status leads to a retry. This is how a client keeps working when the service starts
// sending a status added after that client was released.
func TestUnknownStatusIsRetryable(t *testing.T) {
	err := subscriptionRefusal(responseWith(http.StatusTeapot, ""))

	var connectionLost *app_errors.ConnectionLostError
	if !errors.As(err, &connectionLost) {
		t.Fatalf("expected a ConnectionLostError, got %T", err)
	}
}

func TestRetryAfterIgnoresWhatItCannotRead(t *testing.T) {
	for _, header := range []string{"", "soon", "-5", "0", "Wed, 21 Oct 2026 07:28:00 GMT"} {
		if got := retryAfterFrom(responseWith(http.StatusPaymentRequired, header)); got != 0 {
			t.Errorf("header %q: expected no advice, got %v", header, got)
		}
	}
}

// The service refuses a payload to resume after that the stream never had, so the caller can decide what to do instead.
func TestUnprocessableIsAnUnknownAnchor(t *testing.T) {
	err := subscriptionRefusal(responseWith(http.StatusUnprocessableEntity, ""))

	var unknownAnchor *app_errors.UnknownAnchorError
	if !errors.As(err, &unknownAnchor) {
		t.Fatalf("expected an unknown anchor, got %T", err)
	}
	if unknownAnchor.RetryAfter() != 0 {
		t.Errorf("expected to resume straight away, got %v", unknownAnchor.RetryAfter())
	}
}
