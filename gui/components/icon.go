package components

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/font"
)

func NewIcon(th *material.Theme, gtx layout.Context, ic rune, fg color.NRGBA, s unit.Dp) layout.Dimensions {
	c := gtx.Dp(s)
	gtx.Constraints.Max.X = c
	gtx.Constraints.Max.Y = c
	gtx.Constraints.Min.X = c
	gtx.Constraints.Min.Y = c

	lab := material.Label(th, gtx.Metric.DpToSp(s), string(ic))
	lab.Color = fg
	lab.Font.Typeface = font.SauceCodeProNFMono
	return layout.Stack{
		Alignment: layout.Center,
	}.Layout(
		gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dim := lab.Layout(gtx)
			call := macro.Stop()

			trans := op.Offset(image.Pt(0, -dim.Size.Y/7)).Push(gtx.Ops)
			call.Add(gtx.Ops)
			trans.Pop()
			return dim
		}),
	)
}
