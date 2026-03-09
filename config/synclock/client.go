package synclock

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/config"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DialSync connects to a running sync process using the lockfile
func DialSync() (*http.Client, *SyncLock, error) {
	path := filepath.Join(config.CreateOrGetConfigPath(), "sync-lock.yml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("sync process not found (no lockfile at %s)", path)
	}

	var lock SyncLock
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, nil, fmt.Errorf("corrupt lockfile: %w", err)
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// check for a response on the ping endpoint - is the command server there?
	url := fmt.Sprintf("http://127.0.0.1:%d/ping", lock.Port)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Hypermass-Token", lock.ControlToken)

	resp, err := client.Do(req)
	if err != nil {
		// This covers "Connection Refused" (process dead)
		// or "Timeout" (process hung)
		return nil, nil, fmt.Errorf("sync process is unreachable on port %d: %w", lock.Port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("sync process rejected heartbeat (status: %d)", resp.StatusCode)
	}

	// Reset timeout for the actual command execution
	client.Timeout = 10 * time.Second
	return client, &lock, nil
}

// Dispatch sends a command to the running sync process and returns the response.
// This is the primary helper function for commands like 'replay', 'status', etc.
func Dispatch(action string, params map[string]string) (*CommandResponse, error) {
	client, lock, err := DialSync()
	if err != nil {
		return nil, err
	}

	// Build the URL for the universal Command Bus endpoint
	url := fmt.Sprintf("http://127.0.0.1:%d/cmd", lock.Port)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Hypermass-Token", lock.ControlToken)

	// Query Params contain Command parameters
	q := req.URL.Query()
	q.Add("action", action)
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to communicate with sync process: %w", err)
	}
	defer resp.Body.Close()

	var result CommandResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("sync process returned an invalid response: %w", err)
	}

	return &result, nil
}
