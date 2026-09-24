package helpers

import (
	"hypermass-cli/app_common"
	"os"
)

// StopWithAuthenticationFailure prints why the sync is stopping and exits.
//
// Printed as a banner rather than another log line: it is the last thing the command says, it explains
// why everything above it stopped, and it needs to be findable in a terminal that has just scrolled a
// screenful of per-stream failures past the reader. No timestamp prefix, for the same reason.
func StopWithAuthenticationFailure() {
	app_common.PrintBanner([]string{
		"  STOPPING - your key was rejected.",
		"",
		"  The key may be wrong or revoked, or your account may be locked.",
		"  Sign in at hypermass.io to check.",
		"",
		"  Exiting sync mode: authentication failed and nothing in your",
		"  configuration can work until it is fixed.",
	})

	os.Exit(1)
}
