package okeycrawler

import (
	"fmt"
	"sync"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/Raime-34/crawler.git/internal/cfg"
	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/Raime-34/crawler.git/internal/humanactionemultaion"
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
	// позволяет обходить защиту от ботов (последнее справедливо для форки chromedp-undetected)
	crawlerOptions := []cu.Option{}

	// Либа позволяет работать в "безголовом" режиме
	// На винде это не работает безголовый режим,
	// то есть на время работы тулзы браузер будет открываться
	if config.Headless {
		crawlerOptions = append(crawlerOptions, cu.WithHeadless())
	}

	ctx, cancel, err := cu.New(cu.NewConfig(crawlerOptions...))
	if err != nil {
		return nil, fmt.Errorf("Ошибка инициализации кроулера: %w", err)
	}
	defer cancel()

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
