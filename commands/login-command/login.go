package login_command

import (
	"bufio"
	"fmt"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Login asks for an access key and saves it as auth.yaml, readable by its owner only.
func Login() {
	fmt.Println("Enter your access key (create one at https://hypermass.io/access-keys):")
	fmt.Print("> ")
	key, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	key = strings.TrimSpace(key)

	if key == "" {
		fmt.Println("❌ No key entered, nothing changed.")
		os.Exit(1)
	}

	path := filepath.Join(config.CreateOrGetConfigPath(), "auth.yaml")
	if err := saveKey(path, key); err != nil {
		fmt.Printf("❌ Could not save the key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Access key saved to %s\n", path)

	if _, _, err := synclock.DialSync(); err == nil {
		fmt.Println("Restart 'hypermass sync' to use it.")
	}
}

func saveKey(path string, key string) error {
	data, err := yaml.Marshal(config.HypermassAuth{Type: "bearer-token", Token: key})
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}

	//a key file written by an earlier version may be readable by others, and keeps its permissions when rewritten
	return os.Chmod(path, 0600)
}
