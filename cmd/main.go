package main

import (
	"context"

	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/Raime-34/crawler.git/internal/okeycrawler"
)

func main() {
	var crawler Crawler = okeycrawler.NewOkeyCrawler()
	crawler.LoadMajorCategory()
	// goods, err := crawler.LoadPages()
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("Сохраняем в файл...")
	// data, _ := json.Marshal(goods)

	// fmt.Println(len(goods))

	// category := cfg.GetConfig().Category
	// ind := strings.LastIndex(category, `/`)
	// if ind > 0 {
	// 	category = category[ind:]
	// 	err = os.WriteFile(fmt.Sprintf("reports%v.json", category), data, 0644)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }
}

type Crawler interface {
	LoadMajorCategory() (map[string][]*dto.ProductInfo, error)
	LoadPages(context.Context, string) ([]*dto.ProductInfo, error)
}
