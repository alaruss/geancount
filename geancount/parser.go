package geancount

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Token is a minimal part of input
type Token struct {
	text      string
	isQuoted  bool
	isComment bool
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
		if t.isComment {
			sb.WriteString(" / ")
		} else {
			sb.WriteString(" | ")
		}
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
		token := Token{
			text:      p.sb.String(),
			isQuoted:  p.isQuoted,
			isComment: p.inComment,
		}
		p.tokens = append(p.tokens, token)
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
