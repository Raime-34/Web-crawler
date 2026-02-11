package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Raime-34/crawler.git/internal/cfg"
	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/Raime-34/crawler.git/internal/okeycrawler"
)

func main() {
	var crawler Crawler = okeycrawler.NewOkeyCrawler()
	goods, err := crawler.LoadMajorCategory()
	if err != nil {
		fmt.Printf("Ошибка при обработке страницы: %v\n", err)
		return
	}

	fmt.Println("Сохраняем в файл...")
	data, err := json.Marshal(goods)
	if err != nil {
		fmt.Printf("Ошибка записи: %v\n", err)
	}

	fmt.Printf("Всего товаров извлечено: %v\n", len(goods))

	category := cfg.GetConfig().Category
	ind := strings.LastIndex(category, `/`)
	if ind == -1 {
		ind = 0
		category = fmt.Sprintf("/%v", category)
	}

	category = category[ind:]
	err = os.WriteFile(fmt.Sprintf("reports%v.json", category), data, 0644)
	if err != nil {
		fmt.Println(err)
	}

}

type Crawler interface {
	LoadMajorCategory() ([]dto.ProductInfo, error)
}
