package dto

type ProductInfo struct {
	Id     string
	Name   string
	Weight string
	Image  string
}

func NewProductInfo(id, name, weight, image string) ProductInfo {
	return ProductInfo{
		Id:     id,
		Name:   name,
		Weight: weight,
		Image:  image,
	}
}
