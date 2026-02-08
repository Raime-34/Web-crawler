package okeycrawler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/PuerkitoBio/goquery"
	"github.com/Raime-34/crawler.git/internal/cfg"
	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type okeyCrawler struct{}

func NewOkeyCrawler() *okeyCrawler {
	return &okeyCrawler{}
}

func (c *okeyCrawler) LoadPages() error {
	config := cfg.GetConfig()

	// Для сбора инфорамции используется chromedp
	// Это либа использует хром для перехода на страницы
	// Причем, либа сама хэндлит кукесы и, благодаря испольованию браузера,
	// позволяет обходить защиту от ботов (последнее справедливо для форки chromedp-undetected)

	crawlerOptions := []cu.Option{
		cu.WithTimeout(20 * time.Second),
	}

	// Либа позволяет работать в "безголовом" режиме
	// На винде это не работает безголовый режим,
	// то есть на время работы тулзы браузер будет открываться
	if config.Headless {
		crawlerOptions = append(crawlerOptions, cu.WithHeadless())
	}

	ctx, cancel, err := cu.New(cu.NewConfig(crawlerOptions...))
	if err != nil {
		return fmt.Errorf("Ошибка инициализации кроулера: %w", err)
	}
	defer cancel()

	c.addListeners(ctx)

	// Загружаем первую страницу категории
	initialPageUrl := fmt.Sprintf(okeyBaseUrl, config.Category)
	fmt.Printf("Загружаеим начальную страницу: %s\n", initialPageUrl)

	html, err := c.loadPage(ctx, initialPageUrl)
	if err != nil {
		return fmt.Errorf("Ошибка при загрузке начальной страницы категории: %w", err)
	}

	// Получаем общее число страниц
	amountOfPagesPtr, err := c.getAmountOfPages(*html)
	if err != nil {
		return fmt.Errorf("Не удалось получить общее число страниц: %w", err)
	}

	// Получаем товары с начальной страницы
	goods, err := c.getInfoFromPage(*html)
	if err != nil {
		return fmt.Errorf("Не удалось получить информацию о товарах: %w", err)
	}

	if amountOfPages := *amountOfPagesPtr; amountOfPages > 1 {
		for i := 2; i <= amountOfPages; i++ {
			n := (i - 1) * okeyAmountOfGoods
			nextPageUrl := fmt.Sprintf(okeyBaseUrl, config.Category) + fmt.Sprintf(okeyBaseFilter, n, okeyAmountOfGoods)
			fmt.Printf("Страница %v (%v): загружается...\n", i, nextPageUrl)

			nextHtml, err := c.loadPage(ctx, nextPageUrl)
			if err != nil {
				fmt.Println(err)
			}

			nextGoods, err := c.getInfoFromPage(*nextHtml)
			if err != nil {
				fmt.Printf("Страница %v: ошибка\n", i)
				return fmt.Errorf("Ошибка загрузки страницы")
			}

			goods = append(goods, nextGoods...)
		}
	}

	for _, g := range goods {
		fmt.Println(g)
	}

	return nil
}

func (c *okeyCrawler) addListeners(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev any) {})
}

func (c *okeyCrawler) loadPage(ctx context.Context, endpoint string) (*string, error) {
	var htmlContent string

	if err := chromedp.Run(
		ctx,

		network.Enable(),

		chromedp.Navigate(endpoint),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
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

func (c *okeyCrawler) getInfoFromPage(html string) ([]dto.ProductInfo, error) {
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	sel := doc.Find(productContainerClass).First()

	var productInfo []dto.ProductInfo

	sel.ChildrenFiltered("li").Each(func(i int, li *goquery.Selection) {
		// ищем id товара
		id, _ := li.Find(`.product[data-catentry-id]`).Attr("data-catentry-id")

		// его название
		name, _ := li.Find(`input[id^="ProductInfoName__"]`).Attr("value")
		if name == "" {
			name, _ = li.Find(`.product-name a`).Attr("title")
		}

		// вес
		weight := strings.Join(strings.Fields(li.Find(".product-weight").Text()), " ")

		// ссылку на превьюшку
		img, _ := li.Find(`img[data-src]`).Attr("data-src")
		if img == "" {
			img, _ = li.Find("img").Attr("src")
		}

		productInfo = append(productInfo, dto.NewProductInfo(id, name, weight, img))
	})

	return productInfo, nil
}
