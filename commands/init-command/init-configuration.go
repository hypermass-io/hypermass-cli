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

func InitPrompt() {
	if config.ExistingConfigurationPath() {
		configPath := config.CreateOrGetConfigPath()

		fmt.Printf("Configuration already exists (%s), try running 'hypermass about' for more info", configPath)
	} else {
		initialiseConfiguration()
	}
}

func initialiseConfiguration() {
	usr, _ := user.Current()

	reader := bufio.NewReader(os.Stdin)

	apiKey := promptUser(reader, "Please enter your API key (create in settings at https://hypermass.io/api-keys):", "")

	defaultFolder := filepath.Join(usr.HomeDir, "hypermass")
	hotfolderDirectoryInput := promptUser(reader, fmt.Sprintf("Please enter your Hypermass hotfolder directory (leave blank for default %s):", defaultFolder), defaultFolder)
	subscribeKeysInput := promptUser(reader, "Please enter one or more API keys that you'd like to subscribe to initially (comma separated):", "")

	var subscriptions []config.SubscriptionConfiguration
	if subscribeKeysInput != "" {
		for _, cleanedKey := range strings.Split(subscribeKeysInput, ",") {
			strings.TrimSpace(cleanedKey)

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
		Keyfile:                    apiKey,
		SubscriptionConfigurations: subscriptions,
	}

	configPath := config.CreateOrGetConfigPath()
	fmt.Printf("Initialising Hypermass configuration directory: %s\n\n", configPath)

	file, _ := yaml.Marshal(hypermassConfiguration)
	hypermassConfigFilePath := filepath.Join(configPath, "hypermass-config.yaml")
	err := os.WriteFile(hypermassConfigFilePath, file, 0644)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to write config file: %w", err))
	}

	hypermassAuth := config.HypermassAuth{
		Type:  "bearer-token",
		Token: apiKey,
	}
	authFile, _ := yaml.Marshal(hypermassAuth)
	hypermassAuthFilePath := filepath.Join(configPath, "auth.yaml")
	err = os.WriteFile(hypermassAuthFilePath, authFile, 0644)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to write config file: %w", err))
	}

	fmt.Printf("Saved \n")

	fmt.Printf("\nCredentials saved to %s\n", hypermassAuthFilePath)
	fmt.Printf("\nConfiguration saved to %s\n", hypermassConfigFilePath)
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
