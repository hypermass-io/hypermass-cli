package cmd

import (
	reload_command "hypermass-cli/commands/reload-command"

	"github.com/spf13/cobra"
)

var reloadNoWait bool

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Applies configuration changes to a running sync",
	Long: `Applies changes to hypermass-config.yaml to a running "sync" process.

Streams that were added are started, removed ones are stopped, and changed ones are restarted with their new
settings. Unchanged streams carry on undisturbed. Stopping a stream waits for the file it is transferring.

Changes to the access key need the sync to be restarted.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		reload_command.Reload(reloadNoWait)
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)
	reloadCmd.Flags().BoolVar(&reloadNoWait, "no-wait", false, "Return once the changes are planned, without waiting for them to apply")
}
