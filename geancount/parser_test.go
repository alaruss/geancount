package geancount

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test basic parsing
func TestGetLine(t *testing.T) {
	text := `2000-01-01 open Equity:Opening-Balances

2000-01-01 balance Assets:Bank          0 EUR

2000-01-02 * 
  Assets:Bank                        
  Income:Job -100.00 EUR`
	lines, err := parseInput(strings.NewReader(text))
	assert.Nil(t, err)
	expected := []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false, isComment: false},
			{text: "open", isQuoted: false, isComment: false},
			{text: "Equity:Opening-Balances", isQuoted: false, isComment: false},
		}},
		{lineNum: 2, isIndented: false, tokens: []Token{}},
		{lineNum: 3, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false, isComment: false},
			{text: "balance", isQuoted: false, isComment: false},
			{text: "Assets:Bank", isQuoted: false, isComment: false},
			{text: "0", isQuoted: false, isComment: false},
			{text: "EUR", isQuoted: false, isComment: false},
		}},
		{lineNum: 4, isIndented: false, tokens: []Token{}},
		{lineNum: 5, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false, isComment: false},
			{text: "*", isQuoted: false, isComment: false},
		}},
		{lineNum: 6, isIndented: true, tokens: []Token{
			{text: "Assets:Bank", isQuoted: false, isComment: false},
		}},
		{lineNum: 7, isIndented: true, tokens: []Token{
			{text: "Income:Job", isQuoted: false, isComment: false},
			{text: "-100.00", isQuoted: false, isComment: false},
			{text: "EUR", isQuoted: false, isComment: false},
		}},
		{lineNum: 8, isIndented: false},
	}
	assert.Equal(t, expected, lines)
	assert.True(t, lines[len(lines)-1].IsBlank())
}

func TestParseComment(t *testing.T) {
	text := `2000-01-01 open Equity:Opening-Balances;inline ;comment
; 2000-01-01 open Assets:Bank1
2000-01-01 open Assets:Bank2
	`
	lines, err := parseInput(strings.NewReader(text))
	assert.Nil(t, err)
	expected := []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false, isComment: false},
			{text: "open", isQuoted: false, isComment: false},
			{text: "Equity:Opening-Balances", isQuoted: false, isComment: false},
			{text: "inline", isQuoted: false, isComment: true},
			{text: ";comment", isQuoted: false, isComment: true},
		}},
		{lineNum: 2, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false, isComment: true},
			{text: "open", isQuoted: false, isComment: true},
			{text: "Assets:Bank1", isQuoted: false, isComment: true},
		}},
		{lineNum: 3, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false, isComment: false},
			{text: "open", isQuoted: false, isComment: false},
			{text: "Assets:Bank2", isQuoted: false, isComment: false},
		}},
		{lineNum: 4, isIndented: false},
	}
	assert.Equal(t, expected, lines)
}

func TestParseQuotes(t *testing.T) {
	lines, err := parseInput(strings.NewReader(`2000-01-02 * "Payee" Narration`))
	assert.Nil(t, err)
	expected := []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false, isComment: false},
			{text: "*", isQuoted: false, isComment: false},
			{text: "Payee", isQuoted: true, isComment: false},
			{text: "Narration", isQuoted: false, isComment: false},
		}},
		{lineNum: 2, isIndented: false},
	}
	assert.Equal(t, expected, lines)

	lines, err = parseInput(strings.NewReader(`2000-01-02 * "Payee" "Narration"`))
	assert.Nil(t, err)
	expected = []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false, isComment: false},
			{text: "*", isQuoted: false, isComment: false},
			{text: "Payee", isQuoted: true, isComment: false},
			{text: "Narration", isQuoted: true, isComment: false},
		}},
		{lineNum: 2, isIndented: false},
	}
	assert.Equal(t, expected, lines)

	lines, err = parseInput(strings.NewReader(`2000-01-02 * "Payee""Narration"`))
	assert.Nil(t, err)
	expected = []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false, isComment: false},
			{text: "*", isQuoted: false, isComment: false},
			{text: "Payee", isQuoted: true, isComment: false},
			{text: "Narration", isQuoted: true, isComment: false},
		}},
		{lineNum: 2, isIndented: false},
	}
	assert.Equal(t, expected, lines)

	lines, err = parseInput(strings.NewReader(`2000-01-02 * "Payee" "Narration \" with quote"`))
	assert.Nil(t, err)
	expected = []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false, isComment: false},
			{text: "*", isQuoted: false, isComment: false},
			{text: "Payee", isQuoted: true, isComment: false},
			{text: "Narration \" with quote", isQuoted: true, isComment: false},
		}},
		{lineNum: 2, isIndented: false},
	}
	assert.Equal(t, expected, lines)

	lines, err = parseInput(strings.NewReader(`2000-01-02 * "Payee" "Narration 
multi
multiline"
2000-01-01 open Assets:Bank1`))
	assert.Nil(t, err)
	expected = []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false, isComment: false},
			{text: "*", isQuoted: false, isComment: false},
			{text: "Payee", isQuoted: true, isComment: false},
			{text: "Narration \nmulti\nmultiline", isQuoted: true, isComment: false},
		}},
		{lineNum: 4, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false, isComment: false},
			{text: "open", isQuoted: false, isComment: false},
			{text: "Assets:Bank1", isQuoted: false, isComment: false},
		}},
		{lineNum: 5, isIndented: false},
	}
	assert.Equal(t, expected, lines)
}

func TestLineIsBlank(t *testing.T) {
	line := Line{}
	assert.True(t, line.IsBlank())

	line.tokens = []Token{}
	assert.True(t, line.IsBlank())

	line.tokens = append(line.tokens, Token{})
	assert.False(t, line.IsBlank())
}
