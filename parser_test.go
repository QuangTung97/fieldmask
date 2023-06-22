package fieldmask

import (
	"github.com/QuangTung97/fieldmask/testdata/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParser_Simple_Message(t *testing.T) {
	info := parseMessage(&pb.ProviderInfo{})
	assert.Equal(t, "ProviderInfo", info.typeName)
	assert.Equal(t, "github.com/QuangTung97/fieldmask/testdata/pb", info.importPath)

	assert.Equal(t, []objectField{
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
	}, info.subFields)
}

func TestParser_Complex_Object(t *testing.T) {
	info := parseMessage(&pb.Product{})

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

	assert.Equal(t, []objectField{
		{
			name:      "Sku",
			jsonName:  "sku",
			fieldType: fieldTypeSimple,
		},
		{
			name:      "Provider",
			jsonName:  "provider",
			fieldType: fieldTypeObject,
			info: &objectInfo{
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
			},
		},
		{
			name:      "Attributes",
			jsonName:  "attributes",
			fieldType: fieldTypeArrayOfObjects,
			info:      attribute,
		},
	}, info.subFields)
}

func TestParser_Invalid_Type(t *testing.T) {
	assert.PanicsWithValue(t, "invalid message type", func() {
		parseMessage(nil)
	})
}
