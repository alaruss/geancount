package geancount

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rs/zerolog/log"
)

// Token is a minimal part of input
type Token struct {
	text     string
	isQuoted bool
}

// Line is collection of tokens
type Line struct {
	lineNum    int
	isIndented bool
	tokens     []Token
}

// IsBlank returns true if there is not tokens in the line
func (l Line) IsBlank() bool {
	return len(l.tokens) == 0
}

// String returns a representation of Line used mostly for debug
func (l Line) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%03d: ", l.lineNum))
	if l.isIndented {
		sb.WriteRune('\t')
	}
	for _, t := range l.tokens {
		sb.WriteString(" | ")
		if t.isQuoted {
			sb.WriteRune('"')
		}
		sb.WriteString(t.text)
		if t.isQuoted {
			sb.WriteRune('"')
		}
	}
	return sb.String()
}

// parseState is a helper to store current state of the parser
type parserState struct {
	// token
	sb       strings.Builder
	inQuote  bool
	isQuoted bool
	prev     rune

	// line
	cursorLineNum int // where cursor is
	lineNum       int // where current line is started
	inComment     bool
	tokens        []Token
	isIndented    bool
}

func (p *parserState) addToken() {
	if p.sb.Len() > 0 {
		if !p.inComment {
			token := Token{
				text:     p.sb.String(),
				isQuoted: p.isQuoted,
			}
			p.tokens = append(p.tokens, token)
		}
		p.sb.Reset()
	}
	p.inQuote = false
	p.isQuoted = false
}

func (p *parserState) createLine() Line {
	line := Line{lineNum: p.lineNum, tokens: p.tokens, isIndented: p.isIndented}
	p.tokens = []Token{}
	p.isIndented = false
	p.inComment = false
	return line
}

func parseInput(r io.Reader) ([]Line, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanRunes)
	lines := []Line{}
	state := parserState{lineNum: 1, cursorLineNum: 1}
	for scanner.Scan() {
		r, _ := utf8.DecodeRune(scanner.Bytes())
		switch r {
		case '\n':
			state.cursorLineNum++
			if state.inQuote {
				state.sb.WriteRune(r)
			} else {
				state.addToken()
				lines = append(lines, state.createLine())
				state.lineNum = state.cursorLineNum
			}
		case ' ', '\t':
			if state.inQuote {
				state.sb.WriteRune(r)
				break
			} else {
				if len(state.tokens) == 0 && state.sb.Len() == 0 && !state.inComment {
					state.isIndented = true
				} else {
					state.addToken()
				}
			}
		case '"':
			// TODO: Deal with unbalanced quotes
			if state.inQuote {
				if state.prev == '\\' {
					s := state.sb.String()
					// Remove previos \ from the string
					state.sb.Reset()
					state.sb.WriteString(s[:len(s)-1])
					state.sb.WriteRune(r)
				} else {
					state.isQuoted = true
					state.addToken()
				}
			} else {
				state.inQuote = true
			}
		case ';':
			if !state.inComment {
				state.addToken()
				state.inComment = true
			} else {
				state.sb.WriteRune(r)
			}
		default:
			state.sb.WriteRune(r)
		}
		state.prev = r
	}
	// if there is not EOL in the end add last line
	if state.sb.Len() > 0 {
		state.addToken()
	}
	if len(state.tokens) > 0 {
		lines = append(lines, state.createLine())
		state.lineNum = state.cursorLineNum + 1
	}
	// Ensure the last line is blank
	if len(lines) > 0 && !lines[len(lines)-1].IsBlank() {
		lines = append(lines, Line{lineNum: state.lineNum})
	}
	return lines, nil
}

// LineGroup is collection of lines that form one directive
type LineGroup struct {
	lines []Line
}

func groupLines(lines []Line) ([]LineGroup, error) {
	lineGroups := []LineGroup{}
	lg := LineGroup{}
	for _, line := range lines {
		if line.IsBlank() {
			if len(lg.lines) > 0 {
				lineGroups = append(lineGroups, lg)
				lg = LineGroup{}
			}
		} else if line.isIndented {
			lg.lines = append(lg.lines, line)
		} else {
			if len(lg.lines) > 0 {
				lineGroups = append(lineGroups, lg)
				lg = LineGroup{}
			}
			lg.lines = append(lg.lines, line)
		}
	}
	if len(lg.lines) > 0 {
		lineGroups = append(lineGroups, lg)
	}
	return lineGroups, nil
}

func (l *Ledger) createDirectives(lineGroups []LineGroup) error {
	directives := []Directive{}
	for _, lg := range lineGroups {

		switch lg.lines[0].tokens[0].text {
		case "include", "pushtag", "poptag": // TODO implement
			continue
		case "option":
			err := l.applyOption(lg)
			if err == ErrNotDirective {
				continue
			} else if err != nil {
				return err
			}
		default:
			var err error
			var directive Directive
			switch lg.lines[0].tokens[1].text {
			case "open":
				directive, err = newAccountOpen(lg)
			case "close":
				directive, err = newAccountClose(lg)
			case "balance":
				directive, err = newBalance(lg)
			case "*", "!", "txn":
				directive, err = newTransaction(lg)
			}
			if err == ErrNotDirective { // just ignore
				continue
			} else if err != nil {
				return err
			}
			directives = append(directives, directive)
		}
	}
	slices.SortFunc(directives, func(i, j Directive) int {
		if i.Date().Before(j.Date()) {
			return -1
		} else if i.Date().After(j.Date()) {
			return 1
		}
		return 0
	})
	l.directives = directives
	return nil
}

func (l *Ledger) applyOption(lg LineGroup) error {
	line := lg.lines[0]
	if len(line.tokens) < 2 {
		return ErrNotDirective
	}
	switch optionName := line.tokens[1].text; optionName {
	case "operating_currency":
		if len(line.tokens) < 3 {
			log.Info().Int("line", line.lineNum).Msg("operating_currency has no currency")
			return ErrNotDirective
		}
		l.operatingCurrencies = append(l.operatingCurrencies, Currency(line.tokens[2].text))
	default:
		log.Info().Int("line", line.lineNum).Msgf("Unknown option %s", optionName)
	}
	return nil
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
