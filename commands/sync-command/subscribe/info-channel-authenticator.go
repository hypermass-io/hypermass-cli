package subscribe

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/app_constants"
	"hypermass-cli/app_errors"
	"hypermass-cli/config"
	"io"
	"log"
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
		log.Println(err)
		log.Println("Failed to authenticate, unable to construct auth request. Please report this message to support")
		return "", &app_errors.AuthenticationFailedError{Message: "could not build the authentication request"}
	}

	// Add the Authorization header
	req.Header.Set("Authorization", "Bearer "+auth.Token)
	req.Header.Set("User-Agent", "hypermass-cli/"+app_constants.HypermassCliVersion)

	// Send the request
	client := &http.Client{} // follows redirects by default
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			log.Printf("authentication failed while connecting to service: status=%d err=%v", resp.StatusCode, err)
		} else {
			log.Printf("authentication failed while connecting to service: err=%v", err)
		}
		return "", &app_errors.ConnectionLostError{Message: "could not reach the authentication service"}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body) // response body is []byte

	if resp.StatusCode != 200 {
		log.Printf("authentication failed while connecting to service: status=%d message=%s", resp.StatusCode, body)
		return "", subscriptionRefusal(resp, body)
	}

	var result AuthResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	location := result.ConnectionURL

	return location, err
}

// subscriptionRefusal turns a refusal into an error that this subscription can "back off" on.
//
// Every status returns an error, including one this client does not recognise.
func subscriptionRefusal(resp *http.Response, body []byte) error {
	retryAfter := retryAfterFrom(resp)

	switch resp.StatusCode {
	case http.StatusPaymentRequired:
		//402 covers an exhausted allowance and an unavailable stream, differentiated here
		switch refusedBecause(body) {
		case "SUSPENDED":
			return &app_errors.StreamUnavailableError{
				Message: "this feed is not currently available",
				Advised: retryAfter,
			}
		case "API_KEY_STREAM_DOES_NOT_EXIST":
			return &app_errors.StreamNotFoundError{
				Message: "there is no stream with this id",
				Advised: retryAfter,
			}
		}

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
			Message: "this key may not subscribe to this feed",
			Advised: retryAfter,
		}

	case http.StatusNotFound:
		return &app_errors.StreamUnavailableError{
			Message: "this feed is not currently available",
			Advised: retryAfter,
		}

	default:
		return &app_errors.ConnectionLostError{
			Message: "the service refused the subscription",
			Advised: retryAfter,
		}
	}
}

// refusedBecause reads the reason from a refusal body if present (or an empty string)
func refusedBecause(body []byte) string {
	var refusal struct {
		RefusedBecause string `json:"refusedBecause"`
	}

	if err := json.Unmarshal(body, &refusal); err != nil {
		return ""
	}

	return refusal.RefusedBecause
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
