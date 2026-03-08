package cmd

import (
	replay_command "hypermass-cli/commands/replay-command"

	"github.com/spf13/cobra"
)

// syncCmd represents the subscribe command
var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replays a subscription from a historic point",
	Long:  `Replays a subscription from a historic point for an active "sync"`,
	Run: func(cmd *cobra.Command, args []string) {
		replay_command.Replay()
	},
}

func init() {
	rootCmd.AddCommand(replayCmd)
}
