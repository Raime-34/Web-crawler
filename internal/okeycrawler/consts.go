package okeycrawler

const (
	okeyBaseUrl = "https://www.okeydostavka.ru/msk/%v?storeId=%v"

	majorPageType       = "major"
	minorPageType       = "minor"
	mainProductPageType = "product_main"
	unknownPageType     = "unknown"

	body                 = "body"
	productLinkSelector  = "div.product-name a[title]"
	categoryCardSelector = `.rows.categories > div.col-xs-5.col-sm-4.col-md-3.col-lg-3.col-xl-2.col-xl-special`
	categoryCardXPath    = `//div[contains(@class,'rows') and contains(@class,'categories')]/div[contains(@class,'col-xs-5') and contains(@class,'col-sm-4') and contains(@class,'col-md-3') and contains(@class,'col-lg-3') and contains(@class,'col-xl-2') and contains(@class,'col-xl-special')]`
)
