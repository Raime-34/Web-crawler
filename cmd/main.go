package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/PuerkitoBio/goquery"
	"github.com/Raime-34/crawler.git/internal/okeycrawler"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	crawler := okeycrawler.NewOkeyCrawler()
	err := crawler.LoadPages()
	if err != nil {
		fmt.Println(err)
	}
}

func undetectedChromedp() {
	ctx, cancel, err := cu.New(cu.NewConfig(
		// cu.WithHeadless(),
		cu.WithTimeout(10 * time.Second),
	))
	if err != nil {
		panic(err)
	}
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev any) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			// fmt.Println(ev.Type)
			// fmt.Println(ev.Response.URL)

			if ev.Type != "Document" {
				return
			}

			// go func() {
			// 	c := chromedp.FromContext(ctx)
			// 	rbp := network.GetResponseBody(ev.RequestID)
			// 	body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
			// 	if err != nil {
			// 		fmt.Println(ev.Type)
			// 		fmt.Println("Ошибка: " + err.Error())
			// 	}
			// 	if err == nil {
			// 		if bytes.Contains(body, []byte("FacetCount_updated")) {
			// 			re := regexp.MustCompile(`totalCount\s*:\s*(\d+)`)
			// 			m := re.FindSubmatch(body)
			// 			if len(m) == 2 {
			// 				fmt.Println(m)
			// 				total, _ := strconv.Atoi(string(m[1]))
			// 				log.Println("totalCount =", total)
			// 			}
			// 		}
			// 	}
			// }()
		}
	})

	var htmlContent string
	if err := chromedp.Run(
		ctx,

		network.Enable(),

		chromedp.Navigate("https://www.okeydostavka.ru/msk/molochnye-produkty-iaitso/molochnye-produkty/moloko-i-slivki"),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	); err != nil {
		panic(err)
	}

	log.Println(strings.Contains(htmlContent, "totalPages"))

	reader := strings.NewReader(htmlContent)

	re := regexp.MustCompile(`WCParamJS\.totalPages\s*=\s*'([\d.]+)'`)
	m := re.FindStringSubmatch(htmlContent)
	if len(m) > 1 {
		fmt.Println("totalPages:", m[1])
		f, _ := strconv.ParseFloat(m[1], 64)
		fmt.Println(int(f))
	}

	doc, _ := goquery.NewDocumentFromReader(reader)

	sel := doc.Find(".product_listing_container").First()

	grid := sel.Children().First()

	fmt.Println(grid.Children().Length())
}

type Crawler interface {
	LoadPages() error
}
