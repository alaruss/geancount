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
			order:    3,
		},
		account: AccountName(accountName),
		amount:  amount,
	}
	return d, nil
}
