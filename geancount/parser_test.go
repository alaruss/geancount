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
			{text: "2000-01-01", isQuoted: false},
			{text: "open", isQuoted: false},
			{text: "Equity:Opening-Balances", isQuoted: false},
		}},
		{lineNum: 2, isIndented: false, tokens: []Token{}},
		{lineNum: 3, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false},
			{text: "balance", isQuoted: false},
			{text: "Assets:Bank", isQuoted: false},
			{text: "0", isQuoted: false},
			{text: "EUR", isQuoted: false},
		}},
		{lineNum: 4, isIndented: false, tokens: []Token{}},
		{lineNum: 5, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false},
			{text: "*", isQuoted: false},
		}},
		{lineNum: 6, isIndented: true, tokens: []Token{
			{text: "Assets:Bank", isQuoted: false},
		}},
		{lineNum: 7, isIndented: true, tokens: []Token{
			{text: "Income:Job", isQuoted: false},
			{text: "-100.00", isQuoted: false},
			{text: "EUR", isQuoted: false},
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
			{text: "2000-01-01", isQuoted: false},
			{text: "open", isQuoted: false},
			{text: "Equity:Opening-Balances", isQuoted: false},
		}},
		{lineNum: 2, isIndented: false, tokens: []Token{}},
		{lineNum: 3, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false},
			{text: "open", isQuoted: false},
			{text: "Assets:Bank2", isQuoted: false},
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
			{text: "2000-01-02", isQuoted: false},
			{text: "*", isQuoted: false},
			{text: "Payee", isQuoted: true},
			{text: "Narration", isQuoted: false},
		}},
		{lineNum: 2, isIndented: false},
	}
	assert.Equal(t, expected, lines)

	lines, err = parseInput(strings.NewReader(`2000-01-02 * "Payee" "Narration"`))
	assert.Nil(t, err)
	expected = []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false},
			{text: "*", isQuoted: false},
			{text: "Payee", isQuoted: true},
			{text: "Narration", isQuoted: true},
		}},
		{lineNum: 2, isIndented: false},
	}
	assert.Equal(t, expected, lines)

	lines, err = parseInput(strings.NewReader(`2000-01-02 * "Payee""Narration"`))
	assert.Nil(t, err)
	expected = []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false},
			{text: "*", isQuoted: false},
			{text: "Payee", isQuoted: true},
			{text: "Narration", isQuoted: true},
		}},
		{lineNum: 2, isIndented: false},
	}
	assert.Equal(t, expected, lines)

	lines, err = parseInput(strings.NewReader(`2000-01-02 * "Payee" "Narration \" with quote"`))
	assert.Nil(t, err)
	expected = []Line{
		{lineNum: 1, isIndented: false, tokens: []Token{
			{text: "2000-01-02", isQuoted: false},
			{text: "*", isQuoted: false},
			{text: "Payee", isQuoted: true},
			{text: "Narration \" with quote", isQuoted: true},
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
			{text: "2000-01-02", isQuoted: false},
			{text: "*", isQuoted: false},
			{text: "Payee", isQuoted: true},
			{text: "Narration \nmulti\nmultiline", isQuoted: true},
		}},
		{lineNum: 4, isIndented: false, tokens: []Token{
			{text: "2000-01-01", isQuoted: false},
			{text: "open", isQuoted: false},
			{text: "Assets:Bank1", isQuoted: false},
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

func TestGroupLines(t *testing.T) {
	lineGroups, err := groupLines([]Line{
		{lineNum: 1},
		{lineNum: 2, tokens: []Token{{text: "1"}}},
		{lineNum: 3, tokens: []Token{{text: "2"}}},
		{lineNum: 4, isIndented: true, tokens: []Token{{text: "3"}}},
		{lineNum: 5, isIndented: true, tokens: []Token{{text: "4"}}},
		{lineNum: 6},
		{lineNum: 7, tokens: []Token{{text: "5"}}},
		{lineNum: 8, isIndented: true, tokens: []Token{{text: "6"}}},
		{lineNum: 9, tokens: []Token{{text: "7"}}},
	})
	assert.Nil(t, err)
	expected := []LineGroup{
		{lines: []Line{
			{lineNum: 2, tokens: []Token{{text: "1"}}},
		}},
		{lines: []Line{
			{lineNum: 3, tokens: []Token{{text: "2"}}},
			{lineNum: 4, isIndented: true, tokens: []Token{{text: "3"}}},
			{lineNum: 5, isIndented: true, tokens: []Token{{text: "4"}}},
		}},
		{lines: []Line{
			{lineNum: 7, tokens: []Token{{text: "5"}}},
			{lineNum: 8, isIndented: true, tokens: []Token{{text: "6"}}},
		}},
		{lines: []Line{
			{lineNum: 9, tokens: []Token{{text: "7"}}},
		}},
	}
	assert.Equal(t, expected, lineGroups)
}
