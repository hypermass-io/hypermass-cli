package payload_writers

import (
	"encoding/json"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Each payload's folder holds the payload and a metadata.json carrying the info channel's fields.
func TestAPayloadFolderHoldsThePayloadAndItsMetadata(t *testing.T) {
	folder := t.TempDir()
	algorithm, hash := "sha256", "RBNvo1WzZ4oRRq0W9+hknpT7T8If536DEMBg9hyq/4o="

	write(t, folder, messages.PayloadNotificationMessage{
		PayloadId: "p_001m4dh7t0hh8xpngcphahtex2g", StreamId: "_254gGE43g", FileExtension: "json",
		PublishedTimestamp: "2026-10-08T14:02:11.123Z", BytesCount: 2, ContentHashAlgorithm: &algorithm,
		ContentHash: &hash, ValidationType: "schema",
	})

	payload, err := os.ReadFile(filepath.Join(folder, "p_001m4dh7t0hh8xpngcphahtex2g", "payload.json"))
	if err != nil || string(payload) != "{}" {
		t.Fatalf("expected the payload beside its metadata, got %q (%v)", payload, err)
	}

	expected := `{
  "formatVersion": 1,
  "payloadId": "p_001m4dh7t0hh8xpngcphahtex2g",
  "streamId": "_254gGE43g",
  "publishedTimestamp": "2026-10-08T14:02:11.123Z",
  "payloadFile": "payload.json",
  "bytesCount": 2,
  "contentHashAlgorithm": "sha256",
  "contentHash": "RBNvo1WzZ4oRRq0W9+hknpT7T8If536DEMBg9hyq/4o=",
  "validationType": "schema"
}`
	if metadata := readMetadata(t, folder, "p_001m4dh7t0hh8xpngcphahtex2g"); metadata != expected {
		t.Errorf("unexpected metadata.json:\n%s", metadata)
	}
}

// A payload stored before hashes were recorded still has both hash fields, as nulls.
func TestAPayloadWithoutAHashHasNullHashFields(t *testing.T) {
	folder := t.TempDir()

	write(t, folder, messages.PayloadNotificationMessage{
		PayloadId: "p_001m4dh7t0aaaaaaaaaaaaaaaa", StreamId: "_254gGE43g", FileExtension: "json",
		PublishedTimestamp: "2026-10-08T14:02:11.123Z", BytesCount: 2, ValidationType: "basic",
	})

	var metadata map[string]any
	if err := json.Unmarshal([]byte(readMetadata(t, folder, "p_001m4dh7t0aaaaaaaaaaaaaaaa")), &metadata); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"contentHashAlgorithm", "contentHash"} {
		value, present := metadata[field]
		if !present || value != nil {
			t.Errorf("expected %s to be present and null, got %v (present: %v)", field, value, present)
		}
	}
}

func write(t *testing.T, folder string, msg messages.PayloadNotificationMessage) {
	t.Helper()
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("{}"))}
	if err := (&FolderWithMetadataStrategy{}).WritePayload(resp, msg, folder); err != nil {
		t.Fatal(err)
	}
}

func readMetadata(t *testing.T, folder string, payloadId string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(folder, payloadId, "metadata.json"))
	if err != nil {
		t.Fatalf("expected a metadata.json: %v", err)
	}
	return string(data)
}
