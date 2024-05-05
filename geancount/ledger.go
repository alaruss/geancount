package geancount

import (
	"cmp"
	"fmt"
	"io"
	"math/big"
	"os"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// CurrenciesAmounts is used to hold balances in different currencies
type CurrenciesAmounts map[Currency]decimal.Decimal

// AccountsBalances is balance of various accounts
type AccountsBalances map[AccountName]CurrenciesAmounts

// LedgerState presents current state of all accounts in Ledger
type LedgerState struct {
	accounts map[AccountName]Account
	balances AccountsBalances
}

// Ledger ir representing of transaction history
type Ledger struct {
	directives          []Directive
	operatingCurrencies []Currency
}

// NewLedger creates ledger
func NewLedger() *Ledger {
	l := Ledger{}
	return &l
}

// LoadFile parses the file and loads it into Ledger
func (l *Ledger) LoadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	err = l.Load(file)
	return err
}

// Load parses the input stream and loads it into Ledger
func (l *Ledger) Load(r io.Reader) error {
	lines, err := parseInput(r)
	if err != nil {
		return err
	}
	lineGroups, err := groupLines(lines)
	if err != nil {
		return err
	}

	err = l.createDirectives(lineGroups)
	if err != nil {
		return err
	}
	return nil
}

// GetBalances compute balances for all accounts in ledger
func (l *Ledger) GetBalances() (AccountsBalances, error) {
	ls := LedgerState{}
	ls.accounts = map[AccountName]Account{}
	ls.balances = AccountsBalances{}
	for _, directive := range l.directives {
		err := directive.Apply(&ls)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
	return ls.balances, nil
}

// PrintBalances prints to stdput formatted balances for all accounts
func (l *Ledger) PrintBalances(balances AccountsBalances) error {
	accounts := make([]AccountName, 0, len(balances))
	accountPad := 0
	for a := range balances {
		accounts = append(accounts, a)
		if len(a) > accountPad {
			accountPad = len(a)
		}
	}
	accountPad += 4
	slices.SortFunc(accounts, func(i, j AccountName) int {
		return cmp.Compare(string(i), string(j))
	})
	one := decimal.New(1, 0)
	for _, a := range accounts {
		currencies := make([]Currency, 0, len(balances[a]))
		for c := range balances[a] {
			currencies = append(currencies, c)
		}
		slices.SortFunc(currencies, func(i, j Currency) int {
			return cmp.Compare(string(i), string(j))
		})
		for _, c := range currencies {
			v := balances[a][c]
			frac := v.Mod(one).Coefficient()
			fmt.Printf("%-[1]*[2]s\t", accountPad, a)
			fmt.Printf("%10d", v.IntPart())
			if frac.Cmp(big.NewInt(0)) != 0 {
				fmt.Printf(".%-6d", frac)
			} else {
				fmt.Printf("%-7s", " ")
			}
			fmt.Printf("%s\n", c)
		}
		if len(currencies) == 0 {
			fmt.Printf("%-[1]*[2]s\n", accountPad, a)
		}
	}
	return nil
}
