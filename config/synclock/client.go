package synclock

import (
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
		Timeout: 5 * time.Second,
	}

	// --- Functional Heartbeat Check ---
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
