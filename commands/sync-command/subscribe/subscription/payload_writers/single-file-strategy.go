package payload_writers

import (
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FilePerPayloadStrategy writes the data directly to the specified path.
type FilePerPayloadStrategy struct{}

// Timeout and interval settings for polling
const (
	// Max time to wait for the file to appear
	pollTimeout = 5000 * time.Millisecond
	// How often to check the directory
	pollInterval = 100 * time.Millisecond
)

func (s *FilePerPayloadStrategy) WritePayload(resp *http.Response, msg messages.PayloadNotificationMessage, folderPath string) error {
	filename := msg.PayloadId + "." + msg.FileExtension
	finalOutputPath := filepath.Join(folderPath, filename)
	tempOutputFolderPath := filepath.Join(folderPath, ".hypermass", "temporary")
	tempOutputPath := filepath.Join(tempOutputFolderPath, filename)

	// Create the temp folder if needed
	err := os.MkdirAll(tempOutputFolderPath, 0755)
	if err != nil {
		log.Println(err)
		log.Println("Unable to create temporary payload folder")
		return err
	}

	// Create the file
	out, err := os.Create(tempOutputPath)
	if err != nil {
		log.Println(err)
		log.Println("Unable to create temporary payload file")
		return err
	}

	defer out.Close()

	// Stream to temp file
	_, err = io.Copy(out, resp.Body)

	if err != nil {
		log.Println(err)
		log.Println("Unable to write payload to temporary file")
		return err
	}

	err = updateFileMetadataLastModified(msg.PublishedTimestamp, tempOutputPath)
	if err != nil {
		log.Fatalf("Error modifying timestamp, cannot guarentee ordering: %v", err)
	}

	err = moveTempToFinalPath(tempOutputPath, finalOutputPath)
	if err != nil {
		log.Println(err)
		log.Println("Unable to write payload to file")
		return err
	}

	return nil
}

func updateFileMetadataLastModified(publishedTimestamp string, tempOutputPath string) (error error) {
	pubTime, err := time.Parse(time.RFC3339Nano, publishedTimestamp)
	if err != nil {
		return err
	}

	//modify the payload file time to match the published time
	err = os.Chtimes(tempOutputPath, pubTime, pubTime)
	if err != nil {
		return err
	}

	return nil
}

func moveTempToFinalPath(tempPath string, actualPath string) (err error) {

	err = os.Rename(tempPath, actualPath)
	if err != nil {
		log.Println(err)
		log.Println("Unable to move tmp download file (" + tempPath + ") to final file: " + actualPath)
		return err
	}

	//Give the operating system dcache (or equivalent directory cache) time to do its thing - attempt to ensure visibility ordering
	//exactly matches the call sequence.
	time.Sleep(10 * time.Millisecond)

	return
}
