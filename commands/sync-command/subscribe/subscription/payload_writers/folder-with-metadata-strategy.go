package payload_writers

import (
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// FolderWithMetadataStrategy writes the data directly to the specified path.
type FolderWithMetadataStrategy struct{}

func (s *FolderWithMetadataStrategy) WritePayload(resp *http.Response, msg messages.PayloadNotificationMessage, folderPath string) error {
	payloadFilename := "payload." + msg.FileExtension
	finalFolder := filepath.Join(folderPath, msg.PayloadId)
	tempFolder := filepath.Join(folderPath, ".hypermass", "temporary", msg.PayloadId)
	tempOutputPath := filepath.Join(tempFolder, payloadFilename)

	// Create the folder
	err := os.MkdirAll(tempFolder, 0755)
	if err != nil {
		log.Println(err)
		log.Println("Unable to create payload folder")
		return err
	}

	//create the payload temp folder if needed
	out, err := os.Create(tempOutputPath)
	if err != nil {
		log.Println(err)
		log.Println("Unable to create payload file")
		return err
	}

	// Stream to temp file
	_, err = io.Copy(out, resp.Body)

	//close the open http and file handles
	out.Close()

	if err != nil {
		log.Println(err)
		log.Println("Unable to write payload to temporary file")
		return err
	}

	err = updateFileMetadataLastModified(msg.PublishedTimestamp, tempFolder)
	if err != nil {
		log.Fatalf("Error modifying timestamp, cannot guarentee ordering: %v", err)
	}

	//TODO write metadata file too!

	err = moveTempToFinalPath(tempFolder, finalFolder)

	if err != nil {
		log.Println(err)
		log.Println("Unable to write payload to file")
		return err
	}

	return nil
}
