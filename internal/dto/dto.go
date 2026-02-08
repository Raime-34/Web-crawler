package dto

type ProductInfo struct {
	Id           string
	Name         string
	Weight       string
	Image        string
	OfferedPrice string
	ListPrice    string
}

func NewProductInfo(id, name, weight, image string) ProductInfo {
	return ProductInfo{
		Id:     id,
		Name:   name,
		Weight: weight,
		Image:  image,
	}
}

func (i *ProductInfo) AddPriceInfo(prices PriceInfo) {
	i.OfferedPrice = prices.OfferPrice
	i.ListPrice = prices.ListPrice
}

type PricesDto struct {
	Prices map[string]PriceInfo `json: "prices"`
}

type PriceInfo struct {
	OfferPrice string `json: "offerPrice"`
	ListPrice  string `json: "listPrice"`
}
