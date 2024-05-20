package geancount

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Posting is a leg of a transaction
type Posting struct {
	account AccountName
	amount  Amount
}

// Transaction is a movement from one account to another
type Transaction struct {
	directive
	status    string
	payee     string
	narration string
	postings  []Posting
}

// Apply balance postings of the transaction and change balances
func (t Transaction) Apply(ls *LedgerState) error {
	// Before apply check if all postings can be applied
	for _, p := range t.postings {
		acc, ok := ls.accounts[p.account]
		if !ok {
			return fmt.Errorf("Posting to unknow account %s", p.account)
		}
		if acc.IsClosed(t.Date()) {
			return fmt.Errorf("Account %s is closed", p.account)
		}
		if !acc.CurrencyAllowed(p.amount.currency) {
			return fmt.Errorf("Currency %s can not be used in account %s", p.amount.currency, p.account)
		}
	}
	for _, p := range t.postings {
		if !ls.accounts[p.account].hadTransactions {
			acc := ls.accounts[p.account]
			acc.hadTransactions = true
			ls.accounts[p.account] = acc
		}
		if _, ok := ls.balances[p.account][p.amount.currency]; !ok {
			ls.balances[p.account][p.amount.currency] = p.amount.value
		} else {
			ls.balances[p.account][p.amount.currency] = ls.balances[p.account][p.amount.currency].Add(p.amount.value)
		}
	}
	return nil
}

func (t Transaction) balancePostings() error {
	sum := decimal.NewFromInt(0)
	blankIndex := -1
	var currency Currency = ""
	for i, posting := range t.postings {
		if posting.amount.EffectiveCurrency() == "" {
			if blankIndex != -1 {
				return fmt.Errorf("more than one empty posing")
			}
			blankIndex = i
		} else {
			effectiveCurrency := posting.amount.EffectiveCurrency()
			/* TODO Get effective currency from cost ( 10.0 S1 {})
			if currency != "" && currency != effectiveCurrency {
				return fmt.Errorf("more than one effective currencies: %s and %s", currency, effectiveCurrency)
			}
			*/
			sum = sum.Add(posting.amount.EffectiveValue())
			currency = effectiveCurrency
		}
	}
	if blankIndex != -1 {
		t.postings[blankIndex].amount.currency = currency
		t.postings[blankIndex].amount.value = decimal.Zero.Sub(sum)
	}
	return nil
}

func newTransaction(lg LineGroup, fileName string) (Transaction, error) {
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
		status = "*"
	}
	postings, err := newPostings(lg.lines[1:])
	if err != nil {
		return Transaction{}, err
	}

	d := Transaction{
		directive: directive{date: date, lineNum: line.lineNum, fileName: fileName},
		status:    status,
		payee:     payee,
		narration: narration,
		postings:  postings,
	}
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

			if len(line.tokens) > 3 {
				if line.tokens[3].text == "{" {
					if len(line.tokens) > 4 && line.tokens[4].text == "}" {
						amount.atCost = true
					} else {
						if len(line.tokens) < 7 || line.tokens[6].text != "}" {
							return postings, fmt.Errorf("Unbalanced curled bracked in posting %s %s", accountName, line.tokens[1].text)
						}
						cost, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[4].text, ",", ""))
						if err != nil {
							return postings, fmt.Errorf("can not parse exchange rate %s", line.tokens[4].text)
						}
						priceCurrency := Currency(line.tokens[5].text)
						amount.price = &cost
						amount.priceCurrency = &priceCurrency
						amount.atCost = true
					}
				} else if line.tokens[3].text == "@" && len(line.tokens) > 5 {
					price, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[4].text, ",", ""))
					if err != nil {
						return postings, fmt.Errorf("can not parse exchange rate %s", line.tokens[4].text)
					}
					priceCurrency := Currency(line.tokens[5].text)
					amount.price = &price
					amount.priceCurrency = &priceCurrency
				} else if line.tokens[3].text == "@@" && len(line.tokens) > 5 {
					totalPrice, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[4].text, ",", ""))
					if err != nil {
						return postings, fmt.Errorf("can not parse exchange rate %s", line.tokens[4].text)
					}
					price := totalPrice.Div(amount.value)
					priceCurrency := Currency(line.tokens[5].text)
					amount.price = &price
					amount.priceCurrency = &priceCurrency
				}
			}
		}
		p := Posting{account: AccountName(accountName), amount: amount}
		postings = append(postings, p)
	}
	return postings, nil
}
