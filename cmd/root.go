package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// TestingMode indicates configuration is in the execution directory
	testingMode bool = false
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Hypermass CLI",
	Short: "Hypermass client - publish, subscribe and query",
	Long:  `Hypermass client - publish, subscribe and query streams`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&testingMode, "test", false, "Testing Mode; get the config from running directory")
}
