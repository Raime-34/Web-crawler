package okeycrawler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/PuerkitoBio/goquery"
	"github.com/Raime-34/crawler.git/internal/cfg"
	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/go-vgo/robotgo"
)

type okeyCrawler struct {
	mu       sync.Mutex
	prices   map[string]dto.PriceInfo
	products []dto.ProductInfo2
}

func NewOkeyCrawler() *okeyCrawler {
	return &okeyCrawler{
		prices:   make(map[string]dto.PriceInfo),
		products: make([]dto.ProductInfo2, 0),
	}
}

func (c *okeyCrawler) LoadPages(ctx context.Context, url string) ([]*dto.ProductInfo, error) {
	// Загружаем первую страницу категории
	initialPageUrl := url
	fmt.Printf("Загружаеим начальную страницу: %s\n", initialPageUrl)

	html, err := c.loadPage(ctx, initialPageUrl)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при загрузке начальной страницы категории: %w", err)
	}

	// Получаем общее число страниц
	amountOfPagesPtr, err := c.getAmountOfPages(*html)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить общее число страниц: %w", err)
	}

	// Получаем товары с начальной страницы
	goods, err := c.getInfoFromPage(*html)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить информацию о товарах: %w", err)
	}

	if amountOfPages := *amountOfPagesPtr; amountOfPages > 1 {
		for i := 2; i <= amountOfPages; i++ {
			n := (i - 1) * okeyAmountOfGoods
			nextPageUrl := initialPageUrl + fmt.Sprintf(okeyBaseFilter, n, okeyAmountOfGoods)
			fmt.Printf("Страница %v (%v): загружается...\n", i, nextPageUrl)

			nextHtml, err := c.loadPage(ctx, nextPageUrl)
			if err != nil {
				fmt.Println(err)
			}

			nextGoods, err := c.getInfoFromPage(*nextHtml)
			if err != nil {
				fmt.Printf("Страница %v: ошибка\n", i)
				return nil, fmt.Errorf("Ошибка загрузки страницы")
			}

			goods = append(goods, nextGoods...)
		}
	}

	fmt.Println("Подгружаем цены...")
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, g := range goods {
		if prices, ok := c.prices[g.Id]; ok {
			g.AddPriceInfo(prices)
		} else {
			fmt.Printf("Цена для %v (%v) не найдена\n", g.Name, g.Id)
		}
	}

	return goods, nil
}

func (c *okeyCrawler) addListeners(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev any) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			if ev.Type != "XHR" {
				return
			}

			go func() {
				ctx2 := chromedp.FromContext(ctx)
				rbp := network.GetResponseBody(ev.RequestID)
				body, _ := rbp.Do(cdp.WithExecutor(ctx, ctx2.Target))
				if bytes.Contains(body, []byte("prices")) {
					var data dto.PricesDto
					json.Unmarshal(body, &data)
					c.mu.Lock()
					for k, v := range data.Prices {
						c.prices[k] = v
					}
					c.mu.Unlock()
				}
			}()
		}
	})
}

func (c *okeyCrawler) loadPage(ctx context.Context, endpoint string) (*string, error) {
	var htmlContent string

	if err := chromedp.Run(
		ctx,

		network.Enable(),

		chromedp.Navigate(endpoint),
		chromedp.Reload(),
		chromedp.WaitReady("body"),
		chromedp.WaitVisible("div.product-name a[title]", chromedp.ByQuery),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		return nil, fmt.Errorf("Ошибка загрузки страницы: %w", err)
	}

	if htmlContent == "" {
		return nil, fmt.Errorf("Пустая страница")
	}

	return &htmlContent, nil
}

func (c *okeyCrawler) getAmountOfPages(html string) (*int, error) {
	re := regexp.MustCompile(totalPagesRegexp)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 1 {
		return nil, fmt.Errorf("Поле totalPages не найдено")
	}

	f, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("Ошибка парсинга: %w", err)
	}

	totalPages := int(f)
	return &totalPages, nil
}

