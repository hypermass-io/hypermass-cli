package config

import (
	"net/http"
	"testing"
)

func TestAMissingKeyFileMeansNoKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	auth, err := ReadSecretKey()

	if err != nil {
		t.Fatalf("expected a missing key file to read as no key, got %v", err)
	}
	if auth.HasKey() {
		t.Error("expected no key")
	}
}

func TestARequestWithoutAKeyIsSentAnonymously(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example/", nil)

	HypermassAuth{}.Authorize(req)

	if _, present := req.Header["Authorization"]; present {
		t.Error("expected no Authorization header without a key")
	}
}

func TestARequestWithAKeyCarriesIt(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example/", nil)

	HypermassAuth{Type: "bearer-token", Token: "a-key"}.Authorize(req)

	if req.Header.Get("Authorization") != "Bearer a-key" {
		t.Errorf("unexpected Authorization header %q", req.Header.Get("Authorization"))
	}
}
