package geancount

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

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

const printPrecision = 5

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

// LoadFile reads the file, parses it and adds content Ledger
func (l *Ledger) LoadFile(filename string) error {
	return l.loadFile(filename, true)
}

func (l *Ledger) loadFile(filename string, sortDirectives bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	abs, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	parentDir := filepath.Dir(abs)

	lines, err := parseInput(file)
	if err != nil {
		return err
	}
	lineGroups, err := groupLines(lines)
	if err != nil {
		return err
	}

	err = l.createDirectives(lineGroups, filename, parentDir)
	if sortDirectives {
		l.sortDirectives()
	}
	return err
}

// GetState compute state of ledger
func (l *Ledger) GetState() (LedgerState, error) {
	ls := LedgerState{}
	ls.accounts = map[AccountName]Account{}
	ls.balances = AccountsBalances{}
	errs := []error{}
	for _, directive := range l.directives {
		err := directive.Apply(&ls)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s:%02d %s", directive.FileName(), directive.LineNum(), err.Error()))
		}
	}
	return ls, errors.Join(errs...)
}

// PrintBalances prints to stdput formatted balances for all accounts
func (l *Ledger) PrintBalances(ls LedgerState) error {
	accounts := make([]AccountName, 0, len(ls.accounts))
	accountPad := 0
	for acountName, account := range ls.accounts {
		if !account.hadTransactions {
			continue
		}
		accounts = append(accounts, acountName)
		if len(acountName) > accountPad {
			accountPad = len(acountName)
		}
	}
	accountPad += 4
	slices.SortFunc(accounts, func(i, j AccountName) int {
		return cmp.Compare(string(i), string(j))
	})
	one := decimal.New(1, 0)
	sb := strings.Builder{}
	for _, a := range accounts {
		currencies := make([]Currency, 0, len(ls.balances[a]))
		for c := range ls.balances[a] {
			currencies = append(currencies, c)
		}
		slices.SortFunc(currencies, func(i, j Currency) int {
			return cmp.Compare(string(i), string(j))
		})
		for _, c := range currencies {
			v := ls.balances[a][c]
			// Right padding of a with length accountPad
			sb.WriteString(fmt.Sprintf("%-[1]*[2]s\t", accountPad, a))
			// Left padding of integer part of number
			sb.WriteString(fmt.Sprintf("%10d", v.IntPart()))

			// Decimal part
			frac := v.Mod(one).CoefficientInt64()
			if frac < 0 {
				frac = -frac
			}
			fracS := fmt.Sprintf("%d", frac)
			// If decimal has only one digit add 0 in front
			if len(fracS) == 1 {
				fracS = "0" + fracS
			} else if len(fracS) > printPrecision {
				fracS = fracS[:printPrecision]
			}
			sb.WriteString(fmt.Sprintf(".%-7s", fracS))

			// Currency name
			sb.WriteString(fmt.Sprintf("%s\n", c))
		}
		// Empty account
		if len(currencies) == 0 {
			// Right padding of a with length accountPad
			sb.WriteString(fmt.Sprintf("%-[1]*[2]s\n", accountPad, a))
		}
	}
	fmt.Print(sb.String())
	return nil
}
