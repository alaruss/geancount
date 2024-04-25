package geancount

import (
	"os"
	"testing"

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
