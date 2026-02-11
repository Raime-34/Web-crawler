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
	StoreId  string

	ProxyUrl      string
	ProxyUser     string
	ProxyPassword string
}

func GetConfig() *Cfg {
	once.Do(func() {
		newCfg := Cfg{}

		flag.BoolVar(&newCfg.Headless, "h", false, "Флаг для открытия браузера во время сбора данных. На Windows этот флаг лучше оставить дефолтным")
		flag.StringVar(&newCfg.Category, "c", "molochnye-produkty-iaitso/molochnye-produkty/moloko-i-slivki", "Название категории, к примеру для https://www.okeydostavka.ru/msk/molochnye-produkty-iaitso/molochnye-produkty/moloko-i-slivki это molochnye-produkty-iaitso/molochnye-produkty/moloko-i-slivki")
		flag.StringVar(&newCfg.StoreId, "s", "10151", "ID магазина для которого нужно собрать данные. По умолчанию это ГМ 'Альтуфьево'")
		flag.StringVar(&newCfg.ProxyUrl, "u", "", "Url прокси сервера")
		flag.StringVar(&cfg.ProxyUser, "l", "", "Логин для авторизации прокси")
		flag.StringVar(&cfg.ProxyPassword, "p", "", "Пароль для авторизации прокси")
		flag.Parse()

		cfg = &newCfg
	})

	return cfg
}
