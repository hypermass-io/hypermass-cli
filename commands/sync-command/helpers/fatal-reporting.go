package helpers

import (
	"fmt"
	"hypermass-cli/app_common"
	"os"
	"time"
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
		"",
		fmt.Sprintf("  Exiting in %d seconds (press Ctrl+C to exit now).", int(authenticationFailureExitDelay.Seconds())),
	})

	//a service manager restarting the sync straight away would otherwise retry the rejected key in a tight loop
	time.Sleep(authenticationFailureExitDelay)
	os.Exit(1)
}

const authenticationFailureExitDelay = 60 * time.Second
