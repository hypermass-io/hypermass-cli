package cmd

import (
	subscribe_command "hypermass-cli/commands/subscribe-command"

	"github.com/spf13/cobra"
)

var subscribeCmd = &cobra.Command{
	Use:   "subscribe [streamId]",
	Short: "Subscribes to a stream, with files delivered to your base directory",
	Long: `Adds a subscription to hypermass-config.yaml, after checking that your access key can subscribe to the stream.

Files are delivered to <base-directory>/subscriptions/<streamId>, one file per payload. If a sync is running, the
subscription starts straight away. For other settings, edit the entry in hypermass-config.yaml.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		subscribe_command.Subscribe(args[0])
	},
}

func init() {
	rootCmd.AddCommand(subscribeCmd)
}
