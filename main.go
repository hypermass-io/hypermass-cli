package main

import (
	"hypermass-cli/cmd"
	"log"
)

func main() {
	log.SetFlags(0)

	cmd.Execute()
}
