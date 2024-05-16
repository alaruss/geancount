package geancount

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewLedger(t *testing.T) {
	ledger := NewLedger()
	assert.NotNil(t, ledger)
	err := ledger.LoadFile("testdata/basic.bean")
	assert.Nil(t, err)
	assert.Equal(t, ledger.operatingCurrencies, []Currency{"EUR"})
}

func TestGetLedgerState(t *testing.T) {
	ledger := NewLedger()
	assert.NotNil(t, ledger)
	ledger.LoadFile("testdata/basic.bean")
	ls, err := ledger.GetState()
	assert.Nil(t, err)
	curr := Currency("EUR")
	x, _ := decimal.NewFromString("79.5")
	assert.True(t, ls.balances[AccountName("Assets:Bank")][curr].Equal(x))
	x, _ = decimal.NewFromString("0")
	assert.True(t, ls.balances[AccountName("Equity:Opening-Balances")][curr].Equal(x))
	x, _ = decimal.NewFromString("20.5")
	assert.True(t, ls.balances[AccountName("Expenses:Food")][curr].Equal(x))
	x, _ = decimal.NewFromString("-100")
	assert.True(t, ls.balances[AccountName("Income:Job")][curr].Equal(x))
}
