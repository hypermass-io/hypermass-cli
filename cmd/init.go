package cmd

import (
	init_command "hypermass-cli/commands/init-command"

	"github.com/spf13/cobra"
)

// syncCmd represents the subscribe command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise the Hypermass configuration",
	Long:  `Initialise the Hypermass configuration, creating credential and config files`,
	Run: func(cmd *cobra.Command, args []string) {

		init_command.InitPrompt()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
