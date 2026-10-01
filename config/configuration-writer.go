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
// with the base directory if it is missing. Comments and the order of entries are kept, while the layout is normalised.
func AddSubscription(entry SubscriptionConfiguration) error {
	return addEntry(func(original []byte) ([]byte, error) {
		return appendSubscription(original, entry)
	})
}

// AddPublication appends a publication to the end of publication-sources in hypermass-config.yaml, creating the file
// with the base directory if it is missing. Comments and the order of entries are kept, while the layout is normalised.
func AddPublication(entry PublicationConfiguration) error {
	return addEntry(func(original []byte) ([]byte, error) {
		return appendPublication(original, entry)
	})
}

func addEntry(appendTo func(original []byte) ([]byte, error)) error {
	path := filepath.Join(CreateOrGetConfigPath(), "hypermass-config.yaml")

	original, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		original, err = newConfiguration()
	}
	if err != nil {
		return fmt.Errorf("cannot read config file %s: %w", path, err)
	}

	updated, err := appendTo(original)
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

// newConfiguration is the starting point for a configuration file that does not exist yet, recording the base
// directory so the file shows where its streams are stored.
func newConfiguration() ([]byte, error) {
	baseDirectory, err := HypermassConfig{}.BaseDirectoryOrDefault()
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(HypermassConfig{BaseDirectory: baseDirectory})
}

// appendSubscription returns the configuration with the entry added as the last subscription target.
func appendSubscription(original []byte, entry SubscriptionConfiguration) ([]byte, error) {
	return appendToList(original, "subscription-targets", mapping(
		"key", entry.Key,
		"target-directory", entry.TargetDirectory,
		"writer-type", entry.WriterType,
	))
}

// appendPublication returns the configuration with the entry added as the last publication source.
func appendPublication(original []byte, entry PublicationConfiguration) ([]byte, error) {
	return appendToList(original, "publication-sources", mapping(
		"key", entry.Key,
		"target-directory", entry.TargetDirectory,
		"disposer-type", entry.DisposerType,
	))
}

// appendToList returns the configuration with the item added to the end of the named top level list.
func appendToList(original []byte, listName string, item *yaml.Node) ([]byte, error) {
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

	list := topLevelList(root, listName)
	if list == nil {
		return nil, fmt.Errorf("%s is not a list that can be added to", listName)
	}

	list.Content = append(list.Content, item)

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

// topLevelList finds the named list, adding it or turning an empty value into a list as needed.
func topLevelList(root *yaml.Node, listName string) *yaml.Node {
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != listName {
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

	key := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: listName}
	value := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	root.Content = append(root.Content, key, value)
	return value
}

// mapping builds a mapping from alternating field names and values, keeping their order.
func mapping(fieldsAndValues ...string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, value := range fieldsAndValues {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value})
	}
	return node
}
