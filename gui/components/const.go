package components

import (
	"time"

	"gioui.org/layout"
)

const (
	FPS             = 25
	AnimTime        = 0.3
	updatePerSecond = time.Second / FPS
)

var old time.Time

func AnimSpeed(gtx layout.Context) float32 {
	return 1.0 / AnimTime / float32(FPS)
}

func Time(gtx layout.Context) time.Duration {
	d := gtx.Now.Sub(old)
	if d < 0 {
		d = -d
	}
	old = time.Now()
	return updatePerSecond - d
}
