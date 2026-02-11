package okeycrawler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

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

func (c *okeyCrawler) handlePage(ctx context.Context, html string, prevPageType string) []dto.ProductInfo {
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
