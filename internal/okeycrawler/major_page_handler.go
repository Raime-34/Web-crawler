package okeycrawler

import (
	"context"
	"fmt"

	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/Raime-34/crawler.git/internal/humanactionemultaion"
	"github.com/chromedp/chromedp"
)

func (c *okeyCrawler) handleMajorCategoryPage(ctx context.Context) []dto.ProductInfo {
	var n int
	_ = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelectorAll("`+categoryCardSelector+`").length`, &n),
	)

	var products []dto.ProductInfo
	for i := 1; i <= n; i++ {
		linkXPath := fmt.Sprintf("(%s)[%d]//div[contains(@class,'product-image')]//a[@href]", categoryCardXPath, i)
		nameXPath := fmt.Sprintf("(%s)[%d]//h2/a", categoryCardXPath, i)
		var catName string

		var htmlContent string
		err := chromedp.Run(ctx,
			// скролим к ссылкам на категории
			chromedp.ScrollIntoView(linkXPath, chromedp.BySearch),
			// делаем вид, что пользователь думает
			humanactionemultaion.Thinking(),
			// получаем название категории, в которую переходим (используется тольк для логирования)
			chromedp.Text(nameXPath, &catName, chromedp.BySearch, chromedp.NodeVisible),
			// логируем категорию
			chromedp.ActionFunc(func(ctx context.Context) error {
				fmt.Printf("Переход на %v\n", catName)
				return nil
			}),
			// кликаем на категорию
			chromedp.Click(linkXPath, chromedp.BySearch),

			// Ожидаем прогрузки страницы
			chromedp.WaitVisible(body, chromedp.ByQuery),

			// подгружаем страницу категории
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
			// обрабатываем ее
			chromedp.ActionFunc(func(ctx context.Context) error {
				newProducts := c.handlePage(ctx, htmlContent, majorPageType)
				products = append(products, newProducts...)
				return nil
			}),

			// после обработки возвращаемся назад
			chromedp.NavigateBack(),
			chromedp.WaitVisible(categoryCardSelector, chromedp.ByQuery),
		)
		if err != nil {
			continue
		}
	}

	return products
}
