package fields

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComputeFieldInfos(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		infos, err := ComputeFieldInfos(nil)
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(infos))
	})

	t.Run("single", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{"sku"})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{
				FieldName: "sku",
			},
		}, infos)
	})

	t.Run("two fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{"sku", "name"})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{
				FieldName: "sku",
			},
			{
				FieldName: "name",
			},
		}, infos)
	})

	t.Run("with sub fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{"sku", "provider.id"})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{
				FieldName: "sku",
			},
			{
				FieldName: "provider",
				SubFields: []FieldInfo{
					{
						FieldName: "id",
					},
				},
			},
		}, infos)
	})

	t.Run("complex", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"provider.id",
			"provider.name",
			"seller.name",
			"seller.logo",
			"seller.attr.code",
			"seller.attr.name",
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{
				FieldName: "sku",
			},
			{
				FieldName: "provider",
				SubFields: []FieldInfo{
					{
						FieldName: "id",
					},
					{
						FieldName: "name",
					},
				},
			},
			{
				FieldName: "seller",
				SubFields: []FieldInfo{
					{
						FieldName: "name",
					},
					{
						FieldName: "logo",
					},
					{
						FieldName: "attr",
						SubFields: []FieldInfo{
							{
								FieldName: "code",
							},
							{
								FieldName: "name",
							},
						},
					},
				},
			},
		}, infos)
	})

	t.Run("duplicated fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{"sku", "name", "sku"})
		assert.Equal(t, ErrDuplicatedField("sku"), err)
		assert.Nil(t, infos)
	})

	t.Run("duplicated fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku", "name",
			"provider",
			"provider.name",
		})
		assert.Equal(t, ErrDuplicatedField("provider"), err)
		assert.Nil(t, infos)
	})

	t.Run("duplicated fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku", "name",
			"provider.name",
			"provider",
		})
		assert.Equal(t, ErrDuplicatedField("provider"), err)
		assert.Nil(t, infos)
	})

	t.Run("duplicated sub fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku", "name",
			"provider.name",
			"provider.id",
			"provider.name",
		})
		assert.Equal(t, ErrDuplicatedField("provider.name"), err)
		assert.Equal(t, "fieldmask: duplicated field 'provider.name'", err.Error())
		assert.Nil(t, infos)
	})

	t.Run("too much field", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"name",
			"provider.name",
			"provider.id",
		}, WithMaxFields(4))
		assert.Equal(t, errors.New("fieldmask: exceeded max number of fields"), err)
		assert.Equal(t, 0, len(infos))
	})

	t.Run("near too much field", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"name",
			"provider.name",
			"provider.id",
		}, WithMaxFields(5))
		assert.Equal(t, nil, err)
		assert.Equal(t, 3, len(infos))
	})

	t.Run("too much field depth", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"name",
			"provider.name.code.value",
			"provider.id",
			"provider.logo",
			"provider.imageUrl",
		}, WithMaxFieldDepth(3))
		assert.Equal(t, errors.New("fieldmask: exceeded max number of field depth"), err)
		assert.Equal(t, 0, len(infos))
	})

	t.Run("near too much field depth", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"name",
			"provider.name.code.value",
			"provider.id",
			"provider.logo",
			"provider.imageUrl",
		}, WithMaxFieldDepth(4))
		assert.Equal(t, nil, err)
		assert.Equal(t, 3, len(infos))
	})

	t.Run("too much component length", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"name",
			"helloabcd",
		}, WithMaxFieldComponentLength(8))
		assert.Equal(t, errors.New("fieldmask: exceeded length of field components"), err)
		assert.Nil(t, infos)
	})

	t.Run("near much component length", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"name",
			"helloabc",
		}, WithMaxFieldComponentLength(8))
		assert.Equal(t, nil, err)
		assert.Equal(t, 3, len(infos))
	})

	t.Run("with brackets", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"sku",
			"provider.{id|logo|imageUrl}",
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{FieldName: "sku"},
			{
				FieldName: "provider",
				SubFields: []FieldInfo{
					{FieldName: "id"},
					{FieldName: "logo"},
					{FieldName: "imageUrl"},
				},
			},
		}, infos)
	})

	t.Run("with brackets and another full fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{
			"provider.{id|logo|imageUrl}",
			"seller",
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{
				FieldName: "provider",
				SubFields: []FieldInfo{
					{FieldName: "id"},
					{FieldName: "logo"},
					{FieldName: "imageUrl"},
				},
			},
			{FieldName: "seller"},
		}, infos)
	})
}

func TestComputeFieldInfos_WithLimitedToFields(t *testing.T) {
	t.Run("error when not in limited fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos(
			[]string{"sku"},
			WithLimitedToFields([]string{"name"}),
		)
		assert.Equal(t, ErrFieldNotFound("sku"), err)
		assert.Equal(t, "fieldmask: field not found or not allowed 'sku'", err.Error())
		assert.Equal(t, 0, len(infos))
	})

	t.Run("error multi levels", func(t *testing.T) {
		infos, err := ComputeFieldInfos(
			[]string{"sku", "seller.{id|code|name}"},
			WithLimitedToFields([]string{"sku", "seller.{id|code}"}),
		)
		assert.Equal(t, ErrFieldNotFound("seller.name"), err)
		assert.Equal(t, 0, len(infos))
	})

	t.Run("success multi levels", func(t *testing.T) {
		infos, err := ComputeFieldInfos(
			[]string{"sku", "seller.{id|code|name}"},
			WithLimitedToFields([]string{"sku", "seller.{id|code|name|attr}"}),
		)
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{FieldName: "sku"},
			{
				FieldName: "seller",
				SubFields: []FieldInfo{
					{FieldName: "id"},
					{FieldName: "code"},
					{FieldName: "name"},
				},
			},
		}, infos)
	})

	t.Run("error multi levels", func(t *testing.T) {
		infos, err := ComputeFieldInfos(
			[]string{"sku", "seller.{id|code|name}"},
			WithLimitedToFields([]string{"sku", "seller"}),
		)
		assert.Equal(t, ErrFieldNotFound("seller.id"), err)
		assert.Equal(t, []FieldInfo(nil), infos)
	})

	t.Run("success multi levels, input field is more general", func(t *testing.T) {
		infos, err := ComputeFieldInfos(
			[]string{"sku", "seller"},
			WithLimitedToFields([]string{"sku", "seller.{id|code|name}"}),
		)
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{FieldName: "sku"},
			{FieldName: "seller"},
		}, infos)
	})

	t.Run("invalid allowed fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos(
			[]string{"sku", "seller.{id|code|name}"},
			WithLimitedToFields([]string{"sku", "seller.{id|code|name|code}"}),
		)
		assert.Equal(t, ErrDuplicatedField("seller.code"), err)
		assert.Equal(t, []FieldInfo(nil), infos)
	})
}
