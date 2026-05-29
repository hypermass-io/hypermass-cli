package cmd

import (
	status_command "hypermass-cli/commands/status-command"

	"github.com/spf13/cobra"
)

var statusFormat string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the sync status",
	Long:  `Get the sync status when the sync command is running.`,

	Run: func(cmd *cobra.Command, args []string) {
		status_command.Status(statusFormat)
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusFormat, "format", "human", "Output format: human or json")
	rootCmd.AddCommand(statusCmd)
}
