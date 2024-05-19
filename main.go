package main

import (
	"log"

	"github.com/alaruss/geancount/cmd"
)

func init() {
	log.SetFlags(0)
}

func main() {
	cmd.CreateCLI()
}
