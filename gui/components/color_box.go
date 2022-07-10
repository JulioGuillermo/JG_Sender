package components

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

const (
	Size = 30
)

func NewColorBox(gtx layout.Context, clickable *widget.Clickable, c color.NRGBA) layout.Dimensions {
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := image.Pt(gtx.Dp(Size), gtx.Dp(Size))
		rec := clip.Rect{
			Min: image.Pt(0, 0),
			Max: size,
		}
		paint.FillShape(gtx.Ops, color.NRGBA{255, 255, 255, 255}, rec.Op())
		rec.Min.X++
		rec.Min.Y++
		rec.Max.X--
		rec.Max.Y--
		paint.FillShape(gtx.Ops, color.NRGBA{0, 0, 0, 255}, rec.Op())
		rec.Min.X++
		rec.Min.Y++
		rec.Max.X--
		rec.Max.Y--
		paint.FillShape(gtx.Ops, c, rec.Op())
		return layout.Dimensions{Size: size}
	})
}
