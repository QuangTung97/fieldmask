package fieldmap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type field int

type simpleData struct {
	Root field

	Sku      field `json:"sku"`
	Name     field `json:"name"`
	ImageURL field `json:"imageUrl"`
}

func (d simpleData) GetRoot() field { return d.Root }

type sellerAttr struct {
	Root field

	Code field `json:"code"`
	Name field `json:"name"`
}

type sellerData struct {
	Root field

	ID   field      `json:"id"`
	Name field      `json:"name"`
	Logo field      `json:"logo"`
	Attr sellerAttr `json:"attr"`
}

type productData struct {
	Root     field
	Sku      field      `json:"sku"`
	Name     field      `json:"name"`
	Seller   sellerData `json:"seller"`
	ImageURL field      `json:"imageUrl"`
}

func (d productData) GetRoot() field {
	return d.Root
}

func TestFieldMap__GetMapping(t *testing.T) {
	t.Run("simple-struct", func(t *testing.T) {
		fm := New[field, simpleData]()

		p := fm.GetMapping()

		assert.Equal(t, field(1), p.Root)
		assert.Equal(t, field(2), p.Sku)
		assert.Equal(t, field(3), p.Name)
		assert.Equal(t, field(4), p.ImageURL)
	})

	t.Run("complex-struct", func(t *testing.T) {
		fm := New[field, productData]()

		p := fm.GetMapping()

		assert.Equal(t, field(1), p.Root)
		assert.Equal(t, field(2), p.Sku)
		assert.Equal(t, field(3), p.Name)
		assert.Equal(t, field(4), p.Seller.Root)
		assert.Equal(t, field(5), p.Seller.ID)
		assert.Equal(t, field(6), p.Seller.Name)
		assert.Equal(t, field(7), p.Seller.Logo)
		assert.Equal(t, field(8), p.Seller.Attr.Root)
		assert.Equal(t, field(9), p.Seller.Attr.Code)
		assert.Equal(t, field(10), p.Seller.Attr.Name)
		assert.Equal(t, field(11), p.ImageURL)

		assert.Equal(t, false, fm.IsStruct(p.Sku))
		assert.Equal(t, true, fm.IsStruct(p.Seller.Root))
		assert.Equal(t, false, fm.IsStruct(p.Seller.ID))

		assert.Equal(t, []field{5, 6, 7, 8}, fm.ChildrenOf(p.Seller.Root))

		assert.Equal(t, p.Seller.Root, fm.ParentOf(p.Seller.ID))
		assert.Equal(t, p.Seller.Root, fm.ParentOf(p.Seller.Name))

		assert.Equal(t, field(1), fm.ParentOf(p.Seller.Root))

		assert.Equal(t, field(1), fm.ParentOf(p.Sku))
		assert.Equal(t, field(1), fm.ParentOf(p.Name))

		assert.Equal(t, []field{p.Seller.Name, p.Seller.Root, p.Root}, fm.AncestorOf(p.Seller.Name))
	})

	t.Run("simple struct get field name", func(t *testing.T) {
		fm := New[field, simpleData]()

		p := fm.GetMapping()

		assert.Equal(t, "Sku", fm.GetFieldName(p.Sku))
		assert.Equal(t, "Name", fm.GetFieldName(p.Name))
		assert.Equal(t, "ImageURL", fm.GetFieldName(p.ImageURL))

		assert.Equal(t, "ImageURL", fm.GetFullFieldName(p.ImageURL))
	})

	t.Run("complex struct get field name", func(t *testing.T) {
		fm := New[field, productData]()

		p := fm.GetMapping()

		assert.Equal(t, "Sku", fm.GetFieldName(p.Sku))
		assert.Equal(t, "Name", fm.GetFieldName(p.Name))
		assert.Equal(t, "Seller", fm.GetFieldName(p.Seller.Root))
		assert.Equal(t, "ID", fm.GetFieldName(p.Seller.ID))
		assert.Equal(t, "Name", fm.GetFieldName(p.Seller.Attr.Name))

		assert.Equal(t, "Seller.ID", fm.GetFullFieldName(p.Seller.ID))
		assert.Equal(t, "Seller.Attr.Code", fm.GetFullFieldName(p.Seller.Attr.Code))
	})
}

type dataWithoutStructTagInner struct {
	Root field

	Name field
}

type dataWithoutStructTag struct {
	Root  field
	Sku   field                     `json:"sku"`
	Inner dataWithoutStructTagInner `json:"inner"`
}

