package payload_writers

import (
	"path/filepath"
	"regexp"
	"testing"
)

// A temporary path is unique per write, but still names what it is for.
func TestATemporaryPathIsARandomStringThenTheFinalName(t *testing.T) {
	first := temporaryPath("/subs/_abc", "p_001m4dh7t0aaaaaaaaaaaaaaaa.json")
	second := temporaryPath("/subs/_abc", "p_001m4dh7t0aaaaaaaaaaaaaaaa.json")

	if first == second {
		t.Errorf("expected two writes of one payload to use different temporary paths, both were %s", first)
	}

	if filepath.Dir(first) != filepath.Join("/subs/_abc", ".hypermass", "temporary") {
		t.Errorf("expected the temporary folder, got %s", first)
	}

	name := regexp.MustCompile(`^[A-Z2-7]{26}-p_001m4dh7t0aaaaaaaaaaaaaaaa\.json$`)
	if !name.MatchString(filepath.Base(first)) {
		t.Errorf("expected a random string then the final name, got %s", filepath.Base(first))
	}
}
