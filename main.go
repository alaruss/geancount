package main

import (
	"os"

	"github.com/alaruss/geancount/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stderr,
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		}}
	log.Logger = log.Output(output).Level(zerolog.InfoLevel)
}

func main() {
	cmd.CreateCLI()
}
