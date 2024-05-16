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
	LineNum() int
	FileName() string
	Apply(*LedgerState) error
}

type directive struct {
	date     time.Time
	lineNum  int
	fileName string
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

// ErrNotDirective indecates that directive can not be parsed
var ErrNotDirective = errors.New("not directive") // TODO check if needed

// Balance checks the amount of the account at the date
type Balance struct {
	directive
	account AccountName
	amount  Amount
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
		return fmt.Errorf("Balance of %s expected %s but calcultated %s", b.account, b.amount.value, currencyBalance)
	}
	return nil
}

// Price set the price of a currency on the date
type Price struct {
	directive
	currency Currency
	amount   Amount
}

// AccountOpen opens account and optionaly set currencies which can be used
type AccountOpen struct {
	directive
	account    AccountName
	currencies map[Currency]struct{}
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
	directive
	account AccountName
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

func newAccountOpen(lg LineGroup, fileName string) (AccountOpen, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return AccountOpen{}, ErrNotDirective
	}
	if len(line.tokens) > 5 {
		return AccountOpen{}, fmt.Errorf("more tokens than expected")
	}
	accountName := line.tokens[2].text
	d := AccountOpen{
		directive:  directive{date: date, lineNum: line.lineNum, fileName: fileName},
		account:    AccountName(accountName),
		currencies: map[Currency]struct{}{},
	}
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

func newAccountClose(lg LineGroup, fileName string) (AccountClose, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return AccountClose{}, ErrNotDirective
	}
	if len(line.tokens) > 3 {
		return AccountClose{}, fmt.Errorf("more tokens than expected")
	}
	accountName := line.tokens[2].text
	d := AccountClose{
		directive: directive{date: date, lineNum: line.lineNum, fileName: fileName},
		account:   AccountName(accountName),
	}
	return d, nil
}

func newBalance(lg LineGroup, fileName string) (Balance, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return Balance{}, ErrNotDirective
	}
	if len(line.tokens) > 5 {
		return Balance{}, fmt.Errorf("more tokens than expected")
	}
	accountName := line.tokens[2].text
	amountValue, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[3].text, ",", ""))
	if err != nil {
		return Balance{}, fmt.Errorf("can not parse amount value %s", line.tokens[3].text)
	}
	amount := Amount{value: amountValue, currency: Currency(line.tokens[4].text)}
	d := Balance{
		directive: directive{date: date, lineNum: line.lineNum, fileName: fileName},
		account:   AccountName(accountName),
		amount:    amount,
	}
	return d, nil
}

func newPrice(lg LineGroup, fileName string) (Price, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return Price{}, ErrNotDirective
	}
	if len(line.tokens) > 4 {
		return Price{}, fmt.Errorf("more tokens than expected")
	}
	currency := Currency(line.tokens[2].text)
	amountValue, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[3].text, ",", ""))
	if err != nil {
		return Price{}, fmt.Errorf("can not parse amount value")
	}
	amount := Amount{value: amountValue, currency: Currency(line.tokens[4].text)}
	d := Price{
		directive: directive{date: date, lineNum: line.lineNum, fileName: fileName},
		currency:  currency,
		amount:    amount,
	}
	return d, nil
}
