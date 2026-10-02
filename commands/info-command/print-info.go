package info_command

import (
	"fmt"
	"hypermass-cli/app_constants"
)

// PrintInfo prints the version, where the configuration lives, and whether an access key is saved.
func PrintInfo(configLocation string, hasKey bool) {

	fmt.Printf("<=> Hypermass CLI <=>\n")
	fmt.Printf("---------------------\n")

	fmt.Printf("Version:                   %s\n", app_constants.HypermassCliVersion)
	fmt.Printf("Build date:                %s\n", app_constants.BuildDate)
	fmt.Printf("Commit:                    %s\n", app_constants.Commit)
	fmt.Printf("Hypermass Config Location:  %s\n", configLocation)

	if hasKey {
		fmt.Printf("Access key:                saved\n")
	} else {
		fmt.Printf("Access key:                none - subscribing within the free daily allowance ('hypermass login' to add one)\n")
	}
}
