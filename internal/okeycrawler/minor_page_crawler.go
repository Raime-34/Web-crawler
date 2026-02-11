package okeycrawler

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

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
