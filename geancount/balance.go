package geancount

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Balance checks the amount of the account at the date
type Balance struct {
	directive
	account AccountName
	amount  Amount
}

// Pad inserts transaction to to make Balance assert
type Pad struct {
	directive
	account       AccountName
	sourceAccount AccountName
}

var defaultPrecision decimal.Decimal

func init() {
	defaultPrecision, _ = decimal.NewFromString("0.01")
}

// Apply check the if the calcualted balance is correct
func (b Balance) Apply(ls *LedgerState) error {
	accountBalance, ok := ls.balances[b.account]
	if !ok {
		return fmt.Errorf("Balance of unknown account %s", b.account)
	}
	calculated, ok := accountBalance[b.amount.currency]
	if !ok {
		calculated = decimal.Zero
	}
	acc := ls.accounts[b.account]
	diff := calculated.Sub(b.amount.value).Abs()
	if diff.GreaterThanOrEqual(defaultPrecision) {
		if acc.pad != nil {
			transation, err := acc.pad.createTransaction(b, calculated)
			if err != nil {
				return err
			}
			err = transation.Apply(ls)
			if err != nil {
				return err
			}
			if !ls.balances[b.account][b.amount.currency].Equal(b.amount.value) {
				return fmt.Errorf("Could not create pad transaction for %s", b.account)
			}
			acc.pad = nil
			ls.accounts[b.account] = acc
			return nil
		}
		return fmt.Errorf("Balance of %s expected %s but calcultated %s", b.account, b.amount.value, calculated)
	}
	return nil
}

func (p Pad) createTransaction(balance Balance, calculated decimal.Decimal) (Transaction, error) {
	amount := Amount{
		value:    balance.amount.value.Sub(calculated),
		currency: balance.amount.currency,
	}
	d := Transaction{
		directive: directive{date: p.Date(), lineNum: p.LineNum(), fileName: p.FileName()},
		status:    "P",
		narration: fmt.Sprintf("Padding inserted for Balance of %s for difference %s", balance.amount, amount),
		postings: []Posting{
			{account: p.account, amount: amount},
			{account: p.sourceAccount, amount: amount.Negative()},
		},
	}
	return d, nil
}

// Apply attach pad to the account
func (p Pad) Apply(ls *LedgerState) error {
	acc, ok := ls.accounts[p.account]
	if !ok {
		return fmt.Errorf("Padding of unknow account %s", p.account)
	}
	if acc.IsClosed(p.Date()) {
		return fmt.Errorf("Account %s is closed", acc)
	}
	sourceAcc := ls.accounts[p.sourceAccount]
	if !ok {
		return fmt.Errorf("Padding of unknow account %s", p.sourceAccount)
	}
	if sourceAcc.IsClosed(p.Date()) {
		return fmt.Errorf("Account %s is closed", sourceAcc)
	}
	var err error = nil
	if acc.pad != nil {
		err = fmt.Errorf("Unused Pad entry")
	}
	// Attach pad to the account anyway
	acc.pad = &p
	ls.accounts[p.account] = acc
	return err
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
		directive: directive{
			date:     date,
			lineNum:  line.lineNum,
			fileName: fileName,
			order:    balanceOrder,
		},
		account: AccountName(accountName),
		amount:  amount,
	}
	return d, nil
}

func newPad(lg LineGroup, fileName string) (Pad, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return Pad{}, ErrNotDirective
	}
	if len(line.tokens) > 4 {
		return Pad{}, fmt.Errorf("more tokens than expected")
	}
	accountName := line.tokens[2].text
	sourceAccountName := line.tokens[3].text
	d := Pad{
		directive: directive{
			date:     date,
			lineNum:  line.lineNum,
			fileName: fileName,
			order:    padOrder,
		},
		account:       AccountName(accountName),
		sourceAccount: AccountName(sourceAccountName),
	}
	return d, nil
}