func (d dataWithoutStructTag) GetRoot() field { return d.Root }

func TestFieldMap__StructTag(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		fm := New[field, simpleData](
			WithStructTags("json"),
		)

		p := fm.GetMapping()

		assert.Equal(t, "sku", fm.GetStructTag("json", p.Sku))
		assert.Equal(t, "name", fm.GetStructTag("json", p.Name))
		assert.Equal(t, "imageUrl", fm.GetStructTag("json", p.ImageURL))

		assert.Equal(t, "imageUrl", fm.GetFullStructTag("json", p.ImageURL))
	})

	t.Run("complex struct", func(t *testing.T) {
		fm := New[field, productData](
			WithStructTags("json"),
		)

		p := fm.GetMapping()

		assert.Equal(t, "sku", fm.GetStructTag("json", p.Sku))
		assert.Equal(t, "name", fm.GetStructTag("json", p.Name))
		assert.Equal(t, "imageUrl", fm.GetStructTag("json", p.ImageURL))

		assert.Equal(t, "seller", fm.GetStructTag("json", p.Seller.Root))
		assert.Equal(t, "id", fm.GetStructTag("json", p.Seller.ID))
		assert.Equal(t, "name", fm.GetStructTag("json", p.Seller.Name))

		assert.Equal(t, "seller.name", fm.GetFullStructTag("json", p.Seller.Name))
		assert.Equal(t, "seller.attr", fm.GetFullStructTag("json", p.Seller.Attr.Root))
		assert.Equal(t, "seller.attr.code", fm.GetFullStructTag("json", p.Seller.Attr.Code))

		assert.Equal(t, "code", fm.GetStructTag("json", p.Seller.Attr.Code))
	})

	t.Run("panics when type without struct tag", func(t *testing.T) {
		assert.PanicsWithValue(t, `missing struct tag "json" for field "Inner.Name"`, func() {
			_ = New[field, dataWithoutStructTag](
				WithStructTags("json"),
			)
		})
	})
}

type emptyTestStruct struct {
}

func (emptyTestStruct) GetRoot() field { return 0 }

type structInnerEmpty struct {
	Root field
	Attr emptyTestStruct
}

type structInnerMissingRoot struct {
	Attr emptyTestStruct
}

type structWithInnerEmpty struct {
	Root  field
	Inner structInnerEmpty
}

func (d structWithInnerEmpty) GetRoot() field { return d.Root }

type structWithInvalidGetRoot struct {
	Root field
}

func (structWithInvalidGetRoot) GetRoot() field {
	return 0
}

type structWithInnerMissingRoot struct {
	Root  field
	Inner structInnerMissingRoot
}

func (d structWithInnerMissingRoot) GetRoot() field { return d.Root }

type structWithSku struct {
	Root field
	Sku  string
}

type structWithInvalidType struct {
	Root  field
	Inner structWithSku
}

func (d structWithInvalidType) GetRoot() field { return d.Root }

type structWithInvalidRoot struct {
	Root string
}

type structWithInvalidRootType struct {
	Root  field
	Inner structWithInvalidRoot
}

func (d structWithInvalidRootType) GetRoot() field { return d.Root }

func TestFieldMap__Errors(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		assert.PanicsWithValue(t, `missing field "Root" for root of struct`, func() {
			New[field, emptyTestStruct]()
		})
	})

	t.Run("inner struct is empty", func(t *testing.T) {
		assert.PanicsWithValue(t, `missing field "Root" for field "Inner.Attr"`, func() {
			New[field, structWithInnerEmpty]()
		})
	})

	t.Run("invalid GetRoot implementation", func(t *testing.T) {
		assert.PanicsWithValue(t, `invalid GetRoot implementation`, func() {
			New[field, structWithInvalidGetRoot]()
		})
	})

	t.Run("inner struct without root", func(t *testing.T) {
		assert.PanicsWithValue(t, `missing field "Root" for field "Inner"`, func() {
			New[field, structWithInnerMissingRoot]()
		})
	})

	t.Run("invalid type for field sku", func(t *testing.T) {
		assert.PanicsWithValue(t, `invalid type for field "Inner.Sku"`, func() {
			New[field, structWithInvalidType]()
		})
	})

	t.Run("invalid type for field root", func(t *testing.T) {
		assert.PanicsWithValue(t, `invalid type for field "Inner.Root"`, func() {
			New[field, structWithInvalidRootType]()
		})
	})
}
