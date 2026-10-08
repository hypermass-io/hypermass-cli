package payload_writers

import (
	"encoding/json"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// FolderWithMetadataStrategy writes the data directly to the specified path.
type FolderWithMetadataStrategy struct{}

func (s *FolderWithMetadataStrategy) IsPresent(msg messages.PayloadNotificationMessage, folderPath string) (bool, error) {
	return pathExists(filepath.Join(folderPath, msg.PayloadId))
}

func (s *FolderWithMetadataStrategy) WritePayload(resp *http.Response, msg messages.PayloadNotificationMessage, folderPath string) error {
	payloadFilename := "payload." + msg.FileExtension
	finalFolder := filepath.Join(folderPath, msg.PayloadId)
	tempFolder := temporaryPath(folderPath, msg.PayloadId)
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
	err = copyVerified(out, resp.Body, msg)

	//close the open http and file handles
	out.Close()

	if err != nil {
		_ = os.RemoveAll(tempFolder)
		log.Println(err)
		log.Println("Unable to write payload to temporary file")
		return err
	}

	err = writeMetadata(msg, payloadFilename, filepath.Join(tempFolder, "metadata.json"))
	if err != nil {
		log.Println(err)
		log.Println("Unable to write payload metadata")
		return err
	}

	//after the metadata is written, so writing it does not move the folder's time on
	err = updateFileMetadataLastModified(msg.PublishedTimestamp, tempFolder)
	if err != nil {
		log.Fatalf("Error modifying timestamp, cannot guarentee ordering: %v", err)
	}

	err = moveTempToFinalPath(tempFolder, finalFolder)

	if err != nil {
		log.Println(err)
		log.Println("Unable to write payload to file")
		return err
	}

	return nil
}

// payloadMetadata is the metadata.json written beside each payload: the info channel's fields, by the same names.
type payloadMetadata struct {
	FormatVersion        int     `json:"formatVersion"`
	PayloadId            string  `json:"payloadId"`
	StreamId             string  `json:"streamId"`
	PublishedTimestamp   string  `json:"publishedTimestamp"`
	PayloadFile          string  `json:"payloadFile"`
	BytesCount           int64   `json:"bytesCount"`
	ContentHashAlgorithm *string `json:"contentHashAlgorithm"`
	ContentHash          *string `json:"contentHash"`
	ValidationType       string  `json:"validationType"`
}

func writeMetadata(msg messages.PayloadNotificationMessage, payloadFilename string, path string) error {
	metadata := payloadMetadata{
		FormatVersion:        1,
		PayloadId:            msg.PayloadId,
		StreamId:             msg.StreamId,
		PublishedTimestamp:   msg.PublishedTimestamp,
		PayloadFile:          payloadFilename,
		BytesCount:           msg.BytesCount,
		ContentHashAlgorithm: msg.ContentHashAlgorithm,
		ContentHash:          msg.ContentHash,
		ValidationType:       msg.ValidationType,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
