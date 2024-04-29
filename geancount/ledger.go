package geancount

import (
	"io"
)

// Ledger ir representing of transaction history
type Ledger struct {
	Directives []Directive
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
	l.Directives, err = createDirectives(lineGroups)
	if err != nil {
		return err
	}
	return nil
}
