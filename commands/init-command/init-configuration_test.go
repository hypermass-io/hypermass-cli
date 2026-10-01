package init_command

import (
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

func TestInitAddsAMissingConfigurationAndKeepsTheKey(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDirectory := filepath.Join(configHome, "hypermass")
	if err := os.MkdirAll(configDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	existingKey := "type: bearer-token\ntoken: existing\n"
	if err := os.WriteFile(filepath.Join(configDirectory, "auth.yaml"), []byte(existingKey), 0600); err != nil {
		t.Fatal(err)
	}
	answering(t, "/data/hypermass\n\n")

	InitPrompt()

	key, _ := os.ReadFile(filepath.Join(configDirectory, "auth.yaml"))
	if string(key) != existingKey {
		t.Errorf("expected the existing key to be kept, got %s", key)
	}
	configuration, err := os.ReadFile(filepath.Join(configDirectory, "hypermass-config.yaml"))
	if err != nil {
		t.Fatalf("expected the configuration to be created: %v", err)
	}
	if string(configuration) != "base-directory: /data/hypermass\nsubscription-targets: []\npublication-sources: []\n" {
		t.Errorf("unexpected configuration\n%s", configuration)
	}
}

func TestInitAddsAMissingKeyOwnerReadableOnly(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDirectory := filepath.Join(configHome, "hypermass")
	if err := os.MkdirAll(configDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	existingConfiguration := "base-directory: /data/hypermass\n"
	if err := os.WriteFile(filepath.Join(configDirectory, "hypermass-config.yaml"), []byte(existingConfiguration), 0644); err != nil {
		t.Fatal(err)
	}
	answering(t, "new-key\n")

	InitPrompt()

	configuration, _ := os.ReadFile(filepath.Join(configDirectory, "hypermass-config.yaml"))
	if string(configuration) != existingConfiguration {
		t.Errorf("expected the existing configuration to be kept, got %s", configuration)
	}
	info, err := os.Stat(filepath.Join(configDirectory, "auth.yaml"))
	if err != nil {
		t.Fatalf("expected the key to be created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected the key file to be readable by its owner only, got %v", info.Mode().Perm())
	}
}
