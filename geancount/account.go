package geancount

import (
	"time"

	"github.com/shopspring/decimal"
)

// Currency is name like EUR, USD or BOTTLE_CAP
type Currency string

// AccountName is name of account
type AccountName string

// Account is used to apply transactions to
type Account struct {
	name            AccountName
	currencies      map[Currency]struct{}
	hadTransactions bool
	opened          []time.Time
	closed          []time.Time
}

func (a Account) String() string {
	return string(a.name)
}

// CurrencyAllowed checks if currency can be used in the account
func (a Account) CurrencyAllowed(currency Currency) bool {
	if a.currencies == nil || len(a.currencies) == 0 {
		return true
	}
	_, ok := a.currencies[currency]
	return ok
}

// IsOpen check if account was open on a given date
func (a Account) IsOpen(checkDate time.Time) bool {
	for i, openDate := range a.opened {
		if checkDate.Equal(openDate) || checkDate.After(openDate) {
			if len(a.closed) > i {
				closeDate := a.closed[i]
				if checkDate.Equal(closeDate) || checkDate.Before(closeDate) {
					return true
				}
			} else {
				return true
			}
		}
	}
	return false
}

func (a Account) IsClosed(checkDate time.Time) bool {
	return !a.IsOpen(checkDate)
}

// Amount is value and currency
type Amount struct {
	value    decimal.Decimal
	currency Currency
}
