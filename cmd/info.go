package cmd

import (
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
		if config.ExistingConfigurationPath() {
			info_command.PrintInfo(config.CreateOrGetConfigPath())
		} else {
			info_command.PrintNotYetConfiguredMessage()
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
