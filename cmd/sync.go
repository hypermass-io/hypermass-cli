package cmd

import (
	sync_command "hypermass-cli/commands/sync-command"
	"hypermass-cli/config"

	"github.com/spf13/cobra"
)

// syncCmd represents the subscribe command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Start streaming your subscriptions and publications from local folders",
	Long: `Start Streaming your subscriptions and publications as configured in the hypermass-config.yaml file.

Each stream consists of a sequence of files. When subscribing to a stream, these will be written to local files in 
sequence - from there you can programmatically read and delete them once processed (or do this by hand). There are 
different "writer types" from simple to a little more involved.

Similarly publications will take files from a directory and publish them to a stream.


`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := config.LoadConfig(testingMode)
		sync_command.SyncRunner(settings)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
