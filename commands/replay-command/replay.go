package replay_command

import (
	"fmt"
	"hypermass-cli/config/synclock"
	"net/http"
)

func Replay() {
	streamKey := "asdablhk"
	payloadID := "kjasdbna"
	//TODO the above should be parameters

	client, lock, err := synclock.DialSync()
	if err != nil {
		fmt.Printf("Hot-reload unavailable: %v\n", err)
		// TODO possible fallback logic: If sync isn't running, we could manually edit the state.yaml here instead?
		return
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/replay", lock.Port)

	// Create the request
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Hypermass-Token", lock.ControlToken)

	// Add our payload data (simplified for now)
	q := req.URL.Query()
	q.Add("key", streamKey)
	q.Add("id", payloadID)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to signal sync process: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Replay initiated for stream %s from %s\n", streamKey, payloadID)
	} else {
		fmt.Printf("❌ Sync process rejected the replay: %s\n", resp.Status)
	}
}
