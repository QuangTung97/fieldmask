package fieldmask

import (
	"github.com/QuangTung97/fieldmask/fields"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFieldSelector(t *testing.T) {
	t.Run("limited to all", func(t *testing.T) {
		s := newFieldSelector()

		info := &objectInfo{
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

		s.traverse(info, nil)

		assert.Equal(t, true, s.allowAll(info))

		// do keep
		s.keepSelectedFields(info)
		assert.Equal(t, 4, len(info.subFields))
	})

	t.Run("limited to some fields", func(t *testing.T) {
		s := newFieldSelector()

		info := &objectInfo{
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

		limitedTo := []fields.FieldInfo{
			{FieldName: "id"},
			{FieldName: "logo"},
		}
		s.traverse(info, limitedTo)

		assert.Equal(t, false, s.allowAll(info))
		assert.Equal(t, true, s.allowField(info, "id"))
		assert.Equal(t, true, s.allowField(info, "logo"))
		assert.Equal(t, false, s.allowField(info, "name"))

		// do keep
		s.keepSelectedFields(info)

		expected := &objectInfo{
			typeName:   "ProviderInfo",
			importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
			subFields: []objectField{
				{
					name:     "Id",
					jsonName: "id",
				},
				{
					name:     "Logo",
					jsonName: "logo",
				},
			},
		}
		assert.Equal(t, expected, info)
	})

	t.Run("limited to all in one, but limit to some fields in another", func(t *testing.T) {
		s := newFieldSelector()

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
					name:      "Quantity",
					jsonName:  "quantity",
					fieldType: fieldTypeSpecialField,
				},
				{
					name:      "Stocks",
					jsonName:  "stocks",
					fieldType: fieldTypeSpecialField,
				},
			},
		}

		s.traverse(provider, nil)
		limitedTo := []fields.FieldInfo{
			{
				FieldName: "sku",
			},
			{
				FieldName: "provider",
				SubFields: []fields.FieldInfo{
					{FieldName: "name"},
				},
			},
			{
				FieldName: "stocks",
			},
		}
		s.traverse(product, limitedTo)

		assert.Equal(t, false, s.allowAll(product))
		assert.Equal(t, false, s.allowAll(provider))

		assert.Equal(t, true, s.allowField(product, "sku"))
		assert.Equal(t, false, s.allowField(product, "quantity"))
		assert.Equal(t, true, s.allowField(product, "stocks"))

		assert.Equal(t, false, s.allowField(provider, "id"))
		assert.Equal(t, true, s.allowField(provider, "name"))

		// do keep
		s.keepSelectedFields(provider)
		s.keepSelectedFields(product)

		newProvider := &objectInfo{
			typeName:   "ProviderInfo",
			importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
			subFields: []objectField{
				{
					name:     "Name",
					jsonName: "name",
				},
			},
		}
		assert.Equal(t, newProvider, provider)

		newProduct := &objectInfo{
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
					info:      newProvider,
				},
				{
					name:      "Stocks",
					jsonName:  "stocks",
					fieldType: fieldTypeSpecialField,
				},
			},
		}

		assert.Equal(t, newProduct, product)
	})

	t.Run("limited to some fields in one, but limit to all fields in another", func(t *testing.T) {
		s := newFieldSelector()

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
					name:      "Quantity",
					jsonName:  "quantity",
					fieldType: fieldTypeSpecialField,
				},
				{
					name:      "Stocks",
					jsonName:  "stocks",
					fieldType: fieldTypeSpecialField,
				},
			},
		}

		s.traverse(product, []fields.FieldInfo{
			{
				FieldName: "sku",
			},
			{
				FieldName: "provider",
				SubFields: []fields.FieldInfo{
					{FieldName: "name"},
				},
			},
			{
				FieldName: "stocks",
			},
		})
		s.traverse(provider, nil)

		assert.Equal(t, false, s.allowAll(product))
		assert.Equal(t, false, s.allowAll(provider))

		assert.Equal(t, true, s.allowField(product, "sku"))
		assert.Equal(t, false, s.allowField(product, "quantity"))
		assert.Equal(t, true, s.allowField(product, "stocks"))

		assert.Equal(t, false, s.allowField(provider, "id"))
		assert.Equal(t, true, s.allowField(provider, "name"))
	})

	t.Run("not found field", func(t *testing.T) {
		s := newFieldSelector()

		info := &objectInfo{
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

		assert.PanicsWithValue(t, "not found field 'xxyy'", func() {
			s.traverse(info, []fields.FieldInfo{
				{
					FieldName: "xxyy",
				},
			})
		})
	})

	t.Run("not found field in nested", func(t *testing.T) {
		s := newFieldSelector()

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
					name:      "Attributes",
					jsonName:  "attributes",
					fieldType: fieldTypeArrayOfObjects,
					info:      attribute,
				},
			},
		}

		assert.PanicsWithValue(t, "not found field 'attributes.options.hello'", func() {
			s.traverse(product, []fields.FieldInfo{
				{
					FieldName: "attributes",
					SubFields: []fields.FieldInfo{
						{FieldName: "name"},
						{
							FieldName: "options",
							SubFields: []fields.FieldInfo{
								{FieldName: "code"},
								{FieldName: "hello"},
							},
						},
					},
				},
			})
		})
	})

	t.Run("outer allow all inner fields", func(t *testing.T) {
		s := newFieldSelector()

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
					name:      "Quantity",
					jsonName:  "quantity",
					fieldType: fieldTypeSpecialField,
				},
				{
					name:      "Stocks",
					jsonName:  "stocks",
					fieldType: fieldTypeSpecialField,
				},
			},
		}

		// do traverse
		s.traverse(provider, nil)
		limitedTo := []fields.FieldInfo{
			{FieldName: "sku"},
			{FieldName: "provider"},
			{FieldName: "stocks"},
		}
		s.traverse(product, limitedTo)

		// do keep
		s.keepSelectedFields(provider)
		s.keepSelectedFields(product)

		newProduct := &objectInfo{
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
					fieldType: fieldTypeSimple,
				},
				{
					name:      "Stocks",
					jsonName:  "stocks",
					fieldType: fieldTypeSpecialField,
				},
			},
		}

		assert.Equal(t, newProduct, product)
	})

	t.Run("multiple nested", func(t *testing.T) {
		s := newFieldSelector()

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
					name:      "Attributes",
					jsonName:  "attributes",
					fieldType: fieldTypeArrayOfObjects,
					info:      attribute,
				},
			},
		}

		s.traverse(product, []fields.FieldInfo{
			{
				FieldName: "attributes",
				SubFields: []fields.FieldInfo{
					{FieldName: "name"},
					{
						FieldName: "options",
						SubFields: []fields.FieldInfo{
							{FieldName: "code"},
						},
					},
				},
			},
		})

		// do keep
		s.keepSelectedFields(product)
		s.keepSelectedFields(attribute)
		s.keepSelectedFields(option)

		newOption := &objectInfo{
			typeName:   "Option",
			importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
			subFields: []objectField{
				{
					name:     "Code",
					jsonName: "code",
				},
			},
		}

		newAttr := &objectInfo{
			typeName:   "Attribute",
			importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
			subFields: []objectField{
				{
					name:     "Name",
					jsonName: "name",
				},
				{
					name:      "Options",
					jsonName:  "options",
					fieldType: fieldTypeArrayOfObjects,
					info:      newOption,
				},
			},
		}

		newProduct := &objectInfo{
			typeName:   "Product",
			importPath: "github.com/QuangTung97/fieldmask/testdata/pb",
			subFields: []objectField{
				{
					name:      "Attributes",
					jsonName:  "attributes",
					fieldType: fieldTypeArrayOfObjects,
					info:      newAttr,
				},
			},
		}

		assert.Equal(t, newProduct, product)
	})
}
