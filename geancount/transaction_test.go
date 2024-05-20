package geancount

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriceInPostings(t *testing.T) {
	ledger := NewLedger()
	assert.NotNil(t, ledger)
	ledger.LoadFile("testdata/prices.bean")
	ls, err := ledger.GetState()
	assert.Nil(t, err)
	assert.NotNil(t, ls)
}
