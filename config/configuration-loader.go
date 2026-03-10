package config

import (
	"errors"
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
	SubscriptionConfigurations []SubscriptionConfiguration `yaml:"subscription-targets"`
	PublicationConfigurations  []PublicationConfiguration  `yaml:"publication-sources"`
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
	path := filepath.Join(CreateOrGetConfigPath(), "hypermass-config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read config file %s: %s", path, err)
	}

	if err := yaml.Unmarshal(data, &hypermassProfile.Configuration); err != nil {
		log.Fatalf("invalid YAML in %s: %s", path, err)
	}

	hypermassProfile.Auth = LoadSecretKey()

	return hypermassProfile
}

func LoadSecretKey() HypermassAuth {
	var auth HypermassAuth

	path := filepath.Join(CreateOrGetConfigPath(), "auth.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read config file %s: %s", path, err)
	}

	if err := yaml.Unmarshal(data, &auth); err != nil {
		log.Fatalf("invalid YAML in %s: %s", path, err)
	}

	if !(auth.Type == "bearer-token") {
		log.Fatalf("Unknown auth type: %s", auth.Type)
	}

	return auth
}

//TODO it would be nice if we can warn users about upcoming client deprecation somehow
