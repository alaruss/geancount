package geancount

import (
	"errors"
	"time"
)

const defaultDirectiveOrder = 100000

// Directive in interface for all entries in the Ledger
type Directive interface {
	Date() time.Time
	LineNum() int
	FileName() string
	Order() int

	Apply(*LedgerState) error
}

type directive struct {
	date     time.Time
	lineNum  int
	fileName string
	order    int
}

func (d directive) Date() time.Time {
	return d.date
}

func (d directive) LineNum() int {
	return d.lineNum
}

func (d directive) FileName() string {
	return d.fileName
}

func (d directive) Order() int {
	if d.order == 0 {
		return defaultDirectiveOrder
	}
	return d.order
}

// ErrNotDirective indicates that line can not be parsed
var ErrNotDirective = errors.New("not directive")
