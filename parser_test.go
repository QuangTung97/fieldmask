package fieldmask

import (
	"github.com/QuangTung97/fieldmask/testdata/pb"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParser_Simple_Message(t *testing.T) {
	infos := parseMessages(NewProtoMessage(&pb.ProviderInfo{}))
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
	infos := parseMessages(NewProtoMessage(&pb.Product{}))
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
		parseMessages(NewProtoMessage(nil))
	})
}

func TestParser_Multiple_Objects(t *testing.T) {
	infos := parseMessages(
		NewProtoMessage(&pb.ProviderInfo{}),
		NewProtoMessage(&pb.Product{}),
	)
	assert.Equal(t, 2, len(infos))

	assert.Same(t, infos[0], infos[1].subFields[1].info)
}

func TestParser_Complex_Object__With_Limited_Fields(t *testing.T) {
	limitedToFields := []string{
		"sku",
		"provider.name",
		"attributes.code",
		"attributes.options.name",
		"stocks",
	}

	infos := parseMessages(NewProtoMessageWithFields(&pb.Product{}, limitedToFields))
	assert.Equal(t, 1, len(infos))

	info := infos[0]

	option := &objectInfo{
		typeName:   "Option",
		importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
		subFields: []objectField{
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
				name:     "Code",
				jsonName: "code",
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
						name:     "Name",
						jsonName: "name",
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
			name:      "Stocks",
			jsonName:  "stocks",
			fieldType: fieldTypeSpecialField,
		},
	}, info.subFields)
}

func TestParser_Simple_Message__With_Limited_Fields__Not_Found_Field(t *testing.T) {
	assert.PanicsWithValue(t, "not found field 'xxyy'", func() {
		parseMessages(
			NewProtoMessageWithFields(&pb.ProviderInfo{}, []string{
				"id", "xxyy",
			}),
		)
	})
}

func TestParser_Complex_Object__With_Limited_Fields__Not_Found_Sub_Field(t *testing.T) {
	limitedToFields := []string{
		"sku",
		"provider.name",
		"attributes.code",
		"attributes.options.xxyy",
		"stocks",
	}

	assert.PanicsWithValue(t, "not found field 'attributes.options.xxyy'", func() {
		parseMessages(NewProtoMessageWithFields(&pb.Product{}, limitedToFields))
	})
}

func TestParser_Complex_Object__With_Only_Root_Field(t *testing.T) {
	limitedToFields := []string{
		"sku",
		"provider",
	}

	infos := parseMessages(NewProtoMessageWithFields(&pb.Product{}, limitedToFields))
	assert.Equal(t, 1, len(infos))

	info := infos[0]

	assert.Equal(t, []objectField{
		{
			name:      "Sku",
			jsonName:  "sku",
			fieldType: fieldTypeSimple,
		},
		{
			name:      "Provider",
			jsonName:  "provider",
			fieldType: fieldTypeSimple,
		},
	}, info.subFields)
}

func TestParser_Both_Simple_And_Complex__With_Limited_Fields(t *testing.T) {
	infos := parseMessages(
		NewProtoMessageWithFields(&pb.Product{}, []string{
			"sku",
			"provider",
		}),
		NewProtoMessage(&pb.ProviderInfo{}),
	)
	assert.Equal(t, 2, len(infos))

	assert.Equal(t, []objectField{
		{
			name:      "Sku",
			jsonName:  "sku",
			fieldType: fieldTypeSimple,
		},
		{
			name:      "Provider",
			jsonName:  "provider",
			fieldType: fieldTypeSimple,
		},
	}, infos[0].subFields)

	assert.Equal(t, []objectField{
		{
			name:      "Id",
			jsonName:  "id",
			fieldType: fieldTypeSimple,
		},
		{
			name:      "Name",
			jsonName:  "name",
			fieldType: fieldTypeSimple,
		},
		{
			name:      "Logo",
			jsonName:  "logo",
			fieldType: fieldTypeSimple,
		},
		{
			name:      "ImageUrl",
			jsonName:  "imageUrl",
			fieldType: fieldTypeSimple,
		},
	}, infos[1].subFields)
}

func TestParser_Both_Simple_And_Complex__With_Limited_Fields__Conflicted(t *testing.T) {
	infos := parseMessages(
		NewProtoMessageWithFields(&pb.Product{}, []string{
			"sku",
			"provider.id",
		}),
		NewProtoMessageWithFields(&pb.ProviderInfo{}, []string{"name"}),
	)
	assert.Equal(t, 2, len(infos))

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
		},
	}

	assert.Equal(t, &objectInfo{
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
		},
	}, infos[0])
	assert.Equal(t, provider, infos[1])
}

func TestParser_Special_Type(t *testing.T) {
	assert.PanicsWithValue(t, "not allow type 'DoubleValue'", func() {
		parseMessages(NewProtoMessage(&types.DoubleValue{}))
	})
}
