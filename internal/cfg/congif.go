package cfg

import (
	"flag"
	"sync"
)

var (
	cfg  *Cfg
	once sync.Once
)

type Cfg struct {
	Headless bool
	Category string
}

func GetConfig() *Cfg {
	once.Do(func() {
		newCfg := Cfg{}

		flag.BoolVar(&newCfg.Headless, "h", false, "Флаг для открытия браузера во время сбора данных. На Windows этот флаг лучше оставить дефолтным")
		flag.StringVar(&newCfg.Category, "c", "molochnye-produkty-iaitso/molochnye-produkty/moloko-i-slivki", "Название категории, к примеру для https://www.okeydostavka.ru/msk/molochnye-produkty-iaitso/molochnye-produkty/moloko-i-slivki это molochnye-produkty-iaitso/molochnye-produkty/moloko-i-slivki")
		flag.Parse()

		cfg = &newCfg
	})

	return cfg
}
