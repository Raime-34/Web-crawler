package humanactionemultaion

import (
	"context"
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
)

const (
	activeWindow        = "chrome"
	humanMouseMoveDelay = 5 * time.Second
)

var (
	mouseMovePoints = []point{
		{x: 300, y: 300},
		{x: 600, y: 600},
	}
)

type point struct {
	x, y int
}

// Фунция для контроля мыши
// Она лишь перемещает ее в рамках выбранного приложения
func EmulateMouseAction(ctx context.Context) {
	err := robotgo.ActiveName(activeWindow)
	if err != nil {
		fmt.Printf("Ошибка эмуляции мыши: %v\n", err)
	}

	if len(mouseMovePoints) == 0 {
		return
	}

	for i := 0; i < len(mouseMovePoints); i++ {
		select {
		case <-ctx.Done():
			return
		default:
			nextPoint := mouseMovePoints[i]
			robotgo.MoveSmooth(nextPoint.x, nextPoint.y)
			time.Sleep(humanMouseMoveDelay)

			if len(mouseMovePoints)-1 == i {
				i = 0
			}
		}
	}
}
