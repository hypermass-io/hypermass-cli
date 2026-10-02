package cmd

import (
	"fmt"
	"hypermass-cli/commands/info-command"
	"hypermass-cli/config"

	"github.com/spf13/cobra"
)

// syncCmd represents the subscribe command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Prints information about this tool and it's configuration",
	Long:  `Prints information about this tool and it's configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.ExistingConfigurationPath() {
			info_command.PrintInfo("n/a - created on first use", false)
			return
		}

		auth, err := config.ReadSecretKey()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
		}
		info_command.PrintInfo(config.CreateOrGetConfigPath(), auth.HasKey())
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
