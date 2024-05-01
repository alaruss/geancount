package geancount

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Posting is a leg of a transaction
type Posting struct {
	account Account
	amount  Amount
}

// Transaction is a movement from one account to another
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

func (t Transaction) balancePostings() error {
	// TODO Deal with many currencies
	sum := decimal.NewFromInt(0)
	blankIndex := -1
	var curency Currency
	for i, posting := range t.postings {
		if posting.amount.currency == "" {
			if blankIndex != -1 {
				return fmt.Errorf("more than one empty posing")
			}
			blankIndex = i
		} else {
			sum = sum.Add(posting.amount.value)
			curency = posting.amount.currency
		}
	}
	if blankIndex != -1 {
		t.postings[blankIndex].amount.currency = curency
		t.postings[blankIndex].amount.value = decimal.Zero.Sub(sum)
	}
	return nil
}

func newTransaction(lg LineGroup) (Transaction, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return Transaction{}, ErrNotDirective
	}
	status := line.tokens[1].text
	var payee, narration string
	if len(line.tokens) >= 4 {
		payee = line.tokens[2].text
		narration = line.tokens[3].text
	} else if len(line.tokens) >= 3 {
		narration = line.tokens[2].text
	}
	if status == "txn" {
		status = line.tokens[len(line.tokens)-1].text
	}
	if status != "?" && status != "*" {
		return Transaction{}, fmt.Errorf("unknown status %s", status)
	}
	postings, err := newPostings(lg.lines[1:])
	if err != nil {
		return Transaction{}, err
	}

	d := Transaction{date: date, status: status, payee: payee, narration: narration, postings: postings}
	return d, nil
}

func newPostings(lines []Line) ([]Posting, error) {
	postings := []Posting{}
	for _, line := range lines {
		accountName := line.tokens[0].text
		amount := Amount{}
		if len(line.tokens) > 1 {
			amountValue, err := decimal.NewFromString(line.tokens[1].text)
			if err != nil {
				return postings, fmt.Errorf("can not parse amount value")
			}
			amount.value = amountValue
			amount.currency = Currency(line.tokens[2].text)
		}
		p := Posting{account: Account{name: accountName}, amount: amount}
		postings = append(postings, p)
	}
	return postings, nil
}
