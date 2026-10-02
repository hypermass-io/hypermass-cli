package subscribe

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/app_constants"
	"hypermass-cli/app_errors"
	"hypermass-cli/config"
	"io"
	"net/http"
	"strconv"
	"time"
)

type AuthResponse struct {
	ConnectionURL string `json:"connectionUrl"`
}

func GetAuthorizedSubscriptionUrl(auth config.HypermassAuth, streamId string, lastPayloadId string) (string, error) {

	authenticationUrl := app_constants.GetBulkAuthenticationApiUrl(streamId, lastPayloadId)

	// Create the request
	req, err := http.NewRequest("GET", authenticationUrl, nil)
	if err != nil {
		return "", &app_errors.AuthenticationFailedError{
			Message: fmt.Sprintf("could not build the authentication request, please report this to support: %v", err),
		}
	}

	// Add the Authorization header
	auth.Authorize(req)
	req.Header.Set("User-Agent", "hypermass-cli/"+app_constants.HypermassCliVersion)

	// Send the request
	client := &http.Client{} // follows redirects by default
	resp, err := client.Do(req)
	if err != nil {
		return "", &app_errors.ConnectionLostError{Message: fmt.Sprintf("could not reach the authentication service: %v", err)}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body) // response body is []byte

	if resp.StatusCode != 200 {
		return "", subscriptionRefusal(resp)
	}

	var result AuthResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		return "", &app_errors.ConnectionLostError{Message: "the authentication service sent an unreadable response"}
	}

	location := result.ConnectionURL

	return location, err
}

// subscriptionRefusal turns a refusal into an error that this subscription can "back off" on.
//
// The status says what it is: the allowance is spent, the key is wrong, the feed is out of reach, or
// there is no such stream. Every status returns an error, including one this client does not recognise.
func subscriptionRefusal(resp *http.Response) error {
	retryAfter := retryAfterFrom(resp)

	switch resp.StatusCode {
	case http.StatusPaymentRequired:
		return &app_errors.InsufficientAllowanceError{
			Message: "insufficient allowance to subscribe to this feed",
			Advised: retryAfter,
		}

	case http.StatusUnauthorized:
		return &app_errors.CredentialsRejectedError{
			Message: "this key was rejected",
			Advised: retryAfter,
		}

	case http.StatusForbidden:
		return &app_errors.StreamAccessDeniedError{
			Message: "this feed is not available to this key",
			Advised: retryAfter,
		}

	case http.StatusNotFound:
		return &app_errors.StreamNotFoundError{
			Message: "there is no stream with this id",
			Advised: retryAfter,
		}

	case http.StatusTooManyRequests:
		return &app_errors.RetryLaterError{
			Message:            "asking for this feed too often",
			RetryAfterDuration: retryAfter,
		}

	default:
		return &app_errors.ConnectionLostError{
			Message: "the service refused the subscription",
			Advised: retryAfter,
		}
	}
}

// retryAfterFrom reads the server's advice on when to retry (if provided).
func retryAfterFrom(resp *http.Response) time.Duration {
	header := resp.Header.Get("Retry-After")
	if header == "" {
		return 0
	}

	seconds, err := strconv.Atoi(header)
	if err != nil || seconds <= 0 {
		return 0
	}

	return time.Duration(seconds) * time.Second
}
