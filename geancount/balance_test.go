package geancount

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPad(t *testing.T) {
	ledger := NewLedger()
	assert.NotNil(t, ledger)
	ledger.LoadFile("testdata/pads.bean")
	_, err := ledger.GetState()
	assert.Nil(t, err)
}
