package geancount

import (
	"github.com/shopspring/decimal"
)

// Currency is name like EUR, USD or BOTTLE_CAP
type Currency string

// Account is where data is stored
type Account struct {
	name       string
	currencies []Currency
}

func (a Account) String() string {
	return a.name
}

// Amount is value and curency
type Amount struct {
	value    decimal.Decimal
	currency Currency
}