func (c *okeyCrawler) getInfoFromPage(html string) ([]*dto.ProductInfo, error) {
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	sel := doc.Find(productContainerClass).First()

	var productInfo []*dto.ProductInfo

	sel.ChildrenFiltered("li").Each(func(i int, li *goquery.Selection) {
		// ищем id товара
		id, _ := li.Find(`.product[data-catentry-id]`).Attr("data-catentry-id")

		// его название
		name, _ := li.Find(`div.product-name a`).Attr("title")

		// вес
		weight := strings.Join(strings.Fields(li.Find(".product-weight").Text()), " ")

		// ссылку на превьюшку
		img, _ := li.Find(`img[data-src]`).Attr("data-src")
		img = fmt.Sprintf(imageBaseUrl, img)

		newProductInfo := dto.NewProductInfo(id, name, weight, img)
		productInfo = append(productInfo, &newProductInfo)
	})

	return productInfo, nil
}

func (c *okeyCrawler) checkPageType(html string) string {
	switch {
	case strings.Contains(html, "rows categories"):
		return majorPageType
	case strings.Contains(html, "grid_mode grid rows"):
		return minorPageType
	case strings.Contains(html, "rows product_main_info"):
		return mainProductPageType
	default:
		return unknownPageType
	}
}

func (c *okeyCrawler) LoadMajorCategory() ([]dto.ProductInfo2, error) {
	config := cfg.GetConfig()

	// Для сбора инфорамции используется chromedp
	// Это либа использует хром для перехода на страницы
	// Причем, либа сама хэндлит кукесы и, благодаря испольованию браузера,
	// позволяет обходить защиту от ботов (последнее справедливо для форки chromedp-undetected)

	crawlerOptions := []cu.Option{
		// cu.WithTimeout(60 * time.Minute),
	}

	// Либа позволяет работать в "безголовом" режиме
	// На винде это не работает безголовый режим,
	// то есть на время работы тулзы браузер будет открываться
	if config.Headless {
		crawlerOptions = append(crawlerOptions, cu.WithHeadless())
	}

	ctx, cancel, err := cu.New(cu.NewConfig(crawlerOptions...))
	if err != nil {
		return nil, fmt.Errorf("Ошибка инициализации кроулера: %w", err)
	}
	defer cancel()

	c.addListeners(ctx)
	go c.emulateMouse(ctx)

	// Загружаем первую страницу категории
	initialPageUrl := fmt.Sprintf(okeyBaseUrl, config.Category, config.StoreId)
	fmt.Printf("Загружаеим начальную страницу: %s\n", initialPageUrl)

	html, err := c.loadPage(ctx, initialPageUrl)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при загрузке начальной страницы категории: %w", err)
	}

	products := c.handlePage(ctx, *html, majorPageType)

	return products, nil
}

func (c *okeyCrawler) emulateMouse(ctx context.Context) {
	err := robotgo.ActiveName("chrome")
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			robotgo.MoveSmooth(300, 300)
			time.Sleep(5 * time.Second)
			robotgo.MoveSmooth(600, 600)
			time.Sleep(5 * time.Second)
		}
	}
}

func (c *okeyCrawler) handlePage(ctx context.Context, html string, prevPageType string) []dto.ProductInfo2 {
	pageType := c.checkPageType(html)

	switch pageType {
	case majorPageType:
		newProducts := c.handleMajorCategoryPage(ctx)
		c.products = append(c.products, newProducts...)
	case minorPageType:
		if prevPageType != minorPageType {
			c.handleMinorCategoryPage(ctx)
		} else {
			time.Sleep(10 * time.Second)
			c.products = append(c.products, c.handleMainProduectPage(ctx))
		}
	case mainProductPageType:
		c.products = append(c.products, c.handleMainProduectPage(ctx))
	}

	return c.products
}

