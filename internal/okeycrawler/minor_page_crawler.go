package okeycrawler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Raime-34/crawler.git/internal/humanactionemultaion"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func (c *okeyCrawler) handleMinorCategoryPage(ctx context.Context) {
	currentPage := 1
	isOnePageCatalog := false
	for {
		fmt.Println("Ожидаем контроллера страниц")
		dCtx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		chromedp.Run(
			dCtx,
			chromedp.WaitVisible(".paging_controls a.active.selected", chromedp.ByQuery),
		)
		cancel()
		fmt.Println("Контроллер найден")

		isOnePageCatalog = checkAmountOfPage(ctx)
		if isOnePageCatalog {
			fmt.Println("Страница классифицирона как одностраничная")
		} else {
			fmt.Println("Обнаружено множество страниц")
		}

		if !isOnePageCatalog {
			page, err := getCurrentPageNumber(ctx)
			if err != nil {
				fmt.Printf("Ошибка извлечения номера текущей страницы: %v\n", err)
			} else {
				if page != currentPage {
					continue
				}
			}
		}

		fmt.Printf("		текущая страница: %v\n", currentPage)
		var n int
		_ = chromedp.Run(ctx,
			chromedp.WaitVisible(productListingWidgetSelector, chromedp.ByQuery),
			chromedp.Evaluate(`document.querySelectorAll("`+productGridSelector+` li .product-name a[title][href]").length`, &n),
		)

		fmt.Println("разбираем товары")
		// Проходимся по карточкам товаров
		for i := 1; i <= n; i++ {
			sel := fmt.Sprintf(productLinkByIndexXPath, i)

			if err := c.processProduct(ctx, sel); err != nil {
				continue
			}
		}

		if isOnePageCatalog {
			break
		}

		if !checkIfHasNextPage(ctx) {
			break
		}

		var before string
		_ = chromedp.Run(ctx,
			chromedp.Text(activePageXPath, &before, chromedp.BySearch),
		)

		currentPage++
		err := chromedp.Run(ctx,
			chromedp.ScrollIntoView(nextPageSelector, chromedp.BySearch),
			humanactionemultaion.Thinking(),
			chromedp.Click(nextPageSelector, chromedp.BySearch),
			chromedp.WaitVisible(productGridSelector, chromedp.ByQuery),
		)
		if err != nil {
			break
		}
	}

	fmt.Println("Minor page processed")
	for i := 0; i < currentPage-1; i++ {
		chromedp.Run(
			ctx,
			chromedp.NavigateBack(),
			chromedp.Sleep(1*time.Second),
			chromedp.WaitReady(body),
		)
	}
}

// Функция-хэлпер для определения количества страниц в категории
func checkAmountOfPage(ctx context.Context) bool {
	var roots []*cdp.Node
	_ = chromedp.Run(
		ctx,
		chromedp.WaitReady(body),
		chromedp.Nodes(pagingController, &roots, chromedp.BySearch, chromedp.AtLeast(0)),
	)
	return len(roots) == 0
}

// хэлпер для получения текущей страницы
func getCurrentPageNumber(ctx context.Context) (int, error) {
	var pageStr string
	chromedp.Run(ctx,
		chromedp.Text(activePageLinkXPath, &pageStr, chromedp.BySearch),
	)

	return strconv.Atoi(strings.TrimSpace(pageStr))
}

func (c *okeyCrawler) processProduct(ctx context.Context, sel string) error {
	var name string
	var html string
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(productGridSelector, chromedp.ByQuery),
		chromedp.ScrollIntoView(sel, chromedp.BySearch),
		humanactionemultaion.Thinking(),

		chromedp.WaitReady(body),
		chromedp.AttributeValue(sel, "title", &name, nil, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Printf("	∟ переход на страницу товара %v\n", name)
			return nil
		}),
		chromedp.Click(sel, chromedp.BySearch),
		chromedp.WaitVisible(productLinkSelector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			c.handlePage(ctx, html, minorPageType)
			return nil
		}),

		chromedp.NavigateBack(),
		chromedp.WaitVisible(productGridSelector, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
	)

	return err
}

// Хэлпер для проверки наличия следующей страницы
func checkIfHasNextPage(ctx context.Context) bool {
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
		`, nextPageSelector)
	chromedp.Run(ctx,
		chromedp.EvaluateAsDevTools(js, &hasNext),
	)

	return hasNext
}
