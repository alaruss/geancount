package geancount

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Posting is a leg of a transaction
type Posting struct {
	account AccountName
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

// Date return transaction date
func (t Transaction) Date() time.Time {
	return t.date
}

// Apply balance postings of the transaction and change balances
func (t Transaction) Apply(ls *LedgerState) error {
	// Before apply check if all postings can be applied
	for _, p := range t.postings {
		if _, ok := ls.accounts[p.account]; !ok {
			return fmt.Errorf("Account %s is not open", p.account)
		}
		if !ls.accounts[p.account].CurrencyAllowed(p.amount.currency) {
			return fmt.Errorf("Currency %s can not be used in account %s", p.amount.currency, p.account)
		}
	}
	for _, p := range t.postings {
		if _, ok := ls.balances[p.account][p.amount.currency]; !ok {
			ls.balances[p.account][p.amount.currency] = p.amount.value
		} else {
			ls.balances[p.account][p.amount.currency] = ls.balances[p.account][p.amount.currency].Add(p.amount.value)
		}
	}
	return nil
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
	err = d.balancePostings()
	if err != nil {
		return Transaction{}, err
	}
	return d, nil
}

func newPostings(lines []Line) ([]Posting, error) {
	postings := []Posting{}
	for _, line := range lines {
		accountName := line.tokens[0].text
		if !strings.HasPrefix(accountName, "Assets:") && !strings.HasPrefix(accountName, "Equity:") && !strings.HasPrefix(accountName, "Income:") && !strings.HasPrefix(accountName, "Expenses:") && !strings.HasPrefix(accountName, "Liabilities:") {
			continue
		}
		amount := Amount{}
		if len(line.tokens) > 2 {
			amountValue, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[1].text, ",", ""))
			if err != nil {
				return postings, fmt.Errorf("can not parse amount value %s %s", accountName, line.tokens[1].text)
			}
			amount.value = amountValue
			amount.currency = Currency(line.tokens[2].text)
		}
		p := Posting{account: AccountName(accountName), amount: amount}
		postings = append(postings, p)
	}
	return postings, nil
}
