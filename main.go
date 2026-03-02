package main

import (
	"flag"
	"hypermass-cli/cmd"
	"log"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	cmd.Execute()
}
