package humanactionemultaion

import (
	"math/rand"
	"time"

	"github.com/chromedp/chromedp"
)

func Thinking() chromedp.Action {
	return chromedp.Sleep(time.Duration(1200+rand.Intn(2400)) * time.Millisecond)
}
