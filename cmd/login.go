package cmd

import (
	login_command "hypermass-cli/commands/login-command"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Saves your access key, for a larger allowance and to publish",
	Long: `Saves an access key from your account (create one at https://hypermass.io/access-keys), replacing any key
already saved. Without a key, the CLI subscribes within a free daily allowance per address.

The key is read when the sync starts, so restart a running sync to use a new one.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		login_command.Login()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
