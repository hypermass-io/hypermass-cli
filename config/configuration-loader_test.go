package config

import (
	"path/filepath"
	"testing"
)

func TestBaseDirectoryOrDefaultUsesTheConfiguredDirectory(t *testing.T) {
	configured := filepath.Join(t.TempDir(), "streams")

	directory, err := HypermassConfig{BaseDirectory: configured}.BaseDirectoryOrDefault()

	if err != nil || directory != configured {
		t.Errorf("expected %s, got %s (%v)", configured, directory, err)
	}
}

func TestBaseDirectoryOrDefaultFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	directory, err := HypermassConfig{}.BaseDirectoryOrDefault()

	if err != nil || directory != filepath.Join(home, "hypermass") {
		t.Errorf("expected the home fallback, got %s (%v)", directory, err)
	}
}

func TestBaseDirectoryOrDefaultRefusesARelativePath(t *testing.T) {
	_, err := HypermassConfig{BaseDirectory: "~/hypermass"}.BaseDirectoryOrDefault()

	if err == nil {
		t.Error("expected a relative base-directory to be refused")
	}
}
