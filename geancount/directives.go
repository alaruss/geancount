package geancount

import (
	"time"
)

// Directive in interface for all entries in the Ledger
type Directive interface {
	Date() time.Time
}

// Balance is state of account at the data
type Balance struct {
	date    time.Time
	account Account
	amount  Amount
}

func (b Balance) Date() time.Time {
	return b.date
}

type Price struct {
	date     time.Time
	currency Currency
	amount   Amount
}

func (p Price) Date() time.Time {
	return p.date
}

type AccountOpen struct {
	date    time.Time
	account Account
}

func (a AccountOpen) Date() time.Time {
	return a.date
}

type AccountClose struct {
	date    time.Time
	account Account
}

func (a AccountClose) Date() time.Time {
	return a.date
}
