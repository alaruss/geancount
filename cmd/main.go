// Package cmd is the CLI entrypoint
package cmd

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/alaruss/geancount/geancount"
	"github.com/urfave/cli/v2"
)

func init() {

}

// CreateCLI creates  CLI interface
func CreateCLI() {
	app := &cli.App{
		Name:  "geancount",
		Usage: "Loads beancount file and prints balances",
		Action: func(cCtx *cli.Context) error {
			filename := cCtx.Args().Get(0)
			file, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer file.Close()
			ledger := geancount.NewLedger()
			err = ledger.Load(file)
			if err != nil {
				return err
			}
			balances, err := ledger.GetBalances()
			if err != nil {
				return err
			}
			err = ledger.PrintBalances(balances)
			if err != nil {
				return err
			}
			return nil
		},
		Commands: []*cli.Command{},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
