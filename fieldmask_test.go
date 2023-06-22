package fieldmask

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
				JsonName: "sku",
			},
		}, infos)
	})

	t.Run("two fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{"sku", "name"})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{
				JsonName: "sku",
			},
			{
				JsonName: "name",
			},
		}, infos)
	})

	t.Run("with sub fields", func(t *testing.T) {
		infos, err := ComputeFieldInfos([]string{"sku", "provider.id"})
		assert.Equal(t, nil, err)
		assert.Equal(t, []FieldInfo{
			{
				JsonName: "sku",
			},
			{
				JsonName: "provider",
				SubFields: []FieldInfo{
					{
						JsonName: "id",
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
				JsonName: "sku",
			},
			{
				JsonName: "provider",
				SubFields: []FieldInfo{
					{
						JsonName: "id",
					},
					{
						JsonName: "name",
					},
				},
			},
			{
				JsonName: "seller",
				SubFields: []FieldInfo{
					{
						JsonName: "name",
					},
					{
						JsonName: "logo",
					},
					{
						JsonName: "attr",
						SubFields: []FieldInfo{
							{
								JsonName: "code",
							},
							{
								JsonName: "name",
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
}
