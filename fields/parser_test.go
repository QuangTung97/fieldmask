package fields

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newParserTest(input string) *parser {
	opts := newComputeOptions(nil)
	p := newParser(input, newCollector(opts))
	return p
}

func TestParser(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		p := newParserTest("sku")
		err := p.parse()
		assert.Equal(t, nil, err)

		assert.Equal(t, []FieldInfo{
			{FieldName: "sku"},
		}, p.collector.toFieldInfos())
	})

	t.Run("two level", func(t *testing.T) {
		p := newParserTest("info.sku")
		err := p.parse()
		assert.Equal(t, nil, err)

		assert.Equal(t, []FieldInfo{
			{
				FieldName: "info",
				SubFields: []FieldInfo{
					{FieldName: "sku"},
				},
			},
		}, p.collector.toFieldInfos())
	})

	t.Run("three level", func(t *testing.T) {
		p := newParserTest("info.seller.id")
		err := p.parse()
		assert.Equal(t, nil, err)

		assert.Equal(t, []FieldInfo{
			{
				FieldName: "info",
				SubFields: []FieldInfo{
					{
						FieldName: "seller",
						SubFields: []FieldInfo{
							{FieldName: "id"},
						},
					},
				},
			},
		}, p.collector.toFieldInfos())
	})

	t.Run("with multiple sibling", func(t *testing.T) {
		p := newParserTest("info.{sku|name|image}")
		err := p.parse()
		assert.Equal(t, nil, err)

		assert.Equal(t, []FieldInfo{
			{
				FieldName: "info",
				SubFields: []FieldInfo{
					{FieldName: "sku"},
					{FieldName: "name"},
					{FieldName: "image"},
				},
			},
		}, p.collector.toFieldInfos())
	})

	t.Run("sibling with dot", func(t *testing.T) {
		p := newParserTest("info.{sku|seller.id|seller.code|image}")
		err := p.parse()
		assert.Equal(t, nil, err)

		assert.Equal(t, []FieldInfo{
			{
				FieldName: "info",
				SubFields: []FieldInfo{
					{FieldName: "sku"},
					{
						FieldName: "seller",
						SubFields: []FieldInfo{
							{FieldName: "id"},
							{FieldName: "code"},
						},
					},
					{FieldName: "image"},
				},
			},
		}, p.collector.toFieldInfos())
	})

	t.Run("multiple levels of sibling, only one", func(t *testing.T) {
		p := newParserTest("info.{sku|seller.{id|code}}")
		err := p.parse()
		assert.Equal(t, nil, err)

		assert.Equal(t, []FieldInfo{
			{
				FieldName: "info",
				SubFields: []FieldInfo{
					{FieldName: "sku"},
					{
						FieldName: "seller",
						SubFields: []FieldInfo{
							{FieldName: "id"},
							{FieldName: "code"},
						},
					},
				},
			},
		}, p.collector.toFieldInfos())
	})

	t.Run("multiple levels of sibling", func(t *testing.T) {
		p := newParserTest("info.{sku|seller.{id|code}|image|provider.{logo|name}|shipping}")
		err := p.parse()
		assert.Equal(t, nil, err)

		assert.Equal(t, []FieldInfo{
			{
				FieldName: "info",
				SubFields: []FieldInfo{
					{FieldName: "sku"},
					{
						FieldName: "seller",
						SubFields: []FieldInfo{
							{FieldName: "id"},
							{FieldName: "code"},
						},
					},
					{FieldName: "image"},
					{
						FieldName: "provider",
						SubFields: []FieldInfo{
							{FieldName: "logo"},
							{FieldName: "name"},
						},
					},
					{FieldName: "shipping"},
				},
			},
		}, p.collector.toFieldInfos())
	})
}

func TestParser_Error(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		p := newParserTest("")
		err := p.parse()
		assert.Equal(t, errors.New("fields: missing field identifier"), err)
	})

	t.Run("first token is not ident", func(t *testing.T) {
		p := newParserTest(".")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier, instead found '.'"), err)
	})

	t.Run("first token is not ident, found '{'", func(t *testing.T) {
		p := newParserTest("{")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier, instead found '{'"), err)
	})

	t.Run("expect ident after opening bracket", func(t *testing.T) {
		p := newParserTest("info.{")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier after '{'"), err)
	})

	t.Run("expect closing bracket", func(t *testing.T) {
		p := newParserTest("info.{sku")
		err := p.parse()
		assert.Equal(t, errors.New("fields: missing '}'"), err)
	})

	t.Run("expect ident after vertical line", func(t *testing.T) {
		p := newParserTest("info.{sku|")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier after '|'"), err)
	})

	t.Run("expect ident or bracket after dot", func(t *testing.T) {
		p := newParserTest("info.")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier or a '{' after '.'"), err)
	})

	t.Run("expect ident or bracket after dot, found '|'", func(t *testing.T) {
		p := newParserTest("info.|")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier or a '{' after '.', instead found '|'"), err)
	})

	t.Run("expect ident or bracket after dot, found '}'", func(t *testing.T) {
		p := newParserTest("info.}")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier or a '{' after '.', instead found '}'"), err)
	})

	t.Run("duplicated inside sibling", func(t *testing.T) {
		p := newParserTest("info.{sku|name|sku}")
		err := p.parse()
		assert.Equal(t, ErrDuplicatedField("info.sku"), err)
	})

	t.Run("expect ident after dot inside bracket", func(t *testing.T) {
		p := newParserTest("info.{sku.")
		err := p.parse()
		assert.Equal(t, errors.New("fields: expecting an identifier or a '{' after '.'"), err)
	})

	t.Run("extra content", func(t *testing.T) {
		p := newParserTest("info.{sku|name}}")
		err := p.parse()
		assert.Equal(t, errors.New("fields: not allow extra string after '}'"), err)
	})
}

func TestParser_Error_From_Scanner(t *testing.T) {
	t.Run("question mark", func(t *testing.T) {
		p := newParserTest("?")
		err := p.parse()
		assert.Equal(t, errors.New("fields: character '?' is not allowed"), err)
	})
}
