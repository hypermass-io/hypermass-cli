package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AddSubscription appends a subscription to the end of subscription-targets in hypermass-config.yaml, creating the file
// if it is missing. Comments and the order of entries are kept, while the layout is normalised.
func AddSubscription(entry SubscriptionConfiguration) error {
	path := filepath.Join(CreateOrGetConfigPath(), "hypermass-config.yaml")

	original, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("cannot read config file %s: %w", path, err)
	}

	updated, err := appendSubscription(original, entry)
	if err != nil {
		return err
	}

	//written alongside and renamed into place, so the file is never left half written
	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, updated, 0644); err != nil {
		return fmt.Errorf("cannot write config file %s: %w", temporaryPath, err)
	}

	return os.Rename(temporaryPath, path)
}

// appendSubscription returns the configuration with the entry added as the last subscription target.
func appendSubscription(original []byte, entry SubscriptionConfiguration) ([]byte, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(original, &document); err != nil {
		return nil, fmt.Errorf("invalid YAML in the config file: %w", err)
	}

	//an empty file has no document yet
	if document.Kind == 0 {
		document = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}

	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, errors.New("the config file is not a set of settings that can be added to")
	}

	targets := subscriptionTargets(root)
	if targets == nil {
		return nil, errors.New("subscription-targets is not a list that can be added to")
	}

	targets.Content = append(targets.Content, &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
		text("key"), text(entry.Key),
		text("target-directory"), text(entry.TargetDirectory),
		text("writer-type"), text(entry.WriterType),
	}})

	var updated bytes.Buffer
	encoder := yaml.NewEncoder(&updated)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return updated.Bytes(), nil
}

// subscriptionTargets finds the subscription-targets list, adding it or turning an empty value into a list as needed.
func subscriptionTargets(root *yaml.Node) *yaml.Node {
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != "subscription-targets" {
			continue
		}

		value := root.Content[i+1]
		if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
			value.Kind, value.Tag, value.Value = yaml.SequenceNode, "!!seq", ""
		}
		if value.Kind != yaml.SequenceNode {
			return nil
		}

		//an entry added to a [] list is written as a block list
		value.Style = 0
		return value
	}

	key := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "subscription-targets"}
	value := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	root.Content = append(root.Content, key, value)
	return value
}

func text(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}