func (c *okeyCrawler) handleMajorCategoryPage(ctx context.Context) []dto.ProductInfo2 {
	cardCSS := `.rows.categories > div.col-xs-5.col-sm-4.col-md-3.col-lg-3.col-xl-2.col-xl-special`
	cardXPath := `//div[contains(@class,'rows') and contains(@class,'categories')]/div[contains(@class,'col-xs-5') and contains(@class,'col-sm-4') and contains(@class,'col-md-3') and contains(@class,'col-lg-3') and contains(@class,'col-xl-2') and contains(@class,'col-xl-special')]`

	var n int
	_ = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelectorAll("`+cardCSS+`").length`, &n),
	)

	var products []dto.ProductInfo2
	for i := 1; i <= n; i++ {
		linkXPath := fmt.Sprintf("(%s)[%d]//div[contains(@class,'product-image')]//a[@href]", cardXPath, i)
		nameXPath := fmt.Sprintf("(%s)[%d]//h2/a", cardXPath, i)
		var catName string

		var htmlContent string
		err := chromedp.Run(ctx,
			chromedp.ScrollIntoView(linkXPath, chromedp.BySearch),
			chromedp.Sleep(time.Duration(1200+rand.Intn(2400))*time.Millisecond),
			chromedp.Text(nameXPath, &catName, chromedp.BySearch, chromedp.NodeVisible),
			chromedp.ActionFunc(func(ctx context.Context) error {
				fmt.Printf("Переход на %v\n", catName)
				return nil
			}),
			chromedp.Click(linkXPath, chromedp.BySearch),

			// якорь целевой страницы (лучше не тот же cardCSS)
			chromedp.WaitVisible("body", chromedp.ByQuery),

			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
			chromedp.ActionFunc(func(ctx context.Context) error {
				newProducts := c.handlePage(ctx, htmlContent, majorPageType)
				products = append(products, newProducts...)
				return nil
			}),

			chromedp.NavigateBack(),
			chromedp.WaitVisible(cardCSS, chromedp.ByQuery),
		)
		if err != nil {
			// лог и continue
			continue
		}
	}

	return products
}

func (c *okeyCrawler) handleMinorCategoryPage(ctx context.Context) {
	listCSS := `ul.grid_mode.grid`
	itemLinkXPath := `(//ul[contains(@class,'grid_mode') and contains(@class,'grid')]//li//div[contains(@class,'product-name')]//a[@href and @title])[%d]`

	// nameXPath := itemLinkXPath + `//div[contains(@class,'product-name')]//a[@href and @title]`

	currentPage := 1
	isOnePageCatalog := false
	for {
		pagingRoot := `(//div[contains(@class,'paging_controls')])[1]`
		activeXPath := pagingRoot + `//a[contains(@class,'active') and contains(@class,'selected')]`

		fmt.Println("Ожидаем контроллера страниц")
		dCtx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		chromedp.Run(
			dCtx,
			chromedp.WaitVisible(".paging_controls a.active.selected", chromedp.ByQuery),
		)
		cancel()
		fmt.Println("Контроллер найден")

		var roots []*cdp.Node
		_ = chromedp.Run(
			ctx,
			chromedp.WaitReady("body"),
			chromedp.Nodes(pagingRoot, &roots, chromedp.BySearch, chromedp.AtLeast(0)),
		)
		if len(roots) == 0 {
			isOnePageCatalog = true
		}
		if isOnePageCatalog {
			fmt.Println("Страница классифицирона как одностраничная")
		} else {
			fmt.Println("Обнаружено множество страниц")
		}

		if !isOnePageCatalog {
			activeXPath2 := `(//div[contains(@class,'paging_controls')])[1]//a[contains(@class,'active') and contains(@class,'selected')]`
			var pageStr string
			_ = chromedp.Run(ctx,
				chromedp.Text(activeXPath2, &pageStr, chromedp.BySearch),
			)

			page, _ := strconv.Atoi(strings.TrimSpace(pageStr))

			if page != currentPage {
				continue
			}
		}

		fmt.Printf("		текущая страница: %v\n", currentPage)
		var n int
		_ = chromedp.Run(ctx,
			chromedp.WaitVisible(".productListingWidget", chromedp.ByQuery),
			chromedp.Evaluate(`document.querySelectorAll("`+listCSS+` li .product-name a[title][href]").length`, &n),
		)

		fmt.Println("разбираем товары")
		// Проходимся по карточкам товаров
		for i := 1; i <= n; i++ {
			sel := fmt.Sprintf(itemLinkXPath, i)

			var name string
			var html string
			err := chromedp.Run(ctx,
				chromedp.WaitVisible(listCSS, chromedp.ByQuery),
				chromedp.ScrollIntoView(sel, chromedp.BySearch),
				chromedp.Sleep(time.Duration(1200+rand.Intn(2600))*time.Millisecond),

				chromedp.WaitReady("body"),
				chromedp.AttributeValue(sel, "title", &name, nil, chromedp.BySearch),
				chromedp.ActionFunc(func(ctx context.Context) error {
					fmt.Printf("	∟ переход на страницу товара %v\n", name)
					return nil
				}),
				chromedp.Click(sel, chromedp.BySearch),
				chromedp.WaitVisible("div.product-name a[title]", chromedp.ByQuery),
				chromedp.OuterHTML("html", &html, chromedp.ByQuery),
				chromedp.ActionFunc(func(ctx context.Context) error {
					c.handlePage(ctx, html, minorPageType)
					return nil
				}),

				chromedp.NavigateBack(),
				chromedp.WaitVisible(listCSS, chromedp.ByQuery),
				chromedp.Sleep(1*time.Second),
			)
			if err != nil {
				continue
			}
		}

		if isOnePageCatalog {
			break
		}

		nextXPath := activeXPath + `/following-sibling::a[contains(@class,'hoverover')][1]`
		var hasNext bool
		js := fmt.Sprintf(`
		(function () {
			return document
			.evaluate(
				%q,
				document,
				null,
				XPathResult.FIRST_ORDERED_NODE_TYPE,
				null
			)
			.singleNodeValue !== null;
		})()
		`, nextXPath)
		_ = chromedp.Run(ctx,
			chromedp.EvaluateAsDevTools(js, &hasNext),
		)

		if !hasNext {
			break
		}

		var before string
		_ = chromedp.Run(ctx,
			chromedp.Text(activeXPath, &before, chromedp.BySearch),
		)

		currentPage++
		err := chromedp.Run(ctx,
			chromedp.ScrollIntoView(nextXPath, chromedp.BySearch),
			chromedp.Sleep(time.Duration(1200+rand.Intn(2600))*time.Millisecond),
			chromedp.Click(nextXPath, chromedp.BySearch),
			chromedp.WaitVisible(listCSS, chromedp.ByQuery),
		)
		if err != nil {
			break
		}
	}

	fmt.Println("Minor page processed")
	for i := 0; i < currentPage-1; i++ {
		fmt.Println("NavigateBack()")
		chromedp.Run(
			ctx,
			chromedp.NavigateBack(),
			chromedp.Sleep(1*time.Second),
			chromedp.WaitReady("body"),
		)
	}
}

func txt(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\u00a0", " ")) }

func (c *okeyCrawler) handleMainProduectPage(ctx context.Context) dto.ProductInfo2 {
	var html string
	_ = chromedp.Run(ctx,
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(".product_main_info", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	p := dto.ProductInfo2{Attrs: map[string]string{}}

	// name
	p.Name = txt(doc.Find("h1.main_header[itemprop='name']").First().Text())

	// sku
	if v, ok := doc.Find(`meta[itemprop="sku"]`).Attr("content"); ok {
		p.SKU = txt(v)
	}

	// image (main)
	if v, ok := doc.Find(`#productMainImage`).Attr("src"); ok {
		p.Image = txt(v) // может быть относительный
	}

	// price + currency + availability (schema.org Offer)
	offer := doc.Find(`#product-price-section[itemprop="offers"]`).First()
	if v, ok := offer.Find(`meta[itemprop="price"]`).Attr("content"); ok {
		p.PriceRaw = txt(v)
	}
	if v, ok := offer.Find(`meta[itemprop="priceCurrency"]`).Attr("content"); ok {
		p.Currency = txt(v)
	}
	if v, ok := offer.Find(`link[itemprop="availability"]`).Attr("href"); ok {
		p.Availability = txt(v) // InStock/OutOfStock
	}
	// цена “как на странице”
	p.Price = txt(doc.Find(`input[id^="ProductInfoPrice_"]`).First().AttrOr("value", ""))

	// атрибуты (и сверху, и в табах) — пары name/value
	doc.Find("ul.widget-list.attributes li.attributes__item").Each(func(_ int, li *goquery.Selection) {
		k := txt(strings.TrimSuffix(li.Find(".attributes__name").First().Text(), ":"))
		v := txt(li.Find(".attributes__value").First().Text())
		if k != "" && v != "" {
			p.Attrs[k] = v
		}
	})

	// поля внутри “Описание товара:” / “Состав товара:” / “Меры предосторожности:”
	doc.Find(`ul.widget-list.attributes li.attributes__item`).Each(func(_ int, li *goquery.Selection) {
		k := txt(strings.TrimSuffix(li.Find(".attributes__name").First().Text(), ":"))
		v := txt(li.Find(".attributes__value").First().Text())
		switch k {
		case "Описание товара":
			p.Desc = v
		case "Состав товара":
			p.Composition = v
		case "Меры предосторожности":
			p.Precautions = v
		}
	})

	// fmt.Printf("Извлеченные данные из страницы %v\n", p)
	return p
}

func (c *okeyCrawler) loadMinorCategoriesRefs(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var hrefs []string
	doc.Find(".rows.categories .product-image a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		hrefs = append(hrefs, fmt.Sprintf(okeyBaseUrl2, href))
	})

	return hrefs, nil
}
