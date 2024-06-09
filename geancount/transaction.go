package geancount

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Lot struct {
	amount Amount
	cost   Amount
	date   time.Time
	label  string
}

func (l Lot) String() string {
	return fmt.Sprintf("%s {%s} %s %s", l.amount, l.cost, l.date, l.label)
}

// Posting is a leg of a transaction
type Posting struct {
	account AccountName
	amount  Amount
	price   *Amount
	atCost  bool
}

// Transaction is a movement from one account to another
type Transaction struct {
	directive
	status    string
	payee     string
	narration string
	postings  []Posting
}

// Apply balances postings of the transaction and changes balances
func (t Transaction) Apply(ls *LedgerState) error {
	// Balance postings
	blankPostingValue := decimal.Zero
	blankPostingIndex := -1
	var blankPostingCurrency Currency = ""
	for i, posting := range t.postings {
		if posting.amount.currency == "" {
			if blankPostingIndex != -1 {
				return fmt.Errorf("more than one empty posing")
			}
			blankPostingIndex = i
		} else {
			effValue := posting.amount.value
			effCurrency := posting.amount.currency
			if posting.atCost {
				if posting.price != nil {
					effCurrency = posting.price.currency
					effValue = effValue.Mul(posting.price.value)
				} else {
					accountInventory, ok := ls.inventories[posting.account]
					if !ok || len(accountInventory) == 0 {
						return fmt.Errorf("No lots in %s", posting.account)
					}
					inventory, ok := ls.inventories[posting.account][posting.amount.currency]
					if !ok || len(accountInventory) == 0 {
						return fmt.Errorf("No lots for %s in %s", posting.amount.currency, posting.account)
					}

					totalAmount := decimal.Zero
					totalValue := decimal.Zero
					for _, lot := range inventory {
						totalAmount = totalAmount.Add(lot.amount.value)
						totalValue = totalValue.Add(lot.amount.value.Mul(lot.cost.value))
					}
					if !totalAmount.Equal(posting.amount.value.Neg()) {
						return fmt.Errorf("Not all lots are used")
					}
					effCurrency = inventory[0].cost.currency
					effValue = totalValue.Neg()
				}
			}
			blankPostingValue = blankPostingValue.Add(effValue)
			blankPostingCurrency = effCurrency // Can it be not correct?
		}
	}
	if blankPostingIndex != -1 {
		t.postings[blankPostingIndex].amount.currency = blankPostingCurrency
		t.postings[blankPostingIndex].amount.value = decimal.Zero.Sub(blankPostingValue)
	}

	// Apply postings
	for _, posting := range t.postings {
		acc, ok := ls.accounts[posting.account]
		if !ok {
			return fmt.Errorf("Posting to unknow account %s", posting.account)
		}
		if acc.IsClosed(t.Date()) {
			return fmt.Errorf("Account %s is closed", posting.account)
		}
		if !acc.CurrencyAllowed(posting.amount.currency) {
			return fmt.Errorf("Currency %s can not be used in account %s", posting.amount.currency, posting.account)
		}
	}

	for _, p := range t.postings {
		if !ls.accounts[p.account].hadTransactions {
			acc := ls.accounts[p.account]
			acc.hadTransactions = true
			ls.accounts[p.account] = acc
		}
		if p.atCost {
			if p.price != nil {
				if _, ok := ls.inventories[p.account]; !ok {
					ls.inventories[p.account] = map[Currency][]Lot{}
				}
				if _, ok := ls.inventories[p.account][p.amount.currency]; !ok {
					ls.inventories[p.account][p.amount.currency] = []Lot{}
				}
				ls.inventories[p.account][p.amount.currency] = append(ls.inventories[p.account][p.amount.currency], Lot{
					amount: Amount{p.amount.value, p.amount.currency},
					cost:   Amount{p.price.value, p.price.currency},
					date:   t.date, // TODO it can also be date in posting on in cost
					label:  "",     // TODO add support of labels in cost
				})
			} else {
				ls.inventories[p.account][p.amount.currency] = []Lot{} // Used all lots
			}
		}
		if _, ok := ls.balances[p.account][p.amount.currency]; !ok {
			ls.balances[p.account][p.amount.currency] = p.amount.value
		} else {
			ls.balances[p.account][p.amount.currency] = ls.balances[p.account][p.amount.currency].Add(p.amount.value)
		}
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
	return d, nil
}

func newPostings(lines []Line) ([]Posting, error) {
	postings := []Posting{}
	hasEmptyPosting := false
	for _, line := range lines {
		accountName := line.tokens[0].text
		if !strings.HasPrefix(accountName, "Assets:") && !strings.HasPrefix(accountName, "Equity:") && !strings.HasPrefix(accountName, "Income:") && !strings.HasPrefix(accountName, "Expenses:") && !strings.HasPrefix(accountName, "Liabilities:") {
			continue
		}
		p := Posting{account: AccountName(accountName), amount: Amount{}}
		if len(line.tokens) > 2 {
			amountValue, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[1].text, ",", ""))
			if err != nil {
				return postings, fmt.Errorf("can not parse amount value %s %s", accountName, line.tokens[1].text)
			}
			p.amount.value = amountValue
			p.amount.currency = Currency(line.tokens[2].text)

			if len(line.tokens) > 3 {
				if line.tokens[3].text == "{" {
					if len(line.tokens) > 4 && line.tokens[4].text == "}" { // Implicit cost
						p.atCost = true
					} else {
						if len(line.tokens) < 7 || line.tokens[6].text != "}" { // Cost
							return postings, fmt.Errorf("Unbalanced curled bracked in posting %s %s", accountName, line.tokens[1].text)
						}
						cost, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[4].text, ",", ""))
						if err != nil {
							return postings, fmt.Errorf("can not parse exchange rate %s", line.tokens[4].text)
						}
						priceCurrency := Currency(line.tokens[5].text)
						p.price = &Amount{cost, priceCurrency}
						p.atCost = true
					}
				} else if line.tokens[3].text == "@" && len(line.tokens) > 5 { // Price
					price, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[4].text, ",", ""))
					if err != nil {
						return postings, fmt.Errorf("can not parse exchange rate %s", line.tokens[4].text)
					}
					priceCurrency := Currency(line.tokens[5].text)
					p.price = &Amount{price, priceCurrency}
					p.atCost = true
				} else if line.tokens[3].text == "@@" && len(line.tokens) > 5 { // Total price
					totalPrice, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[4].text, ",", ""))
					if err != nil {
						return postings, fmt.Errorf("can not parse exchange rate %s", line.tokens[4].text)
					}
					price := totalPrice.Div(p.amount.value)
					priceCurrency := Currency(line.tokens[5].text)
					p.price = &Amount{price, priceCurrency}
					p.atCost = true
				}
			}
		}
		if p.amount.currency == "" && p.price != nil {
			if hasEmptyPosting {
				return postings, fmt.Errorf("has more than one empty posting %s", accountName)
			}
			hasEmptyPosting = true
		}
		postings = append(postings, p)
	}
	return postings, nil
}
