package geancount

import "time"

// Posting is a leg of a transaction
type Posting struct {
	account Account
	amount  *Amount
}

type Transaction struct {
	date      time.Time
	status    string
	payee     string
	narration string
	postings  []Posting
}

func (t Transaction) Date() time.Time {
	return t.date
}
