package fieldmask

import (
	"bytes"
	_ "embed"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed testdata/generated/provider.go
var generatedCode string

func TestGenerate(t *testing.T) {
	var buf bytes.Buffer

	provider := &objectInfo{
		typeName:   "ProviderInfo",
		importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
		subFields: []objectField{
			{
				name:     "Id",
				jsonName: "id",
			},
			{
				name:     "Name",
				jsonName: "name",
			},
			{
				name:     "Logo",
				jsonName: "logo",
			},
			{
				name:     "ImageUrl",
				jsonName: "imageUrl",
			},
		},
	}

	option := &objectInfo{
		typeName:   "Option",
		importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
		subFields: []objectField{
			{
				name:     "Code",
				jsonName: "code",
			},
			{
				name:     "Name",
				jsonName: "name",
			},
		},
	}

	attribute := &objectInfo{
		typeName:   "Attribute",
		importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
		subFields: []objectField{
			{
				name:     "Id",
				jsonName: "id",
			},
			{
				name:     "Code",
				jsonName: "code",
			},
			{
				name:     "Name",
				jsonName: "name",
			},
			{
				name:      "Options",
				jsonName:  "options",
				fieldType: fieldTypeArrayOfObjects,
				info:      option,
			},
		},
	}

	product := &objectInfo{
		typeName:   "Product",
		importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
		subFields: []objectField{
			{
				name:     "Sku",
				jsonName: "sku",
			},
			{
				name:      "Provider",
				jsonName:  "provider",
				fieldType: fieldTypeObject,
				info:      provider,
			},
			{
				name:      "Attributes",
				jsonName:  "attributes",
				fieldType: fieldTypeArrayOfObjects,
				info:      attribute,
			},
		},
	}

	generateCode(&buf, []*objectInfo{
		provider,
		product,
	}, "generated")

	assert.Equal(t, generatedCode, buf.String())
}
