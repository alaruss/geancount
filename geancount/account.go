package geancount

import (
	"github.com/shopspring/decimal"
)

// Currency is name like EUR, USD or BOTTLE_CAP
type Currency string

// AccountName is name of account
type AccountName string

// Account is where data is stored
type Account struct {
	name       AccountName
	currencies map[Currency]struct{}
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

// Amount is value and currency
type Amount struct {
	value    decimal.Decimal
	currency Currency
}
