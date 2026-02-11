package okeycrawler

import (
	"testing"

	"github.com/Raime-34/crawler.git/internal/cfg"
	"github.com/Raime-34/crawler.git/internal/dto"
	"github.com/stretchr/testify/assert"
)

func Test_txt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trim spaces",
			in:   "  milk  ",
			want: "milk",
		},
		{
			name: "replace non breaking spaces",
			in:   "10\u00a0шт",
			want: "10 шт",
		},
		{
			name: "trim and replace",
			in:   " \u00a0  цена\u00a0100 \u00a0",
			want: "цена 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := txt(tt.in)
			if got != tt.want {
				t.Fatalf("txt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_okeyCrawler_handleMainProduectPage(t *testing.T) {
	crawler := NewOkeyCrawler()
	cfg.GetConfig().Category = "bzmzh-moloko-utp-selo-zelenoe-3-2-950ml"
	products, err := crawler.LoadMajorCategory()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(products))

	expected := dto.ProductInfo{
		Name:         "Молоко ультрапастеризованное Село Зеленое 3,2% 950мл БЗМЖ",
		SKU:          "863166",
		Price:        "119,99 ₽",
		PriceRaw:     "119.99",
		Currency:     "RUB",
		Availability: "http://schema.org/InStock",
		Image:        "/wcsstore/OKMarketCAS/cat_entries/863166/863166_fullimage.jpg",
		Attrs: map[string]string{
			"Белки, г:":       "3",
			"Бренд :":         "Село Зеленое",
			"Вес (в кг):":     "0,95",
			"Вид обработки:":  "Стерилизованное",
			"Вид:":            "Молоко коровье",
			"Группа товаров:": "Молоко и сливки",
			"Жирность:":       "3,2%",
			"Жиры, г:":        "3,2",
			"Изготовитель:":   `ОАО "Милком"`,
			"Литраж:":         "0,95",
			"Максимальная температура хранения (С⁰):": "25",
			"Минимальная температура хранения (С⁰):":  "2",
			"Срок годности (в днях):":                 "210",
			"Углеводы, г:": "4,7",
			"Энергетическая ценность (Ккал):": "60",
			"Описание товара":                 `Молоко ультрапастеризованное Село Зелёное 3,2%, 950 л. "Село Зелёное" издавна славится своим вкусным и полезным молоком. Его дают коровы, которые пасутся на бескрайних зеленых лугах и пьют чистую родниковую воду. Бережно сохраняя все самое лучшее, что есть в молоке, "Село Зелёное" дарит его вам. Товар может быть как с винтовой крышкой, так и с крышкой клапаном`,
			"Состав товара":                   "Молоко нормализованное.",
			"Меры предосторожности":           "Вскрытую упаковку хранить в холодильнике при температуре 2С-4С не более 3-х суток в пределах срока годности.",
		},
		Desc:        `Молоко ультрапастеризованное Село Зелёное 3,2%, 950 л. "Село Зелёное" издавна славится своим вкусным и полезным молоком. Его дают коровы, которые пасутся на бескрайних зеленых лугах и пьют чистую родниковую воду. Бережно сохраняя все самое лучшее, что есть в молоке, "Село Зелёное" дарит его вам. Товар может быть как с винтовой крышкой, так и с крышкой клапаном`,
		Composition: "Молоко нормализованное.",
		Precautions: "Вскрытую упаковку хранить в холодильнике при температуре 2С-4С не более 3-х суток в пределах срока годности.",
	}

	assert.Equal(
		t,
		expected,
		products[0],
	)
}
