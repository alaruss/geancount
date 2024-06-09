package geancount

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Currency is name like EUR, USD or BOTTLE_CAP
type Currency string

// AccountName is name of account
type AccountName string

// Account stores information about account - check account, cash, expnense, etc.
type Account struct {
	name            AccountName
	currencies      map[Currency]struct{}
	hadTransactions bool
	opened          []time.Time
	closed          []time.Time
	pad             *Pad
}

func (a Account) String() string {
	return string(a.name)
}

// CurrencyAllowed checks if currency can be used in the account
func (a Account) CurrencyAllowed(currency Currency) bool {
	if a.currencies == nil || len(a.currencies) == 0 {
		return true
	}
	_, ok := a.currencies[currency]
	return ok
}

// IsOpen check if account is open on a given date
func (a Account) IsOpen(checkDate time.Time) bool {
	for i, openDate := range a.opened {
		if checkDate.Equal(openDate) || checkDate.After(openDate) {
			if len(a.closed) > i {
				closeDate := a.closed[i]
				if checkDate.Equal(closeDate) || checkDate.Before(closeDate) {
					return true
				}
			} else {
				return true
			}
		}
	}
	return false
}

// IsClosed check if account is closed on a given date
func (a Account) IsClosed(checkDate time.Time) bool {
	return !a.IsOpen(checkDate)
}

// Amount is value and currency
type Amount struct {
	value    decimal.Decimal
	currency Currency
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.value, a.currency)
}

// Negative return -a
func (a Amount) Negative() Amount {
	return Amount{
		value:    a.value.Neg(),
		currency: a.currency,
	}
}

// AccountOpen opens account and optionaly set currencies which can be used
type AccountOpen struct {
	directive
	account    AccountName
	currencies map[Currency]struct{}
}

// Apply adds account to the LedgerState
func (a AccountOpen) Apply(ls *LedgerState) error {
	acc, ok := ls.accounts[a.account]
	if !ok {
		ls.accounts[a.account] = Account{name: a.account, currencies: a.currencies, opened: []time.Time{a.Date()}}
		ls.balances[a.account] = CurrenciesAmounts{}
	} else {
		if acc.IsOpen(a.Date()) {
			return fmt.Errorf("Account %s is already open", a.account)
		}
		// If not first AccountOpen doesn't hove currencies we assume thier are the same
		// Otherwise ensure that are equal to the Account currencies
		if len(a.currencies) > 0 {
			actualCurrencies := make([]Currency, 0, len(acc.currencies))
			for c := range acc.currencies {
				actualCurrencies = append(actualCurrencies, c)
			}
			newCurrencies := make([]Currency, 0, len(a.currencies))
			for c := range a.currencies {
				newCurrencies = append(newCurrencies, c)
			}
			if len(newCurrencies) != len(actualCurrencies) {
				return fmt.Errorf("Account %s can not change currencies", a.account)
			}
			slices.SortStableFunc(actualCurrencies, func(i, j Currency) int {
				return cmp.Compare(string(i), string(j))
			})
			slices.SortStableFunc(newCurrencies, func(i, j Currency) int {
				return cmp.Compare(string(i), string(j))
			})
			for i := range actualCurrencies {
				if actualCurrencies[i] != newCurrencies[i] {
					return fmt.Errorf("Account %s can not change currencies", a.account)
				}
			}
		}

		acc.opened = append(acc.opened, a.Date())
		ls.accounts[a.account] = acc
	}
	return nil
}

// AccountClose closes the account
type AccountClose struct {
	directive
	account AccountName
}

// Apply close account to the LedgerState
func (a AccountClose) Apply(ls *LedgerState) error {
	acc, ok := ls.accounts[a.account]
	if !ok {
		return fmt.Errorf("Account %s was not opened", a.account)
	}
	if acc.IsClosed(a.Date()) {
		return fmt.Errorf("Account %s is already closed", a.account)
	}
	acc.closed = append(acc.closed, a.Date())
	ls.accounts[a.account] = acc
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
		directive: directive{
			date:     date,
			lineNum:  line.lineNum,
			fileName: fileName,
			order:    accountOpenOrder,
		},
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
		directive: directive{
			date:     date,
			lineNum:  line.lineNum,
			fileName: fileName,
			order:    accountCloseOrder,
		},
		account: AccountName(accountName),
	}
	return d, nil
}
