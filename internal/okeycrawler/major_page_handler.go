package okeycrawler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/chromedp/chromedp"
)

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
