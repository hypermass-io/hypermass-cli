package payload_writers

import (
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sha256 of "{}", base64 encoded
const emptyObjectHash = "RBNvo1WzZ4oRRq0W9+hknpT7T8If536DEMBg9hyq/4o="

// A payload is only written when it matches the hash the info channel announced; one that does not leaves nothing
// behind, so the download can be retried.
func TestAPayloadIsCheckedAgainstItsHash(t *testing.T) {
	sha256Algorithm, md5Algorithm := "sha256", "md5"
	matching, mismatching := emptyObjectHash, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	for _, writer := range []struct {
		name   string
		writer PayloadWriterStrategy
	}{
		{"file-per-payload", &FilePerPayloadStrategy{}},
		{"folder-with-metadata", &FolderWithMetadataStrategy{}},
	} {
		for _, check := range []struct {
			name      string
			algorithm *string
			hash      *string
			written   bool
		}{
			{"a matching hash", &sha256Algorithm, &matching, true},
			{"a mismatching hash", &sha256Algorithm, &mismatching, false},
			{"no hash", nil, nil, true},
			{"an algorithm this version cannot check", &md5Algorithm, &mismatching, true},
		} {
			t.Run(writer.name+" with "+check.name, func(t *testing.T) {
				folder := t.TempDir()
				msg := messages.PayloadNotificationMessage{
					PayloadId: "p_001m4dh7t0aaaaaaaaaaaaaaaa", StreamId: "_abc", FileExtension: "json",
					PublishedTimestamp: "2026-10-08T14:02:11.123Z", ContentHashAlgorithm: check.algorithm,
					ContentHash: check.hash, ValidationType: "basic",
				}

				err := writer.writer.WritePayload(&http.Response{Body: io.NopCloser(strings.NewReader("{}"))}, msg, folder)

				present, _ := writer.writer.IsPresent(msg, folder)
				if present != check.written || (err == nil) != check.written {
					t.Errorf("expected written: %v, got present: %v, error: %v", check.written, present, err)
				}

				leftovers, _ := os.ReadDir(filepath.Join(folder, ".hypermass", "temporary"))
				if len(leftovers) != 0 {
					t.Errorf("expected nothing left in the temporary folder, found %d", len(leftovers))
				}
			})
		}
	}
}
