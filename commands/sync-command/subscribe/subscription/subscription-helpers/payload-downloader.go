package subscription_helpers

import (
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	"hypermass-cli/config"
	"log"
	"net/http"
	"os"
	"strconv"
)

func DownloadPayload(config config.HypermassConfig, folderPath string, writer payload_writers.PayloadWriterStrategy, msg messages.PayloadNotificationMessage) (err error) {

	// Create the request
	req, err := http.NewRequest("GET", msg.DownloadUrl, nil)
	if err != nil {
		log.Println(err)
		log.Println("Unable to build request, possible internal error, please report this message to support")
		os.Exit(1)
	}

	// Add the Authorization header
	req.Header.Set("Authorization", "Bearer "+config.Token)

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
			os.Exit(1)
		}

		return nil

	} else if resp.StatusCode == http.StatusPaymentRequired {
		log.Println("Account Limits exceeded, please see https://hypermass.io/usage")
		return err
	} else {
		log.Println("Unexpected response " + strconv.Itoa(resp.StatusCode) + " from API, please report this message to support")
		return err
	}
}
