package geancount

import (
	"os"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewLedger(t *testing.T) {
	ledger := NewLedger()
	assert.NotNil(t, ledger)
	file, err := os.Open("testdata/basic.bean")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	ledger.Load(file)
}

func TestGetBalances(t *testing.T) {
	ledger := NewLedger()
	assert.NotNil(t, ledger)
	file, err := os.Open("testdata/basic.bean")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	ledger.Load(file)
	balances, err := ledger.GetBalances()
	assert.Nil(t, err)
	x, _ := decimal.NewFromString("79.5")
	assert.True(t, balances[AccountName("Assets:Bank")].Equal(x))
	x, _ = decimal.NewFromString("0")
	assert.True(t, balances[AccountName("Equity:Opening-Balances")].Equal(x))
	x, _ = decimal.NewFromString("20.5")
	assert.True(t, balances[AccountName("Expenses:Food")].Equal(x))
	x, _ = decimal.NewFromString("-100")
	assert.True(t, balances[AccountName("Income:Job")].Equal(x))
}
