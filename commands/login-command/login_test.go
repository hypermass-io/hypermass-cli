package login_command

import (
	"hypermass-cli/config"
	"os"
	"path/filepath"
	"testing"
)

// answering replaces standard input with the given answers for the duration of the test.
func answering(t *testing.T, answers string) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = writer.WriteString(answers)
	_ = writer.Close()

	original := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() { os.Stdin = original })
}

func TestLoginSavesTheKeyReadableByItsOwnerOnly(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	answering(t, "a-new-key\n")

	Login()

	auth, err := config.ReadSecretKey()
	if err != nil || auth.Token != "a-new-key" {
		t.Fatalf("expected the key to be saved, got %+v (%v)", auth, err)
	}
	info, _ := os.Stat(filepath.Join(configHome, "hypermass", "auth.yaml"))
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected the key file to be readable by its owner only, got %v", info.Mode().Perm())
	}
}

func TestLoginReplacesAKeyAndTightensAnOlderFile(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDirectory := filepath.Join(configHome, "hypermass")
	if err := os.MkdirAll(configDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	//as written by earlier versions, readable by everyone
	if err := os.WriteFile(filepath.Join(configDirectory, "auth.yaml"), []byte("type: bearer-token\ntoken: old-key\n"), 0644); err != nil {
		t.Fatal(err)
	}
	answering(t, "a-new-key\n")

	Login()

	auth, _ := config.ReadSecretKey()
	if auth.Token != "a-new-key" {
		t.Errorf("expected the key to be replaced, got %q", auth.Token)
	}
	info, _ := os.Stat(filepath.Join(configDirectory, "auth.yaml"))
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected the key file to be readable by its owner only, got %v", info.Mode().Perm())
	}
}
