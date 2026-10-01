package publish_command

import (
	"hypermass-cli/app_constants"
	"hypermass-cli/config"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// serviceAnswering points the publication check at a local server answering with the status.
func serviceAnswering(t *testing.T, status int) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(`{"fileExtension":"json","fileType":"JSON"}`))
		}
	}))
	t.Cleanup(server.Close)

	original := app_constants.PublicApiUrl
	app_constants.PublicApiUrl = server.URL + "/api"
	t.Cleanup(func() { app_constants.PublicApiUrl = original })
}

func TestCheckCanPublish(t *testing.T) {
	cases := []struct {
		status  int
		refusal string
	}{
		{http.StatusOK, ""},
		{http.StatusNotFound, "no stream with the id _abc"},
		{http.StatusForbidden, "belongs to another account"},
		{http.StatusUnauthorized, "access key was rejected"},
		{http.StatusInternalServerError, "could not check stream _abc"},
	}

	for _, c := range cases {
		serviceAnswering(t, c.status)

		err := checkCanPublish(config.HypermassAuth{Type: "bearer-token", Token: "key"}, "_abc")

		if c.refusal == "" && err != nil {
			t.Errorf("status %d: unexpected refusal %v", c.status, err)
		}
		if c.refusal != "" && (err == nil || !strings.Contains(err.Error(), c.refusal)) {
			t.Errorf("status %d: expected a refusal containing %q, got %v", c.status, c.refusal, err)
		}
	}
}

func TestPublishAddsTheEntryUnderTheBaseDirectory(t *testing.T) {
	serviceAnswering(t, http.StatusOK)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDirectory := filepath.Join(configHome, "hypermass")
	if err := os.MkdirAll(configDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	original := "# my config\nbase-directory: /data/hypermass\n"
	if err := os.WriteFile(filepath.Join(configDirectory, "hypermass-config.yaml"), []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDirectory, "auth.yaml"), []byte("type: bearer-token\ntoken: key\n"), 0600); err != nil {
		t.Fatal(err)
	}

	Publish("_abc")

	updated, _ := os.ReadFile(filepath.Join(configDirectory, "hypermass-config.yaml"))
	expected := "# my config\nbase-directory: /data/hypermass\npublication-sources:\n" +
		"  - key: _abc\n    target-directory: /data/hypermass/publications/_abc\n    disposer-type: delete-on-success\n"
	if string(updated) != expected {
		t.Errorf("unexpected config\n%s", updated)
	}
}
