package publication_helpers

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

type BulkStreamConfiguration struct {
	FileExtension string `json:"fileExtension"`
	FileType      string `json:"fileType"`
}

// GetConfigurationForStream retrieve key information about the stream (particularly the type of data)
func GetConfigurationForStream(hypermassProfile config.HypermassProfile, streamId string) (BulkStreamConfiguration, error) {
	url := app_constants.PublicApiUrl + "/data/bulk/id/" + streamId + "/write-configuration"

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		log.Println("Failed to authenticate, unable to construct auth request. Please report this message to support")
		return BulkStreamConfiguration{}, &app_errors.AuthenticationFailedError{
			Message: "could not build the configuration request",
		}
	}

	// Add the Authorization header
	req.Header.Set("Authorization", "Bearer "+hypermassProfile.Auth.Token)
	req.Header.Set("User-Agent", "hypermass-cli/"+app_constants.HypermassCliVersion)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		log.Println("Failed to authenticate, unable to connect to service")
		return BulkStreamConfiguration{}, &app_errors.ConnectionLostError{
			Message: "could not reach the service",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return BulkStreamConfiguration{}, configurationRefusal(resp, streamId)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return BulkStreamConfiguration{}, &app_errors.ConnectionLostError{Message: "could not read the response"}
	}

	var result BulkStreamConfiguration
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
		return BulkStreamConfiguration{}, &app_errors.ConnectionLostError{
			Message: "the service sent a configuration we could not read",
		}
	}

	return result, nil
}

// configurationRefusal turns a refusal into an error the publication loop can wait on.
//
// The service returns 401 both for a bad key and for a locked account, so the message covers both and
// points at the website. Where the service sends a Retry-After, that delay travels on the error and
// sets how long the loop waits.
func configurationRefusal(resp *http.Response, streamId string) error {
	retryAfter := retryAfterFrom(resp)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		//the credentials were refused, so every other stream will be refused the same way
		log.Println("Not authorized to upload to this stream. The key may be wrong or revoked, or the " +
			"account may be locked - sign in at hypermass.io to check")
		return &app_errors.CredentialsRejectedError{
			Message: "this key was rejected",
			Advised: retryAfter,
		}

	case http.StatusForbidden:
		log.Printf("Not authorized to upload to stream %s, it must be owned by your account", streamId)
		return &app_errors.StreamAccessDeniedError{
			Message: "this key may not publish to " + streamId,
			Advised: retryAfter,
		}

	case http.StatusNotFound:
		log.Println("Unable to get stream metadata, stream not found")
		return &app_errors.StreamNotFoundError{
			Message: "there is no stream with this id",
			Advised: retryAfter,
		}

	default:
		return &app_errors.ConnectionLostError{
			Message: "could not read the stream configuration",
			Advised: retryAfter,
		}
	}
}

// retryAfterFrom reads the Retry-After header. It returns zero when the service sent no advice.
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
