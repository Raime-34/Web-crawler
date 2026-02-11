package okeycrawler

import (
	"context"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/chromedp/chromedp"
)

func (c *okeyCrawler) handleMainProduectPage(ctx context.Context) dto.ProductInfo {
	var html string
	_ = chromedp.Run(ctx,
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(".product_main_info", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	p := dto.ProductInfo{Attrs: map[string]string{}}

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

func txt(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\u00a0", " ")) }
