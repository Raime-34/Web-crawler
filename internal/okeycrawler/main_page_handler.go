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
	chromedp.Run(ctx,
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(productMainInfoSelector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	p := dto.ProductInfo{Attrs: map[string]string{}}

	// name
	p.Name = txt(doc.Find(productNameSelector).First().Text())

	// sku
	if v, ok := doc.Find(productSKUMetaSelector).Attr("content"); ok {
		p.SKU = txt(v)
	}

	// image (main)
	if v, ok := doc.Find(productMainImageSelector).Attr("src"); ok {
		p.Image = txt(v) // может быть относительный
	}

	// price + currency + availability (schema.org Offer)
	offer := doc.Find(productOfferSelector).First()
	if v, ok := offer.Find(productPriceMetaSelector).Attr("content"); ok {
		p.PriceRaw = txt(v)
	}
	if v, ok := offer.Find(productCurrencyMetaSelector).Attr("content"); ok {
		p.Currency = txt(v)
	}
	if v, ok := offer.Find(productAvailabilitySelector).Attr("href"); ok {
		p.Availability = txt(v) // InStock/OutOfStock
	}
	// цена “как на странице”
	p.Price = txt(doc.Find(productPriceInputSelector).First().AttrOr("value", ""))

	// атрибуты (и сверху, и в табах) — пары name/value
	doc.Find(productAttributesItemSelector).Each(func(_ int, li *goquery.Selection) {
		k := txt(strings.TrimSuffix(li.Find(productAttributeNameSelector).First().Text(), ":"))
		v := txt(li.Find(productAttributeValueSelector).First().Text())
		if k != "" && v != "" {
			p.Attrs[k] = v
		}
	})

	// поля внутри “Описание товара:” / “Состав товара:” / “Меры предосторожности:”
	doc.Find(productAttributesItemSelector).Each(func(_ int, li *goquery.Selection) {
		k := txt(strings.TrimSuffix(li.Find(productAttributeNameSelector).First().Text(), ":"))
		v := txt(li.Find(productAttributeValueSelector).First().Text())
		switch k {
		case "Описание товара":
			p.Desc = v
		case "Состав товара":
			p.Composition = v
		case "Меры предосторожности":
			p.Precautions = v
		}
	})

	return p
}

func txt(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\u00a0", " ")) }
