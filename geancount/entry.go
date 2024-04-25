package geancount

import "time"

// Entry in interface for all entries in the Ledger
type Entry interface {
	Date() time.Time
}
