package geancount

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Price set the price of a currency on the date
type Price struct {
	directive
	currency Currency
	amount   Amount
}

type PricePoint struct {
	date   time.Time
	amount Amount
}

// Apply adds price to the inventory
func (p Price) Apply(ls *LedgerState) error {
	if _, ok := ls.prices[p.currency]; !ok {
		ls.prices[p.currency] = []PricePoint{}
	}
	ls.prices[p.currency] = append(ls.prices[p.currency], PricePoint{date: p.date, amount: p.amount})
	return nil
}

func newPriceFromTransaction(t Transaction) ([]Price, error) {
	prices := []Price{}
	for _, posting := range t.postings {
		if posting.price != nil {
			prices = append(prices, Price{
				directive: directive{
					date:     t.Date(),
					lineNum:  t.lineNum,
					fileName: t.fileName,
					order:    priceOrder,
				},
				currency: posting.amount.currency,
				amount:   Amount{value: *&posting.price.value, currency: *&posting.price.currency},
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
			order:    priceOrder,
		},
		currency: currency,
		amount:   amount,
	}
	return d, nil
}
