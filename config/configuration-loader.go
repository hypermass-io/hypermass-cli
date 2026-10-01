package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SubscriptionConfiguration struct {
	Key             string `yaml:"key"`
	TargetDirectory string `yaml:"target-directory"`
	StartPoint      string `yaml:"start-point"`
	WriterType      string `yaml:"writer-type"`
}

type PublicationConfiguration struct {
	Key             string `yaml:"key"`
	TargetDirectory string `yaml:"target-directory"`
	DisposerType    string `yaml:"disposer-type"`
}

// HypermassProfile The full execution profile of the hypermass command, including configuration and authentication data
type HypermassProfile struct {
	Configuration HypermassConfig
	Auth          HypermassAuth
}

type HypermassConfig struct {
	// BaseDirectory is where commands such as subscribe place new streams' target directories
	BaseDirectory              string                      `yaml:"base-directory,omitempty"`
	SubscriptionConfigurations []SubscriptionConfiguration `yaml:"subscription-targets"`
	PublicationConfigurations  []PublicationConfiguration  `yaml:"publication-sources"`
}

// BaseDirectoryOrDefault returns the base directory, or <home>/hypermass when none is configured.
func (c HypermassConfig) BaseDirectoryOrDefault() (string, error) {
	if c.BaseDirectory != "" {
		if !filepath.IsAbs(c.BaseDirectory) {
			return "", fmt.Errorf("base-directory must be an absolute path, got %s", c.BaseDirectory)
		}
		return c.BaseDirectory, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to find the home directory for the default base-directory: %w", err)
	}

	return filepath.Join(home, "hypermass"), nil
}

type HypermassAuth struct {
	Type  string `yaml:"type"`
	Token string `yaml:"token"`
}

func ExistingConfigurationPath() bool {
	cfgRoot, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("failed to resolve config dir: %v", err)
	}

	cfgRoot = filepath.Join(cfgRoot, "hypermass")
	_, err = os.Stat(cfgRoot)

	if err == nil {
		return true
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false
	}

	log.Fatalf("failed to resolve config dir: %v", err)
	return false
}

// CreateOrGetConfigPath gets the config path, creating the hypermass folder if needed
func CreateOrGetConfigPath() string {
	cfgRoot, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("failed to resolve config dir: %v", err)
	}
	cfgRoot = filepath.Join(cfgRoot, "hypermass")

	err = os.MkdirAll(cfgRoot, 0700)
	if err != nil {
		log.Fatalf("failed to create missing config dir(%s): %v", cfgRoot, err)
	}

	return cfgRoot
}

func LoadProfile() HypermassProfile {
	var hypermassProfile HypermassProfile

	configuration, err := ReadConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	configuration, warnings := configuration.Deduplicated()
	for _, warning := range warnings {
		log.Println("⚠️ " + warning)
	}

	hypermassProfile.Configuration = configuration
	hypermassProfile.Auth = LoadSecretKey()

	return hypermassProfile
}

// ReadConfiguration reads hypermass-config.yaml. A missing file reads as an empty configuration, like an empty file.
func ReadConfiguration() (HypermassConfig, error) {
	var configuration HypermassConfig
	path := filepath.Join(CreateOrGetConfigPath(), "hypermass-config.yaml")

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return configuration, nil
	}
	if err != nil {
		return HypermassConfig{}, fmt.Errorf("cannot read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &configuration); err != nil {
		return HypermassConfig{}, fmt.Errorf("invalid YAML in %s: %w", path, err)
	}

	return configuration, nil
}

func LoadSecretKey() HypermassAuth {
	auth, err := ReadSecretKey()
	if err != nil {
		log.Fatal(err)
	}

	return auth
}

// ReadSecretKey reads auth.yaml
func ReadSecretKey() (HypermassAuth, error) {
	var auth HypermassAuth

	path := filepath.Join(CreateOrGetConfigPath(), "auth.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return HypermassAuth{}, fmt.Errorf("cannot read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &auth); err != nil {
		return HypermassAuth{}, fmt.Errorf("invalid YAML in %s: %w", path, err)
	}

	if !(auth.Type == "bearer-token") {
		return HypermassAuth{}, fmt.Errorf("Unknown auth type: %s", auth.Type)
	}

	return auth, nil
}

//TODO it would be nice if we can warn users about upcoming client deprecation somehow
