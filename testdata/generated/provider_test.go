package generated

import (
	"github.com/QuangTung97/fieldmask/fields"
	"github.com/QuangTung97/fieldmask/testdata/pb"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviderFieldMask(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		fm, err := NewProviderInfoFieldMask([]string{"id", "name"})
		assert.Equal(t, nil, err)

		newInfo := fm.Mask(&pb.ProviderInfo{
			Id:       21,
			Name:     "Provider Name",
			Logo:     "Logo 01",
			ImageUrl: "image-url",
		})
		assert.Equal(t, &pb.ProviderInfo{
			Id:   21,
			Name: "Provider Name",
		}, newInfo)
	})

	t.Run("empty", func(t *testing.T) {
		fm, err := NewProviderInfoFieldMask(nil)
		assert.Equal(t, nil, err)

		newInfo := fm.Mask(&pb.ProviderInfo{
			Id:       21,
			Name:     "Provider Name",
			Logo:     "Logo 01",
			ImageUrl: "image-url",
		})
		assert.Equal(t, &pb.ProviderInfo{
			Id:       21,
			Name:     "Provider Name",
			Logo:     "Logo 01",
			ImageUrl: "image-url",
		}, newInfo)
	})

	t.Run("full fields", func(t *testing.T) {
		fm, err := NewProviderInfoFieldMask([]string{
			"id", "name", "logo", "imageUrl",
		})
		assert.Equal(t, nil, err)

		newInfo := fm.Mask(&pb.ProviderInfo{
			Id:       21,
			Name:     "Provider Name",
			Logo:     "Logo 01",
			ImageUrl: "image-url",
		})
		assert.Equal(t, &pb.ProviderInfo{
			Id:       21,
			Name:     "Provider Name",
			Logo:     "Logo 01",
			ImageUrl: "image-url",
		}, newInfo)
	})

	t.Run("invalid field", func(t *testing.T) {
		fm, err := NewProviderInfoFieldMask([]string{"id", "sku"})
		assert.Equal(t, fields.ErrFieldNotFound("sku"), err)
		assert.Nil(t, fm)
	})
}

