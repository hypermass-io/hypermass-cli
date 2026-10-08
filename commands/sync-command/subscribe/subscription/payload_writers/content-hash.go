package payload_writers

import (
	"crypto/sha256"
	"encoding/base64"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"io"
	"log"
)

// copyVerified streams the payload into out, checking it against the hash the info channel announced. A payload with no
// hash, or one this version cannot check, is written unchecked.
func copyVerified(out io.Writer, body io.Reader, msg messages.PayloadNotificationMessage) error {
	if msg.ContentHash == nil || msg.ContentHashAlgorithm == nil {
		_, err := io.Copy(out, body)
		return err
	}

	if *msg.ContentHashAlgorithm != "sha256" {
		log.Printf("Payload %s has a %s hash, which this version of the CLI cannot check - writing it unchecked",
			msg.PayloadId, *msg.ContentHashAlgorithm)
		_, err := io.Copy(out, body)
		return err
	}

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, hash), body); err != nil {
		return err
	}

	actual := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	if actual != *msg.ContentHash {
		return &app_errors.PayloadHashMismatchError{PayloadId: msg.PayloadId, Expected: *msg.ContentHash, Actual: actual}
	}
	return nil
}
