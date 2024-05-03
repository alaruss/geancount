// Package cmd is the CLI entrypoint
package cmd

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/alaruss/geancount/geancount"
	"github.com/urfave/cli/v2"
)

func init() {

}

// CreateCLI creates  CLI interface
func CreateCLI() {
	app := &cli.App{
		Name:  "geancount",
		Usage: "Loads and process beancount file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "main.bean",
				Usage:   "File to process",
			},
		},
		Action: func(cCtx *cli.Context) error {
			filename := cCtx.String("file")
			file, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			ledger := geancount.NewLedger()
			err = ledger.Load(file)
			if err != nil {
				panic(err)
			}
			balances, err := ledger.GetBalances()
			if err != nil {
				panic(err)
			}
			keys := make([]geancount.AccountName, 0, len(balances))
			for k := range balances {
				keys = append(keys, k)
			}
			slices.SortFunc(keys, func(i, j geancount.AccountName) int {
				return cmp.Compare(string(i), string(j))
			})
			for _, k := range keys {
				fmt.Printf("%s: %s\n", k, balances[k])
			}
			return nil
		},
		Commands: []*cli.Command{},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
