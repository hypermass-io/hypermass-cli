package cmd

import (
	replay_command "hypermass-cli/commands/replay-command"

	"github.com/spf13/cobra"
)

var payloadId string

var replayCmd = &cobra.Command{
	// The <> brackets tell the user this is a required positional arg
	Use:   "replay [streamId]",
	Short: "Replays a subscription from a historic point",
	Long: `Replays a subscription from a historic point for an active "sync" process.
    
The 'from-payload-id' is exclusive; the stream will resume with the first payload 
appearing AFTER the provided ID.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		streamId := args[0]
		replay_command.Replay(streamId, payloadId)
	},
}

func init() {
	rootCmd.AddCommand(replayCmd)
	replayCmd.Flags().StringVarP(&payloadId, "from-payload-id", "p", "", "The payload Id to reset to")
	replayCmd.MarkFlagRequired("from-payload-id")
}
