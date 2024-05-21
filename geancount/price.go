package geancount

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Price set the price of a currency on the date
type Price struct {
	directive
	currency Currency
	amount   Amount
}

// Apply adds price to the inventory
func (p Price) Apply(ls *LedgerState) error {
	return nil
}

func newPriceFromTransaction(t Transaction) ([]Price, error) {
	prices := []Price{}
	for _, posting := range t.postings {
		if posting.amount.price != nil && posting.amount.priceCurrency != nil {
			prices = append(prices, Price{
				directive: directive{
					date:     t.Date(),
					lineNum:  t.lineNum,
					fileName: t.fileName,
				},
				currency: posting.amount.currency,
				amount:   Amount{value: *posting.amount.price, currency: *posting.amount.priceCurrency},
			})
		}
	}
	return prices, nil
}

func newPrice(lg LineGroup, fileName string) (Price, error) {
	line := lg.lines[0]
	date, err := parseDate(line.tokens[0].text)
	if err != nil {
		return Price{}, ErrNotDirective
	}
	if len(line.tokens) > 5 {
		return Price{}, fmt.Errorf("more tokens than expected")
	}
	currency := Currency(line.tokens[2].text)
	amountValue, err := decimal.NewFromString(strings.ReplaceAll(line.tokens[3].text, ",", ""))
	if err != nil {
		return Price{}, fmt.Errorf("can not parse amount value")
	}
	amount := Amount{value: amountValue, currency: Currency(line.tokens[4].text)}
	d := Price{
		directive: directive{
			date:     date,
			lineNum:  line.lineNum,
			fileName: fileName,
		},
		currency: currency,
		amount:   amount,
	}
	return d, nil
}
