package fieldmask

import (
	"github.com/QuangTung97/fieldmask/testdata/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParser_Simple_Message(t *testing.T) {
	infos := parseMessages(&pb.ProviderInfo{})
	assert.Equal(t, 1, len(infos))

	info := infos[0]

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
	infos := parseMessages(&pb.Product{})
	assert.Equal(t, 1, len(infos))

	info := infos[0]

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
		{
			name:      "SellerIds",
			jsonName:  "sellerIds",
			fieldType: fieldTypeArrayOfPrimitives,
		},
		{
			name:      "BrandCodes",
			jsonName:  "brandCodes",
			fieldType: fieldTypeArrayOfPrimitives,
		},
		{
			name:      "CreatedAt",
			jsonName:  "createdAt",
			fieldType: fieldTypeSpecialField,
		},
		{
			name:      "Quantity",
			jsonName:  "quantity",
			fieldType: fieldTypeSpecialField,
		},
		{
			name:      "Stocks",
			jsonName:  "stocks",
			fieldType: fieldTypeSpecialField,
		},
	}, info.subFields)
}

func TestParser_Invalid_Type(t *testing.T) {
	assert.PanicsWithValue(t, "invalid message type", func() {
		parseMessages(nil)
	})
}

func TestParser_Multiple_Objects(t *testing.T) {
	infos := parseMessages(&pb.ProviderInfo{}, &pb.Product{})
	assert.Equal(t, 2, len(infos))

	assert.Same(t, infos[0], infos[1].subFields[1].info)
}
