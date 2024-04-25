package geancount

import "io"

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
func (l *Ledger) Load(rc io.Reader) error {
	return nil
}
