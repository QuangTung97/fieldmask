package fields

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func scanStringTest(input string) ([]tokenType, []string) {
	s := newScanner(input)
	var tokens []tokenType
	var idents []string
	for s.next() {
		tokens = append(tokens, s.getTokenType())
		if s.getTokenType() == tokenTypeIdent {
			idents = append(idents, s.getIdentString())
		}
	}
	return tokens, idents
}

func TestScanner(t *testing.T) {
	t.Run("only indent", func(t *testing.T) {
		s := newScanner("sku")

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "sku", s.getIdentString())

		assert.Equal(t, false, s.next())
		assert.Equal(t, tokenTypeUnspecified, s.getTokenType())
		assert.Equal(t, nil, s.getErr())
	})

	t.Run("ident and dot", func(t *testing.T) {
		s := newScanner("provider.name")

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "provider", s.getIdentString())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeDot, s.getTokenType())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "name", s.getIdentString())

		assert.Equal(t, false, s.next())
		assert.Equal(t, tokenTypeUnspecified, s.getTokenType())
	})

	t.Run("ident and dot and bracket", func(t *testing.T) {
		s := newScanner("provider.{id|name}")

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "provider", s.getIdentString())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeDot, s.getTokenType())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeOpeningBracket, s.getTokenType())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "id", s.getIdentString())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeVerticalLine, s.getTokenType())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "name", s.getIdentString())

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeClosingBracket, s.getTokenType())

		assert.Equal(t, false, s.next())
	})

	t.Run("multiple bracket", func(t *testing.T) {
		tokens, idents := scanStringTest("info.{sku|name|seller.{id|code}}")
		assert.Equal(t, []tokenType{
			tokenTypeIdent,
			tokenTypeDot,
			tokenTypeOpeningBracket,
			tokenTypeIdent,
			tokenTypeVerticalLine,
			tokenTypeIdent,
			tokenTypeVerticalLine,
			tokenTypeIdent,
			tokenTypeDot,
			tokenTypeOpeningBracket,
			tokenTypeIdent,
			tokenTypeVerticalLine,
			tokenTypeIdent,
			tokenTypeClosingBracket,
			tokenTypeClosingBracket,
		}, tokens)
		assert.Equal(t, []string{"info", "sku", "name", "seller", "id", "code"}, idents)
	})

	t.Run("with space", func(t *testing.T) {
		s := newScanner("sku name")

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "sku", s.getIdentString())

		assert.Equal(t, false, s.next())
		assert.Equal(t, tokenTypeUnspecified, s.getTokenType())
		assert.Equal(t, errors.New("fields: not allow spaces"), s.getErr())
	})

	t.Run("with comma", func(t *testing.T) {
		s := newScanner("sku,name")

		assert.Equal(t, true, s.next())
		assert.Equal(t, tokenTypeIdent, s.getTokenType())
		assert.Equal(t, "sku", s.getIdentString())

		assert.Equal(t, false, s.next())
		assert.Equal(t, tokenTypeUnspecified, s.getTokenType())
		assert.Equal(t, errors.New("fields: character ',' is not allowed"), s.getErr())
	})
}
