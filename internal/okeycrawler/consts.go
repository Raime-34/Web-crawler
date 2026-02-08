package okeycrawler

const (
	okeyBaseUrl           = "https://www.okeydostavka.ru/msk/%v"
	imageBaseUrl          = "https://www.okeydostavka.ru%v"
	totalPagesRegexp      = `WCParamJS\.totalPages\s*=\s*'([\d.]+)'`
	productContainerClass = ".grid_mode.grid"
	okeyAmountOfGoods     = 72
	okeyBaseFilter        = "#facet:&productBeginIndex:%v&orderBy:2&pageView:grid&pageSize:%v&"
)
