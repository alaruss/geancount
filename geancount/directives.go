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

// ErrNotDirective indecates that directive can not be parsed
var ErrNotDirective = errors.New("not directive") // TODO check if needed

// Balance checks the amount of the account at the date
type Balance struct {
	date    time.Time
	account AccountName
	amount  Amount
}

// Date returns date of the balance
func (b Balance) Date() time.Time {
	return b.date
}

// Apply check the if the calcualted balance is correct
func (b Balance) Apply(ls *LedgerState) error {
	accountBalance, ok := ls.balances[b.account]
	if !ok {
		return fmt.Errorf("Balance of unknown account %s", b.account)
	}
	currencyBalance, ok := accountBalance[b.amount.currency]
	if !ok {
		currencyBalance = decimal.Zero
	}
	if !currencyBalance.Equal(b.amount.value) {
		return fmt.Errorf("Balance expected %s but calcultated %s", b.amount.value, currencyBalance)
	}
	return nil
}

// Price set the price of a currency on the date
type Price struct {
	date     time.Time
	currency Currency
	amount   Amount
}

// Date returns date
func (p Price) Date() time.Time {
	return p.date
}

// AccountOpen opens account and optionaly set currencies which can be used
type AccountOpen struct {
	date       time.Time
	account    AccountName
	currencies map[Currency]struct{}
}

// Date returns date
func (a AccountOpen) Date() time.Time {
	return a.date
}

// Apply adds account to the LedgerState
func (a AccountOpen) Apply(ls *LedgerState) error {
	if _, ok := ls.accounts[a.account]; ok {
		return fmt.Errorf("Account %s already open", a.account)
	}
	ls.accounts[a.account] = Account{name: a.account, currencies: a.currencies}
	ls.balances[a.account] = CurrenciesAmounts{}
	return nil
}

// AccountClose closes the account
type AccountClose struct {
	date    time.Time
	account AccountName
}

// Date returns date
func (a AccountClose) Date() time.Time {
	return a.date
}

// Apply removes account for the LedgerState
func (a AccountClose) Apply(ls *LedgerState) error {
	if _, ok := ls.accounts[a.account]; !ok {
		return fmt.Errorf("Account %s is not open", a.account)
	}
	delete(ls.accounts, a.account)
	delete(ls.balances, a.account)
	return nil
}

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
	d := AccountOpen{date: date, account: AccountName(accountName), currencies: map[Currency]struct{}{}}
	if len(line.tokens) == 4 {
		for _, currName := range strings.Split(line.tokens[3].text, ";") {
			currName = strings.Trim(currName, "")
			if len(currName) > 0 {
				d.currencies[Currency(currName)] = struct{}{}
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
