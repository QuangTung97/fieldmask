package fieldmap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type sourceField int
type destField int

type sourceDataSimple struct {
	Root sourceField
	Sku  sourceField
	Name sourceField
	Body sourceField
}

func (d sourceDataSimple) GetRoot() sourceField {
	return d.Root
}

type destDataSimple struct {
	Root   destField
	Info   destField
	Detail destField
}

func (d destDataSimple) GetRoot() destField {
	return d.Root
}

func TestMapping_Simple_Structs(t *testing.T) {
	sourceFm := New[sourceField, sourceDataSimple]()
	destFm := New[destField, destDataSimple]()

	source := sourceFm.GetMapping()
	dest := destFm.GetMapping()

	m := NewMapper(
		sourceFm, destFm,
		WithSimpleMapping(sourceFm, destFm,
			NewMapping(source.Sku, dest.Info),
			NewMapping(source.Name, dest.Info),
			NewMapping(source.Body, dest.Detail),
		),
	)

	assert.Equal(t, 0, len(m.FindMappedFields(nil)))

	assert.Equal(t, []destField{dest.Info}, m.FindMappedFields([]sourceField{source.Sku}))
	assert.Equal(t, []destField{dest.Info}, m.FindMappedFields([]sourceField{source.Sku, source.Name}))

	assert.Equal(t, []destField{dest.Info, dest.Detail}, m.FindMappedFields([]sourceField{source.Sku, source.Body}))
}

type sourceSellerInfo struct {
	Root sourceField

	Logo sourceField
}

type sourceSeller struct {
	Root sourceField

	ID   sourceField
	Name sourceField
	Info sourceSellerInfo
}

func (d sourceSeller) GetRoot() sourceField {
	return d.Root
}

type sourceDataComplex struct {
	Root     sourceField
	Sku      sourceField
	Name     sourceField
	Body     sourceField
	Seller   sourceSeller
	ImageURL sourceField
}

func (d sourceDataComplex) GetRoot() sourceField { return d.Root }

func (d sourceDataComplex) GetSeller() sourceSeller { return d.Seller }

type destInfo struct {
	Root destField

	Sku  destField
	Name destField
}

type destDetail struct {
	Root destField
	Body destField
}

func (d destDetail) GetRoot() destField {
	return d.Root
}

type destDataComplex struct {
	Root       destField
	Info       destInfo
	Detail     destDetail
	SearchText destField
}

func (d destDataComplex) GetRoot() destField {
	return d.Root
}

func (d destDataComplex) GetDetail() destDetail {
	return d.Detail
}

