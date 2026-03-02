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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
