package okeycrawler

const (
	okeyBaseUrl = "https://www.okeydostavka.ru/msk/%v?storeId=%v"

	majorPageType       = "major"
	minorPageType       = "minor"
	mainProductPageType = "product_main"
	unknownPageType     = "unknown"

	body                = "body"
	productLinkSelector = "div.product-name a[title]"

	// major page

	categoryCardSelector = `.rows.categories > div.col-xs-5.col-sm-4.col-md-3.col-lg-3.col-xl-2.col-xl-special`
	categoryCardXPath    = `//div[contains(@class,'rows') and contains(@class,'categories')]/div[contains(@class,'col-xs-5') and contains(@class,'col-sm-4') and contains(@class,'col-md-3') and contains(@class,'col-lg-3') and contains(@class,'col-xl-2') and contains(@class,'col-xl-special')]`

	// minor page

	pagingController             = `(//div[contains(@class,'paging_controls')])[1]`
	productLinkByIndexXPath      = `(//ul[contains(@class,'grid_mode') and contains(@class,'grid')]//li//div[contains(@class,'product-name')]//a[@href and @title])[%d]`
	activePageLinkXPath          = `(//div[contains(@class,'paging_controls')])[1]//a[contains(@class,'active') and contains(@class,'selected')]`
	productListingWidgetSelector = ".productListingWidget"
	productGridSelector          = `ul.grid_mode.grid`
	productMainInfoSelector      = ".product_main_info"

	// main page

	productNameSelector           = "h1.main_header[itemprop='name']"
	productSKUMetaSelector        = `meta[itemprop="sku"]`
	productMainImageSelector      = `#productMainImage`
	productOfferSelector          = `#product-price-section[itemprop="offers"]`
	productPriceMetaSelector      = `meta[itemprop="price"]`
	productCurrencyMetaSelector   = `meta[itemprop="priceCurrency"]`
	productAvailabilitySelector   = `link[itemprop="availability"]`
	productPriceInputSelector     = `input[id^="ProductInfoPrice_"]`
	productAttributesItemSelector = "ul.widget-list.attributes li.attributes__item"
	productAttributeNameSelector  = ".attributes__name"
	productAttributeValueSelector = ".attributes__value"

	activePageXPath  = pagingController + `//a[contains(@class,'active') and contains(@class,'selected')]`
	nextPageSelector = activePageXPath + `/following-sibling::a[contains(@class,'hoverover')][1]`
)
