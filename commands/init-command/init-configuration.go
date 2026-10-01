package init_command

import (
	"bufio"
	"fmt"
	"hypermass-cli/config"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// InitPrompt sets up whichever of the access key and the configuration file are missing, leaving existing ones alone.
func InitPrompt() {
	configPath := config.CreateOrGetConfigPath()
	authFilePath := filepath.Join(configPath, "auth.yaml")
	configFilePath := filepath.Join(configPath, "hypermass-config.yaml")

	authExists := fileExists(authFilePath)
	configExists := fileExists(configFilePath)

	if authExists && configExists {
		fmt.Printf("Already set up: access key in %s, configuration in %s. Try running 'hypermass info' for more info\n",
			authFilePath, configFilePath)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	if authExists {
		fmt.Printf("Access key already set (%s), skipping\n", authFilePath)
	} else {
		initialiseAuth(reader, authFilePath)
	}

	if configExists {
		fmt.Printf("Configuration already exists (%s), skipping\n", configFilePath)
	} else {
		initialiseConfiguration(reader, configFilePath)
	}
}

func initialiseAuth(reader *bufio.Reader, authFilePath string) {
	apiKey := strings.TrimSpace(promptUser(reader, "Please enter your API key (create in settings at https://hypermass.io/access-keys):", ""))

	hypermassAuth := config.HypermassAuth{
		Type:  "bearer-token",
		Token: apiKey,
	}
	authFile, _ := yaml.Marshal(hypermassAuth)
	//readable by the owner only, as it holds the access key
	err := os.WriteFile(authFilePath, authFile, 0600)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to write config file: %w", err))
	}

	fmt.Printf("Credentials saved to %s\n", authFilePath)
}

func initialiseConfiguration(reader *bufio.Reader, configFilePath string) {
	usr, _ := user.Current()

	defaultFolder := filepath.Join(usr.HomeDir, "hypermass")
	hotfolderDirectoryInput := strings.TrimSpace(promptUser(reader, fmt.Sprintf("Please enter your Hypermass hotfolder directory (leave blank for default %s):", defaultFolder), defaultFolder))
	//a relative path would depend on where the sync is started from
	for !filepath.IsAbs(hotfolderDirectoryInput) {
		hotfolderDirectoryInput = strings.TrimSpace(promptUser(reader, fmt.Sprintf("The directory must be a full path, such as %s:", defaultFolder), defaultFolder))
	}
	subscribeKeysInput := promptUser(reader, "Please enter one or more stream IDs that you'd like to subscribe to initially (comma separated, or leave blank):", "")

	var subscriptions []config.SubscriptionConfiguration
	if subscribeKeysInput != "" {
		for _, rawKey := range strings.Split(subscribeKeysInput, ",") {
			cleanedKey := strings.TrimSpace(rawKey)

			if len(cleanedKey) > 0 {

				//default subscription
				subscription := config.SubscriptionConfiguration{
					Key:             cleanedKey,
					TargetDirectory: filepath.Join(hotfolderDirectoryInput, "subscriptions", cleanedKey),
					StartPoint:      "latest",
					WriterType:      "file-per-payload",
				}

				subscriptions = append(subscriptions, subscription)
			}
		}
	}

	hypermassConfiguration := config.HypermassConfig{
		BaseDirectory:              hotfolderDirectoryInput,
		SubscriptionConfigurations: subscriptions,
	}

	file, _ := yaml.Marshal(hypermassConfiguration)
	err := os.WriteFile(configFilePath, file, 0644)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to write config file: %w", err))
	}

	fmt.Printf("Configuration saved to %s\n", configFilePath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helper to handle the repetition of prompting
func promptUser(reader *bufio.Reader, label string, defaultValue string) string {
	fmt.Println(label)
	fmt.Print("> ") // Added a simple prompt char for a better "input" feel
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}
