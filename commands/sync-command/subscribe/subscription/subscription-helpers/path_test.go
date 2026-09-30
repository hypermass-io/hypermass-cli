package subscription_helpers

import (
	"hypermass-cli/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeStateFile(t *testing.T, baseFilePath string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(baseFilePath, ".hypermass"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(baseFilePath, ".hypermass", "state.yaml"), []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func readStateFile(t *testing.T, baseFilePath string) StreamState {
	t.Helper()
	state, err := readState(baseFilePath)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func TestNewDirectoryRecordsItsStreamAndDirection(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "stream")

	if err := InitialiseAndCheckDirectory(directory, "_abc", config.DirectionSubscription); err != nil {
		t.Fatal(err)
	}

	state := readStateFile(t, directory)
	if state.StreamId != "_abc" || state.Direction != config.DirectionSubscription {
		t.Errorf("expected the stream and direction to be recorded, got %+v", state)
	}
}

func TestOlderStateIsClaimedAndKeepsItsPosition(t *testing.T) {
	directory := t.TempDir()
	writeStateFile(t, directory, "version: 1\nlast_payload_id: _payload\n")

	if err := InitialiseAndCheckDirectory(directory, "_abc", config.DirectionSubscription); err != nil {
		t.Fatal(err)
	}

	state := readStateFile(t, directory)
	if state.StreamId != "_abc" || state.Direction != config.DirectionSubscription || state.LastPayloadId != "_payload" {
		t.Errorf("expected the stream and direction to be added to the existing state, got %+v", state)
	}
}

func TestDirectoryOfAnotherStreamIsRefused(t *testing.T) {
	directory := t.TempDir()
	writeStateFile(t, directory, "version: 1\nstream_id: _other\ndirection: subscription\nlast_payload_id: _payload\n")

	err := InitialiseAndCheckDirectory(directory, "_abc", config.DirectionSubscription)

	if err == nil || !strings.Contains(err.Error(), "belongs to stream _other") {
		t.Errorf("expected the directory to be refused, got %v", err)
	}
	if state := readStateFile(t, directory); state.StreamId != "_other" {
		t.Errorf("expected the state to be left alone, got %+v", state)
	}
}

func TestDirectoryOfTheOtherDirectionIsRefused(t *testing.T) {
	directory := t.TempDir()
	writeStateFile(t, directory, "version: 1\nstream_id: _abc\ndirection: publication\n")

	err := InitialiseAndCheckDirectory(directory, "_abc", config.DirectionSubscription)

	if err == nil || !strings.Contains(err.Error(), "used by a publication") {
		t.Errorf("expected the directory to be refused, got %v", err)
	}
}

func TestDirectoryOfTheSameStreamAndDirectionIsAccepted(t *testing.T) {
	directory := t.TempDir()
	writeStateFile(t, directory, "version: 1\nstream_id: _abc\ndirection: subscription\nlast_payload_id: _payload\n")

	if err := InitialiseAndCheckDirectory(directory, "_abc", config.DirectionSubscription); err != nil {
		t.Errorf("expected the directory to be accepted, got %v", err)
	}
}

func TestWriteLastPayloadIdKeepsTheStream(t *testing.T) {
	directory := t.TempDir()

	if err := WriteLastPayloadId(directory, "_abc", "_payload"); err != nil {
		t.Fatal(err)
	}

	state := readStateFile(t, directory)
	if state.StreamId != "_abc" || state.Direction != config.DirectionSubscription || state.LastPayloadId != "_payload" {
		t.Errorf("unexpected state %+v", state)
	}
}
