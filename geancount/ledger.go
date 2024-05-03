package geancount

import (
	"io"

	"github.com/shopspring/decimal"
	"golang.org/x/exp/slices"
)

type LedgerState struct {
	accounts map[AccountName]Account
	balances map[AccountName]decimal.Decimal
}

// Ledger ir representing of transaction history
type Ledger struct {
	directives []Directive
}

// NewLedger creates ledger
func NewLedger() *Ledger {
	l := Ledger{}
	return &l
}

// Load parses the beancount input and load it into Ledger
func (l *Ledger) Load(r io.Reader) error {
	lines, err := parseInput(r)
	if err != nil {
		return err
	}
	lineGroups, err := groupLines(lines)
	if err != nil {
		return err
	}
	directives, err := createDirectives(lineGroups)
	slices.SortFunc(directives, func(i, j Directive) int {
		if i.Date().Before(j.Date()) {
			return -1
		} else if i.Date().After(j.Date()) {
			return 1
		}
		return 0
	})
	l.directives = directives
	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetBalances() (map[AccountName]decimal.Decimal, error) {
	ls := LedgerState{}
	ls.accounts = map[AccountName]Account{}
	ls.balances = map[AccountName]decimal.Decimal{}
	for _, directive := range l.directives {
		err := directive.Apply(&ls)
		if err != nil {
			return nil, err
		}
	}
	return ls.balances, nil
}
