package publication_helpers

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/app_constants"
	"hypermass-cli/app_errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Define constants for the timeout and default retry value
const (
	// UPLOAD_TIMEOUT the timeout for uploading a single payload file
	UPLOAD_TIMEOUT = 10 * time.Minute
	// DEFAULT_RETRY_AFTER a fallback value for the retry interval (in seconds)
	DEFAULT_RETRY_AFTER = 10 * 60
)

// UploadResponse maps the expected fields from the successful API response body.
type UploadResponse struct {
	PayloadId string `json:"payloadId"`
}

// UploadResult defines the possible outcomes.
type UploadResult struct {
	PayloadId string // Only relevant for "ok"
}

// PublishFileToStream publishes a payload to the specified stream
func PublishFileToStream(filepath string, streamId string, token string) (*UploadResult, error) {
	//get the signed URL
	signedURL, signedUrlErr := getSignedUploadURL(token, streamId)
	if signedUrlErr != nil {
		return nil, signedUrlErr
	}

	//start writing the file to a pipe
	pr, writer, err := writeFileToPipe(filepath)
	if err != nil {
		return nil, err
	}

	//perform the upload and return the result
	return upload(signedURL, pr, writer)
}

// upload send the file to the server
func upload(signedURL string, pr *io.PipeReader, writer *multipart.Writer) (*UploadResult, error) {

	client := http.Client{
		Timeout: UPLOAD_TIMEOUT,
	}

	req, err := http.NewRequest("POST", signedURL, pr)
	if err != nil {
		pr.Close()
		return nil, fmt.Errorf("failed to upload payload to signed url: %s", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	//the actual upload
	resp, err := client.Do(req)

	if err != nil {
		if os.IsTimeout(err) {
			return nil, fmt.Errorf("failed to upload payload, timeout after %s", UPLOAD_TIMEOUT)
		} else {
			return nil, fmt.Errorf("failed to upload payload to signed url: %s", err)
		}
	}
	defer resp.Body.Close()

	//Success path
	if resp.StatusCode == http.StatusOK {
		var uploadResp UploadResponse

		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			fmt.Printf("upload succeeded but failed to parse JSON response: %w", err)
			return &UploadResult{PayloadId: ""}, nil
		}

		return &UploadResult{PayloadId: uploadResp.PayloadId}, nil
	}

	//Failed path
	//TODO HYP-275 - scaling efforts will make this redundant (all upload checks will be moved to the signed url)
	if resp.StatusCode == http.StatusServiceUnavailable {
		return nil, buildRetryLaterError(resp)
	}

	//TODO HYP-275 - scaling efforts will make this redundant (all upload checks will be moved to the signed url)
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, &app_errors.InsufficientAllowanceError{Message: "insufficient allowance to publish to this feed"}
	} else {
		// All other errors
		return nil, fmt.Errorf("Unable to upload payload, response code: %d\n", resp.StatusCode)
	}
}

// getSignedUploadURL performs a preliminary GET request to obtain a signed upload URL.
// It expects a 200 response with the 'Location' header set.
func getSignedUploadURL(token string, streamId string) (string, error) {
	authURL := app_constants.PublicApiUrl + "/data/bulk/authorise/uploadReferral/" + streamId
	client := http.Client{Timeout: 30 * time.Second} // Use a shorter timeout for this quick auth request

	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating auth request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error performing auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("preliminary auth failed with status %d. Check token", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusServiceUnavailable {
		return "", buildRetryLaterError(resp)
	}

	if resp.StatusCode == http.StatusPaymentRequired {
		return "", &app_errors.InsufficientAllowanceError{Message: "insufficient allowance to publish to this feed"}
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("preliminary auth failed with unexpected status code: %d", resp.StatusCode)
	}

	signedURL := resp.Header.Get("Location")
	if signedURL == "" {
		return "", fmt.Errorf("preliminary auth successful (200), but 'Location' header is missing")
	}

	return signedURL, nil
}

func buildRetryLaterError(resp *http.Response) error {
	retryAfterHeader := resp.Header.Get("Retry-After")
	retryAfterSec := DEFAULT_RETRY_AFTER

	if retryAfterHeader != "" {
		val, parseErr := strconv.Atoi(retryAfterHeader)
		if parseErr == nil {
			retryAfterSec = val
		} else {
			fmt.Printf("Could not parse Retry-After header '%s' as integer. Using default.\n", retryAfterHeader)
		}
	} else {
		fmt.Printf("Retry-After header missing. Using default value: %d seconds.\n", DEFAULT_RETRY_AFTER)
	}

	//sanity check in all cases, prevent busy loops hammering the server
	if retryAfterSec < 5 {
		retryAfterSec = 5
	}

	return &app_errors.RetryLaterError{
		RetryAfter: time.Duration(retryAfterSec) * time.Second,
	}
}

// writeFileToPipe creates a pipe reader to which the file is written
func writeFileToPipe(filepath string) (*io.PipeReader, *multipart.Writer, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw) // The multipart writer directs output to the PipeWriter

	go func() {
		defer pw.Close()   // Close the PipeWriter when we are done or hit an error
		defer file.Close() // Close the File       when we are done or hit an error

		part, err := writer.CreateFormFile("file", filepath)
		if err != nil {
			fmt.Printf("Error creating form file: %v\n", err)
			return
		}

		// Copy the file contents into the form part (which writes through the pipe)
		_, err = io.Copy(part, file)
		if err != nil {
			fmt.Printf("Error copying file contents: %v\n", err)
			return
		}

		if err := writer.Close(); err != nil {
			fmt.Printf("Error closing multipart writer: %v\n", err)
		}
	}()

	return pr, writer, nil
}
