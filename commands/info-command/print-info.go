package info_command

import (
	"fmt"
	"hypermass-cli/app_constants"
)

func PrintInfo(configLocation string) {

	fmt.Printf("<=> Hypermass CLI <=>\n")
	fmt.Printf("---------------------\n")

	fmt.Printf("Version:                   %s\n", app_constants.HypermassCliVersion)
	fmt.Printf("HypermassConfig Location:  %s\n", configLocation)
}

func PrintNotYetConfiguredMessage() {

	PrintInfo("n/a")

	fmt.Printf("\n## Please run 'hypermass init' to initialise the configuration. ##\n")
}