func TestProductFieldMask(t *testing.T) {
	ts := types.TimestampNow()

	product := &pb.Product{
		Sku: "SKU01",
		Provider: &pb.ProviderInfo{
			Id:       21,
			Name:     "Provider Name",
			Logo:     "Provider Logo",
			ImageUrl: "provider-image-url",
		},
		Attributes: []*pb.Attribute{
			{
				Id:   31,
				Code: "ATTR01",
				Name: "Attr Name 01",
				Options: []*pb.Option{
					{
						Code: "OPTION01",
						Name: "Option Name 01",
					},
					{
						Code: "OPTION02",
						Name: "Option Name 02",
					},
				},
			},
			{
				Id:   32,
				Code: "ATTR02",
				Name: "Attr Name 02",
				Options: []*pb.Option{
					{
						Code: "OPTION03",
						Name: "Option Name 03",
					},
				},
			},
		},
		SellerIds:  []int32{51, 52},
		BrandCodes: []string{"BRAND01", "BRAND02"},
		CreatedAt:  ts,
		Quantity:   &types.DoubleValue{Value: 886},
		Stocks: []*types.Int32Value{
			nil,
			{Value: 228},
		},
	}

	t.Run("empty", func(t *testing.T) {
		fm, err := NewProductFieldMask(nil)
		assert.Equal(t, nil, err)

		assert.Equal(t, product, fm.Mask(product))
	})

	t.Run("only sku", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{"sku"})
		assert.Equal(t, nil, err)

		assert.Equal(t, &pb.Product{
			Sku: "SKU01",
		}, fm.Mask(product))
	})

	t.Run("sku and provider", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku", "provider",
		})
		assert.Equal(t, nil, err)

		assert.Equal(t, &pb.Product{
			Sku: "SKU01",
			Provider: &pb.ProviderInfo{
				Id:       21,
				Name:     "Provider Name",
				Logo:     "Provider Logo",
				ImageUrl: "provider-image-url",
			},
		}, fm.Mask(product))
	})

	t.Run("sku and provider sub fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku", "provider.id", "provider.name",
		})
		assert.Equal(t, nil, err)

		assert.Equal(t, &pb.Product{
			Sku: "SKU01",
			Provider: &pb.ProviderInfo{
				Id:   21,
				Name: "Provider Name",
			},
		}, fm.Mask(product))
	})

	t.Run("invalid provider field", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku", "provider.id", "provider.invalid",
		})
		assert.Equal(t, fields.ErrFieldNotFound("provider.invalid"), err)
		assert.Nil(t, fm)
	})

	t.Run("with attributes", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku", "attributes",
		})
		assert.Equal(t, nil, err)

		assert.Equal(t, &pb.Product{
			Sku: "SKU01",
			Attributes: []*pb.Attribute{
				{
					Id:   31,
					Code: "ATTR01",
					Name: "Attr Name 01",
					Options: []*pb.Option{
						{
							Code: "OPTION01",
							Name: "Option Name 01",
						},
						{
							Code: "OPTION02",
							Name: "Option Name 02",
						},
					},
				},
				{
					Id:   32,
					Code: "ATTR02",
					Name: "Attr Name 02",
					Options: []*pb.Option{
						{
							Code: "OPTION03",
							Name: "Option Name 03",
						},
					},
				},
			},
		}, fm.Mask(product))
	})

	t.Run("with attributes sub fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku", "attributes.id", "attributes.name",
		})
		assert.Equal(t, nil, err)

		assert.Equal(t, &pb.Product{
			Sku: "SKU01",
			Attributes: []*pb.Attribute{
				{
					Id:   31,
					Name: "Attr Name 01",
				},
				{
					Id:   32,
					Name: "Attr Name 02",
				},
			},
		}, fm.Mask(product))
	})

	t.Run("with attributes & options sub fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku", "attributes.id", "attributes.options.code",
		})
		assert.Equal(t, nil, err)

		assert.Equal(t, &pb.Product{
			Sku: "SKU01",
			Attributes: []*pb.Attribute{
				{
					Id: 31,
					Options: []*pb.Option{
						{
							Code: "OPTION01",
						},
						{
							Code: "OPTION02",
						},
					},
				},
				{
					Id: 32,
					Options: []*pb.Option{
						{
							Code: "OPTION03",
						},
					},
				},
			},
		}, fm.Mask(product))
	})

	t.Run("invalid option field", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku", "attributes.id", "attributes.options.invalid",
		})
		assert.Equal(t, "fieldmask: field not found 'attributes.options.invalid'", err.Error())
		assert.Nil(t, fm)
	})

	t.Run("with seller ids", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sellerIds",
		})
		assert.Equal(t, nil, err)

		assert.Equal(t, &pb.Product{
			SellerIds: []int32{51, 52},
		}, fm.Mask(product))
	})

	t.Run("with all fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku",
			"provider.id", "provider.name", "provider.logo", "provider.imageUrl",
			"attributes.id", "attributes.name", "attributes.code",
			"attributes.options.code", "attributes.options.name",
			"sellerIds", "brandCodes",
			"createdAt", "quantity", "stocks",
		})
		assert.Equal(t, nil, err)

		assert.Equal(t, product, fm.Mask(product))
	})

	t.Run("invalid sub fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku.invalid",
		})
		assert.Equal(t, fields.ErrFieldNotFound("sku.invalid"), err)
		assert.Nil(t, fm)
	})

	t.Run("invalid sub provider fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"provider.logo.invalid",
		})
		assert.Equal(t, fields.ErrFieldNotFound("provider.logo.invalid"), err)
		assert.Nil(t, fm)
	})

	t.Run("invalid sub attributes fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"attributes.name.invalid",
		})
		assert.Equal(t, fields.ErrFieldNotFound("attributes.name.invalid"), err)
		assert.Nil(t, fm)
	})

	t.Run("invalid sub option fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"attributes.options.code.invalid",
		})
		assert.Equal(t, fields.ErrFieldNotFound("attributes.options.code.invalid"), err)
		assert.Nil(t, fm)
	})

	t.Run("reach limit max fields", func(t *testing.T) {
		fm, err := NewProductFieldMask([]string{
			"sku",
			"provider.id",
			"provider.name",
		}, fields.WithMaxFields(3))
		assert.Equal(t, fields.ErrExceedMaxFields, err)
		assert.Nil(t, fm)
	})
}
