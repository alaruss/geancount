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
	currencies []Currency
}

func (a Account) String() string {
	return string(a.name)
}

// Amount is value and curency
type Amount struct {
	value    decimal.Decimal
	currency Currency
}
