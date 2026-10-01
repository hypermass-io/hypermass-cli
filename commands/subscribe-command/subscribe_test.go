package subscribe_command

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

// serviceAnswering points the subscription check at a local server answering with the status.
func serviceAnswering(t *testing.T, status int) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(`{"connectionUrl":"ws://localhost/unused"}`))
		}
	}))
	t.Cleanup(server.Close)

	original := app_constants.BulkAuthenticationApiUrl
	app_constants.BulkAuthenticationApiUrl = server.URL + "/api/data/bulk/authorise/infochannel"
	t.Cleanup(func() { app_constants.BulkAuthenticationApiUrl = original })
}

func TestCheckCanSubscribe(t *testing.T) {
	cases := []struct {
		status  int
		warning string
		refusal string
	}{
		{http.StatusOK, "", ""},
		{http.StatusPaymentRequired, "no allowance left", ""},
		{http.StatusNotFound, "", "no stream with the id _abc"},
		{http.StatusForbidden, "", "not available to your access key"},
		{http.StatusUnauthorized, "", "access key was rejected"},
		{http.StatusBadRequest, "", "could not check stream _abc"},
	}

	for _, c := range cases {
		serviceAnswering(t, c.status)

		warning, err := checkCanSubscribe(config.HypermassAuth{Type: "bearer-token", Token: "key"}, "_abc")

		if !strings.Contains(warning, c.warning) || (c.warning == "" && warning != "") {
			t.Errorf("status %d: unexpected warning %q", c.status, warning)
		}
		if c.refusal == "" && err != nil {
			t.Errorf("status %d: unexpected refusal %v", c.status, err)
		}
		if c.refusal != "" && (err == nil || !strings.Contains(err.Error(), c.refusal)) {
			t.Errorf("status %d: expected a refusal containing %q, got %v", c.status, c.refusal, err)
		}
	}
}

func TestSubscribeAddsTheEntryUnderTheBaseDirectory(t *testing.T) {
	serviceAnswering(t, http.StatusOK)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDirectory := filepath.Join(configHome, "hypermass")
	if err := os.MkdirAll(configDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	original := "# my config\nbase-directory: /data/hypermass\n\nsubscription-targets: []\n"
	if err := os.WriteFile(filepath.Join(configDirectory, "hypermass-config.yaml"), []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDirectory, "auth.yaml"), []byte("type: bearer-token\ntoken: key\n"), 0600); err != nil {
		t.Fatal(err)
	}

	Subscribe("_abc")

	updated, _ := os.ReadFile(filepath.Join(configDirectory, "hypermass-config.yaml"))
	expected := "# my config\nbase-directory: /data/hypermass\nsubscription-targets:\n" +
		"  - key: _abc\n    target-directory: /data/hypermass/subscriptions/_abc\n    writer-type: file-per-payload\n"
	if string(updated) != expected {
		t.Errorf("unexpected config\n%s", updated)
	}
}

func TestSubscribeCreatesAMissingConfigFile(t *testing.T) {
	serviceAnswering(t, http.StatusOK)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", "/home/someone")
	configDirectory := filepath.Join(configHome, "hypermass")
	if err := os.MkdirAll(configDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDirectory, "auth.yaml"), []byte("type: bearer-token\ntoken: key\n"), 0600); err != nil {
		t.Fatal(err)
	}

	Subscribe("_abc")

	created, _ := os.ReadFile(filepath.Join(configDirectory, "hypermass-config.yaml"))
	expected := "subscription-targets:\n" +
		"  - key: _abc\n    target-directory: /home/someone/hypermass/subscriptions/_abc\n    writer-type: file-per-payload\n"
	if string(created) != expected {
		t.Errorf("unexpected config\n%s", created)
	}
}
