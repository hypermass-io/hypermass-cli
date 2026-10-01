package cmd

import (
	publish_command "hypermass-cli/commands/publish-command"

	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish [streamId]",
	Short: "Publishes files from a folder in your base directory to one of your streams",
	Long: `Adds a publication to hypermass-config.yaml, after checking that your access key can publish to the stream.

Files placed in <base-directory>/publications/<streamId> are published to the stream, then deleted. To keep them,
set disposer-type to move-on-success in hypermass-config.yaml. If a sync is running, the publication starts straight
away.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		publish_command.Publish(args[0])
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
