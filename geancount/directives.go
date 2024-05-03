package geancount

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Directive in interface for all entries in the Ledger
type Directive interface {
	Date() time.Time
	Apply(*LedgerState) error
}

// Balance is state of account at the data
type Balance struct {
	date    time.Time
	account AccountName
	amount  Amount
}

func (b Balance) Date() time.Time {
	return b.date
}

func (b Balance) Apply(ls *LedgerState) error {
	currentBalance, ok := ls.balances[b.account]
	if !ok {
		return fmt.Errorf("Balance of unknown account %s", b.account)
	}
	if !currentBalance.Equal(b.amount.value) {
		return fmt.Errorf("Balance expected %s but calcultated %s", b.amount.value, currentBalance)
	}
	return nil
}

type Price struct {
	date     time.Time
	currency Currency
	amount   Amount
}

func (p Price) Date() time.Time {
	return p.date
}

type AccountOpen struct {
	date       time.Time
	account    AccountName
	currencies []Currency
}

func (a AccountOpen) Date() time.Time {
	return a.date
}

func (a AccountOpen) Apply(ls *LedgerState) error {
	if _, ok := ls.accounts[a.account]; ok {
		return fmt.Errorf("Account %s already open", a.account)
	}
	ls.accounts[a.account] = Account{name: a.account, currencies: a.currencies}
	ls.balances[a.account] = decimal.Zero
	return nil
}

type AccountClose struct {
	date    time.Time
	account AccountName
}

func (a AccountClose) Date() time.Time {
	return a.date
}

func (a AccountClose) Apply(ls *LedgerState) error {
	if _, ok := ls.accounts[a.account]; !ok {
		return fmt.Errorf("Account %s is not open", a.account)
	}
	delete(ls.accounts, a.account)
	delete(ls.balances, a.account)
	return nil
}

var ErrNotDirective = errors.New("not directive")

func newAccountOpen(lg LineGroup) (AccountOpen, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return AccountOpen{}, ErrNotDirective
	}
	if len(line.tokens) > 4 {
		return AccountOpen{}, fmt.Errorf("more tokens than expected")
	}
	accountName := line.tokens[2].text
	d := AccountOpen{date: date, account: AccountName(accountName)}
	if len(line.tokens) == 4 {
		for _, currName := range strings.Split(line.tokens[3].text, ";") {
			currName = strings.Trim(currName, "")
			if len(currName) > 0 {
				d.currencies = append(d.currencies, Currency(currName))
			}
		}
	}
	return d, nil
}

func newAccountClose(lg LineGroup) (AccountClose, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return AccountClose{}, ErrNotDirective
	}
	if len(line.tokens) > 3 {
		return AccountClose{}, fmt.Errorf("more tokens than expected")
	}
	accountName := line.tokens[2].text
	d := AccountClose{date: date, account: AccountName(accountName)}
	return d, nil
}

func newBalance(lg LineGroup) (Balance, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return Balance{}, ErrNotDirective
	}
	if len(line.tokens) > 5 {
		return Balance{}, fmt.Errorf("more tokens than expected")
	}
	accountName := line.tokens[2].text
	amountValue, err := decimal.NewFromString(line.tokens[3].text)
	if err != nil {
		return Balance{}, fmt.Errorf("can not parse amount value")
	}
	amount := Amount{value: amountValue, currency: Currency(line.tokens[4].text)}
	d := Balance{date: date, account: AccountName(accountName), amount: amount}
	return d, nil
}

func newPrice(lg LineGroup) (Price, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return Price{}, ErrNotDirective
	}
	if len(line.tokens) > 4 {
		return Price{}, fmt.Errorf("more tokens than expected")
	}
	currency := Currency(line.tokens[2].text)
	amountValue, err := decimal.NewFromString(line.tokens[3].text)
	if err != nil {
		return Price{}, fmt.Errorf("can not parse amount value")
	}
	amount := Amount{value: amountValue, currency: Currency(line.tokens[4].text)}
	d := Price{date: date, currency: currency, amount: amount}
	return d, nil
}
