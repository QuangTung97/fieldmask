package fields

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type scanner struct {
	state tokenType
	data  []rune
	pos   int

	lastToken tokenType
	ident     []rune

	errChar rune
	err     error
}

type tokenType int

const (
	tokenTypeUnspecified tokenType = iota
	tokenTypeIdent
	tokenTypeDot
	tokenTypeOpeningBracket
	tokenTypeClosingBracket
	tokenTypeVerticalLine
)

func newScanner(s string) *scanner {
	data := make([]rune, 0, utf8.RuneCountInString(s)+1)
	for _, r := range s {
		data = append(data, r)
	}
	data = append(data, 0)
	return &scanner{
		state: tokenTypeUnspecified,
		data:  data,
		pos:   0,
	}
}

func isIdentChar(ch rune) bool {
	return unicode.IsDigit(ch) || unicode.IsLetter(ch)
}

func (s *scanner) handleNextChar(ch rune) (endOfToken bool, err error) {
	switch s.state {
	case tokenTypeUnspecified:
		switch ch {
		case '.':
			s.state = tokenTypeDot
		case '{':
			s.state = tokenTypeOpeningBracket
		case '}':
			s.state = tokenTypeClosingBracket
		case '|':
			s.state = tokenTypeVerticalLine

		default:
			if isIdentChar(ch) {
				s.state = tokenTypeIdent
				s.ident = s.ident[:0]
				s.ident = append(s.ident, ch)
				return false, nil
			}
			if ch == 0 {
				return false, nil
			}
			if ch == ' ' {
				return false, fmt.Errorf("fields: not allow spaces")
			}
			return false, fmt.Errorf("fields: character '%c' is not allowed", ch)
		}
		return false, nil

	case tokenTypeIdent:
		if isIdentChar(ch) {
			s.ident = append(s.ident, ch)
			return false, nil
		}
		return true, nil

	case tokenTypeDot, tokenTypeOpeningBracket, tokenTypeClosingBracket, tokenTypeVerticalLine:
		return true, nil

	default:
		panic("unreachable state")
	}
}

func (s *scanner) next() bool {
	for s.pos < len(s.data) {
		ch := s.data[s.pos]
		endOfToken, err := s.handleNextChar(ch)
		if err != nil {
			s.errChar = ch
			s.err = err
			s.lastToken = tokenTypeUnspecified
			return false
		}
		if endOfToken {
			s.lastToken = s.state
			s.state = tokenTypeUnspecified
			return true
		}
		s.pos++
	}
	s.lastToken = tokenTypeUnspecified
	return false
}

func (s *scanner) getTokenType() tokenType {
	return s.lastToken
}

func (s *scanner) getTokenString() string {
	switch s.getTokenType() {
	case tokenTypeDot:
		return "."
	case tokenTypeVerticalLine:
		return "|"
	case tokenTypeOpeningBracket:
		return "{"
	case tokenTypeClosingBracket:
		return "}"
	default:
		return ""
	}
}

func (s *scanner) getIdentString() string {
	return string(s.ident)
}

func (s *scanner) getErr() error {
	return s.err
}

func (s *scanner) withErrorf(format string, args ...any) error {
	if s.err != nil {
		return s.err
	}
	return fmt.Errorf("fields: "+format, args...)
}
