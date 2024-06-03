// Package cmd is the CLI entrypoint
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/alaruss/geancount/geancount"
	"github.com/urfave/cli/v2"
)

func init() {

}

func printBalances(cCtx *cli.Context) error {
	errs := []error{}
	filename := cCtx.Args().Get(0)
	ledger := geancount.NewLedger()
	err := ledger.LoadFile(filename)
	if err != nil {
		errs = append(errs, err)
	}
	ls, err := ledger.GetState()
	if err != nil {
		errs = append(errs, err)
	}
	err = errors.Join(errs...)
	if err != nil {
		fmt.Printf("%s\n\n", err)
	}
	filterExpression := cCtx.String("filter-expression")

	err = ledger.PrintBalances(ls, filterExpression)
	return err
}

func checkLedger(cCtx *cli.Context) error {
	errs := []error{}
	filename := cCtx.Args().Get(0)
	ledger := geancount.NewLedger()
	err := ledger.LoadFile(filename)
	if err != nil {
		errs = append(errs, err)
	}
	_, err = ledger.GetState()
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// CreateCLI creates  CLI interface
func CreateCLI() {
	app := &cli.App{
		Name:   "geancount",
		Usage:  "A utility to process beancount files. By default check the file",
		Action: checkLedger,
		Commands: []*cli.Command{
			{
				Name:    "balances",
				Aliases: []string{"bal"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "filter-expression",
						Aliases: []string{"e"},
						Usage:   "Filter expression for which account balances to display",
					},
				},
				Usage:  "Prints balances",
				Action: printBalances,
			},
			{
				Name:   "check",
				Usage:  "Check ledger",
				Action: checkLedger,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
