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
	"os"
)

type AuthResponse struct {
	ConnectionURL string `json:"connectionUrl"`
}

func GetAuthorizedSubscriptionUrl(config config.HypermassConfig, streamId string, lastPayloadId string) (string, error) {

	authenticationUrl := app_constants.GetBulkAuthenticationApiUrl(streamId, lastPayloadId)

	// Create the request
	req, err := http.NewRequest("GET", authenticationUrl, nil)
	if err != nil {
		log.Println(err)
		log.Println("Failed to authenticate, unable to construct auth request. Please report this message to support")
		os.Exit(1)
	}

	// Add the Authorization header
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("User-Agent", "hypermass-cli/"+app_constants.HypermassCliVersion)

	// Send the request
	client := &http.Client{} // follows redirects by default
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		log.Println("Failed to authenticate, unable to connect to service")
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 402 {
			return "", &app_errors.InsufficientAllowanceError{Message: "insufficient allowance to subscribe to this feed"}
		} else {
			log.Println("Failed to authenticate, please check that your API key is valid")
			os.Exit(1)
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body) // response body is []byte

	var result AuthResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	location := result.ConnectionURL

	return location, err
}
