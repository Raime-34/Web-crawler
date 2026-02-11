package okeycrawler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
