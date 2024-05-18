package geancount

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Dire
// Price set the price of a currency on the date
type Price struct {
	directive
	currency Currency
	amount   Amount
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
