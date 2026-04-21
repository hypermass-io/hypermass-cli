package subscription_helpers

import (
	"fmt"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	"hypermass-cli/config"
	"log"
	"net/http"
	"os"
	"strconv"
)

func DownloadPayload(auth config.HypermassAuth, folderPath string, writer payload_writers.PayloadWriterStrategy, msg messages.PayloadNotificationMessage) (err error) {

	// Create the request
	req, err := http.NewRequest("GET", msg.DownloadUrl, nil)
	if err != nil {
		log.Println(err)
		log.Println("Unable to build request, possible internal error, please report this message to support")
		os.Exit(1)
	}

	// Add the Authorization header
	req.Header.Set("Authorization", "Bearer "+auth.Token)

	// Send the request
	client := &http.Client{} // Note, this follows redirects by default - we need this to occur!
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Println(err)
		log.Println("Unable to download payload")
		return err
	}

	if resp.StatusCode == http.StatusOK {
		err := writer.WritePayload(resp, msg, folderPath)

		if err != nil {
			return &app_errors.DownloadFailedError{
				Message: fmt.Sprintf("failed to download the payload %s", err),
			}
		}

		return nil

	} else if resp.StatusCode == http.StatusPaymentRequired {
		return fmt.Errorf("account Limits exceeded, please see https://hypermass.io/usage")
	} else if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("token expired while fetching payload %s", strconv.Itoa(resp.StatusCode))
	} else {
		return fmt.Errorf("unexpected response fetching payload %s", strconv.Itoa(resp.StatusCode))
	}
}
