package geancount

import (
	"io"
)

// Ledger ir representing of transaction history
type Ledger struct {
	Entires []Entry
}

// NewLedger creates ledger
func NewLedger() *Ledger {
	l := Ledger{}
	return &l
}

// Load parses the beancount input and load it into Ledger
func (l *Ledger) Load(r io.Reader) error {
	l.parse(r)
	return nil
}

func (l *Ledger) parse(r io.Reader) error {
	_, err := parseInput(r)
	if err != nil {
		panic(err)
	}
	return nil
}
