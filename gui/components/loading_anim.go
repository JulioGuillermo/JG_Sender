package components

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/x/outlay"
)

type LoadingAnim struct {
	Size   unit.Dp
	Margin unit.Dp
	Len    int
	Anim   outlay.Animation
	Color  color.NRGBA
}

func NewLoadingAnim(size, margin unit.Dp, len int, duration time.Duration, color color.NRGBA) *LoadingAnim {
	anim := &LoadingAnim{
		Size:   size,
		Len:    len,
		Color:  color,
		Margin: margin,
	}
	anim.Anim.Duration = duration
	return anim
}

func (p *LoadingAnim) Reset() {
	p.Anim.Start(time.Now())
}

func (p *LoadingAnim) render(gtx layout.Context, progress, offset, size, margin float64, in, out bool) {
	var x1, x2, y1, y2 float64

	if in {
		anim_offset := size * (1 - progress) / 2
		x1 = offset + (size+margin)*progress
		x2 = margin * progress
		y1 = anim_offset
		y2 = size - anim_offset
	} else if out {
		anim_offset := size * progress / 2
		x1 = offset + (size+margin)*progress - size
		x2 = offset + margin*progress
		y1 = anim_offset
		y2 = size - anim_offset
	} else {
		anim_offset := (size + margin) * progress
		x1 = offset + anim_offset - size
		x2 = offset + anim_offset
		y1 = 0
		y2 = size
	}

	c := clip.Ellipse{
		Min: image.Pt(int(x1), int(y1)),
		Max: image.Pt(int(x2), int(y2)),
	}
	paint.FillShape(gtx.Ops, p.Color, c.Op(gtx.Ops))
}

func flow(x, b float64) float64 {
	return 1.0 / (1.0 + math.Pow(b, -x))
}

func (p *LoadingAnim) Layout(gtx layout.Context) layout.Dimensions {
	if !p.Anim.Animating(gtx) {
		p.Reset()
	}
	size := gtx.Dp(p.Size)
	margin := gtx.Dp(p.Margin)
	if p.Len < 2 {
		p.Len = 3
	}

	progress := float64(p.Anim.Progress(gtx))
	progress = flow((progress*2 - 1), 1000)
	for i := 0; i < p.Len; i++ {
		p.render(gtx, progress, float64(i*(size+margin)), float64(size), float64(margin), i == 0, i == p.Len-1)
	}

	return layout.Dimensions{
		Size: image.Pt((size+margin)*(p.Len-1)+margin, size),
	}
}
