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

func printBalances(cCtx *cli.Context) error {
	filename := cCtx.Args().Get(0)
	ledger := geancount.NewLedger()
	err := ledger.LoadFile(filename)
	if err != nil {
		return err
	}
	ls, err := ledger.GetState()
	if err != nil {
		return err
	}
	err = ledger.PrintBalances(ls)
	if err != nil {
		return err
	}
	return nil
}

func checkLedger(cCtx *cli.Context) error {
	filename := cCtx.Args().Get(0)
	ledger := geancount.NewLedger()
	err := ledger.LoadFile(filename)
	if err != nil {
		return err
	}
	_, err = ledger.GetState()
	if err != nil {
		return err
	}
	return nil
}

// CreateCLI creates  CLI interface
func CreateCLI() {
	app := &cli.App{
		Name:   "geancount",
		Usage:  "A utility to process beancount files. By default print balances",
		Action: printBalances,
		Commands: []*cli.Command{
			{
				Name:    "balances",
				Aliases: []string{"bal"},
				Usage:   "Prints balances",
				Action:  printBalances,
			},
			{
				Name:   "check",
				Usage:  "Check ledger",
				Action: checkLedger,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
