package dto

type ProductInfo struct {
	Name         string
	SKU          string
	Price        string
	PriceRaw     string
	Currency     string
	Availability string
	Image        string
	Attrs        map[string]string
	Desc         string
	Composition  string
	Precautions  string
}
