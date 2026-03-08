package publication_helpers

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/app_constants"
	"hypermass-cli/config"
	"io"
	"log"
	"net/http"
	"os"
)

type BulkStreamConfiguration struct {
	FileExtension string `json:"fileExtension"`
	FileType      string `json:"fileType"`
}

// GetConfigurationForStream retrieve key information about the stream (particularly the type of data)
func GetConfigurationForStream(hypermassProfile config.HypermassProfile, streamId string) BulkStreamConfiguration {
	url := app_constants.PublicApiUrl + "/data/bulk/id/" + streamId + "/write-configuration"

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		log.Println("Failed to authenticate, unable to construct auth request. Please report this message to support")
		os.Exit(1)
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
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			log.Println("Not authorized to upload to this stream, it must be owned by your account (also check key credentials are configured correctly)")
			os.Exit(1)
		} else if resp.StatusCode == 404 {
			log.Println("Unable to get stream metadata, stream not found")
			os.Exit(1)
		} else {
			log.Printf("Unable to get stream metadata, error response: %d \n", resp.StatusCode)
			os.Exit(1)
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var result BulkStreamConfiguration
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
		os.Exit(1)
	}

	return result
}