func TestMapping_Complex(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.Info.Root),
				NewMapping(source.Name, dest.Info.Root),
				NewMapping(source.Seller.Root, dest.Detail.Root),
				NewMapping(source.Body, dest.Detail.Body),
				NewMapping(source.ImageURL, dest.Detail.Root),
			),
		)

		assert.Equal(t, 0, len(m.FindMappedFields(nil)))

		assert.Equal(t, []destField{dest.Info.Root}, m.FindMappedFields([]sourceField{source.Sku}))
		assert.Equal(t, []destField{dest.Info.Root}, m.FindMappedFields([]sourceField{source.Sku, source.Name}))

		assert.Equal(t, []destField{dest.Detail.Body}, m.FindMappedFields([]sourceField{source.Body}))

		assert.Equal(t, []destField{dest.Detail.Root}, m.FindMappedFields([]sourceField{source.Seller.Root}))

		// From Parent
		assert.Equal(t, []destField{dest.Detail.Root}, m.FindMappedFields([]sourceField{source.Seller.ID}))
		assert.Equal(t, []destField{dest.Detail.Root}, m.FindMappedFields([]sourceField{source.Seller.Name}))

		assert.Equal(t, []destField{dest.Detail.Root}, m.FindMappedFields([]sourceField{source.Seller.Info.Root}))

		assert.Equal(t, []destField{dest.Detail.Root}, m.FindMappedFields([]sourceField{source.Seller.Info.Logo}))

		assert.Equal(t, []destField{dest.Detail.Root},
			m.FindMappedFields([]sourceField{source.Seller.ID, source.Seller.Info.Logo}))

		assert.Equal(t, []destField{dest.Info.Root, dest.Detail.Root},
			m.FindMappedFields([]sourceField{source.Sku, source.Seller.Info.Logo}))
	})

	t.Run("one field to multiple dest fields using logical AND", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.Info.Root, dest.SearchText),
			),
		)

		assert.Equal(t, []destField{dest.Info.Root, dest.SearchText}, m.FindMappedFields([]sourceField{source.Sku}))
	})

	t.Run("one field to multiple dest fields using logical OR, found the first source field", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Name, dest.Info.Root),
				NewMapping(source.Sku, dest.Info.Root),
				NewMapping(source.Sku, dest.SearchText),
			),
		)

		assert.Equal(t, []destField{dest.Info.Root}, m.FindMappedFields([]sourceField{source.Sku}))
	})

	t.Run("one field to multiple dest fields using logical OR, found the first source field", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.SearchText),
				NewMapping(source.Sku, dest.Info.Root),
			),
		)

		assert.Equal(t, []destField{dest.SearchText}, m.FindMappedFields([]sourceField{source.Sku}))
	})

	t.Run("children before parent", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Seller.Root, dest.SearchText),
				NewMapping(source.Seller.Name, dest.Detail.Body),
			),
		)

		assert.Equal(t, []destField{dest.Detail.Body}, m.FindMappedFields([]sourceField{source.Seller.Name}))
	})

	t.Run("not found any mapping", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.Info.Sku),
			),
		)

		assert.Equal(t, 0, len(m.FindMappedFields([]sourceField{source.Seller.Name})))
	})

	t.Run("duplicated", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		assert.PanicsWithValue(t, `duplicated destination field "Info.Sku" for source field "Sku"`, func() {
			NewMapper(
				sourceFm, destFm,
				WithSimpleMapping(sourceFm, destFm,
					NewMapping(source.Sku, dest.Info.Sku),
					NewMapping(source.Sku, dest.Info.Sku),
				),
			)
		})
	})

	t.Run("with AND not duplicated", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		_ = NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.Info.Sku, dest.SearchText),
				NewMapping(source.Sku, dest.Info.Sku),
			),
		)
	})

	t.Run("multiple source fields to one dest field", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.SearchText),
				NewMapping(source.Name, dest.SearchText),
				NewMapping(source.Seller.Name, dest.SearchText),
			),
		)

		assert.Equal(t, []destField{dest.SearchText}, m.FindMappedFields([]sourceField{source.Sku}))
	})

	t.Run("panics when mapping without dest fields", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()

		assert.PanicsWithValue(t, "missing destination fields", func() {
			NewMapper(
				sourceFm, destFm,
				WithSimpleMapping(sourceFm, destFm,
					NewMapping[sourceField, destField](source.Sku),
				),
			)
		})
	})

	t.Run("with inherit mapping", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		subSourceFm := New[sourceField, sourceSeller]()
		subDestFm := New[destField, destDetail]()

		subSource := subSourceFm.GetMapping()
		subDest := subDestFm.GetMapping()

		subMapper := NewMapper(
			subSourceFm, subDestFm,
			WithSimpleMapping(subSourceFm, subDestFm,
				NewMapping(subSource.ID, subDest.Body),
				NewMapping(subSource.Name, subDest.Body),
				NewMapping(subSource.Info.Logo, subDest.Body),
			),
		)

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.Info.Sku),
			),
			WithInheritMapping(
				sourceFm, destFm,
				subMapper,
				sourceDataComplex.GetSeller,
				destDataComplex.GetDetail,
			),
		)

		assert.Equal(t, []destField{dest.Info.Sku}, m.FindMappedFields([]sourceField{source.Sku}))
		assert.Equal(t, []destField{dest.Detail.Body}, m.FindMappedFields([]sourceField{source.Seller.ID}))
		assert.Equal(t, []destField{dest.Detail.Body}, m.FindMappedFields([]sourceField{source.Seller.Name}))

		assert.Equal(t, 0, len(m.FindMappedFields([]sourceField{source.Seller.Root})))

		assert.Equal(t, []destField{dest.Detail.Body}, m.FindMappedFields([]sourceField{source.Seller.Info.Logo}))
		assert.Equal(t, 0, len(m.FindMappedFields([]sourceField{source.Seller.Info.Root})))
	})

	t.Run("with inherit mapping to destination itself", func(t *testing.T) {
		sourceFm := New[sourceField, sourceDataComplex]()
		destFm := New[destField, destDataComplex]()

		source := sourceFm.GetMapping()
		dest := destFm.GetMapping()

		subSourceFm := New[sourceField, sourceSeller]()

		subSource := subSourceFm.GetMapping()

		subMapper := NewMapper(
			subSourceFm, destFm,
			WithSimpleMapping(subSourceFm, destFm,
				NewMapping(subSource.ID, dest.SearchText),
				NewMapping(subSource.Name, dest.Detail.Body),
				NewMapping(subSource.Info.Logo, dest.Info.Name),
			),
		)

		m := NewMapper(
			sourceFm, destFm,
			WithSimpleMapping(sourceFm, destFm,
				NewMapping(source.Sku, dest.Info.Sku),
			),
			WithInheritMapping(
				sourceFm, destFm,
				subMapper,
				sourceDataComplex.GetSeller,
				func(m destDataComplex) destDataComplex { return m },
			),
		)

		assert.Equal(t, []destField{dest.Info.Sku}, m.FindMappedFields([]sourceField{source.Sku}))
		assert.Equal(t, []destField{dest.SearchText}, m.FindMappedFields([]sourceField{source.Seller.ID}))
		assert.Equal(t, []destField{dest.Detail.Body}, m.FindMappedFields([]sourceField{source.Seller.Name}))

		assert.Equal(t, 0, len(m.FindMappedFields([]sourceField{source.Seller.Root})))

		assert.Equal(t, []destField{dest.Info.Name}, m.FindMappedFields([]sourceField{source.Seller.Info.Logo}))
		assert.Equal(t, 0, len(m.FindMappedFields([]sourceField{source.Seller.Info.Root})))
	})
}
