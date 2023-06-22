package fieldmap

type Field int

type ProductFieldMap struct {
	Root Field

	Sku        Field             `json:"sku"`
	Provider   ProviderFieldMap  `json:"provider"`
	Attributes AttributeFieldMap `json:"attributes"`
	SellerIds  Field             `json:"sellerIds"`
	BrandCodes Field             `json:"brandCodes"`
}

type ProviderFieldMap struct {
	Root Field

	Id       Field `json:"id"`
	Name     Field `json:"name"`
	Logo     Field `json:"logo"`
	ImageUrl Field `json:"imageUrl"`
}

type AttributeFieldMap struct {
	Root Field

	Id     Field          `json:"id"`
	Code   Field          `json:"code"`
	Name   Field          `json:"name"`
	Option OptionFieldMap `json:"option"`
}

type OptionFieldMap struct {
	Root Field

	Code Field `json:"code"`
	Name Field `json:"name"`
}
