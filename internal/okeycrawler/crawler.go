package okeycrawler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Raime-34/crawler.git/internal/cfg"
	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/Raime-34/crawler.git/internal/humanactionemultaion"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
)

type okeyCrawler struct {
	mu       sync.Mutex
	products []dto.ProductInfo
}

func NewOkeyCrawler() *okeyCrawler {
	return &okeyCrawler{
		products: make([]dto.ProductInfo, 0),
	}
}

func (c *okeyCrawler) LoadMajorCategory() ([]dto.ProductInfo, error) {
	config := cfg.GetConfig()

	// Для сбора инфорамции используется chromedp
	// Это либа использует хром для перехода на страницы
	// Причем, либа сама хэндлит кукесы и, благодаря испольованию браузера,
	// позволяет обходить защиту от ботов

	// Либа позволяет работать в "безголовом" режиме
	// На винде это не работает безголовый режим,
	// то есть на время работы тулзы браузер будет открываться

	proxyURL := config.ProxyUrl
	proxyUser := config.ProxyUser
	proxyPass := config.ProxyPassword

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(proxyURL),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("headless", config.Headless),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// конектимся к прокси
	authorize(ctx, proxyURL, proxyUser, proxyPass)

	go humanactionemultaion.EmulateMouseAction(ctx)

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

func authorize(ctx context.Context, proxyURL, proxyUser, proxyPass string) {
	if proxyURL == "" || proxyUser == "" || proxyPass == "" {
		return
	}

	lctx, lcancel := context.WithCancel(ctx)
	chromedp.ListenTarget(lctx, func(ev any) {
		switch e := ev.(type) {
		case *fetch.EventRequestPaused:
			go func() { _ = chromedp.Run(ctx, fetch.ContinueRequest(e.RequestID)) }()
		case *fetch.EventAuthRequired:
			if e.AuthChallenge.Source != fetch.AuthChallengeSourceProxy {
				return
			}
			go func() {
				_ = chromedp.Run(ctx,
					fetch.ContinueWithAuth(e.RequestID, &fetch.AuthChallengeResponse{
						Response: fetch.AuthChallengeResponseResponseProvideCredentials,
						Username: proxyUser,
						Password: proxyPass,
					}),
					fetch.Disable(),

					chromedp.Sleep(10*time.Second),
				)
				lcancel()
			}()
		}
	})
}
