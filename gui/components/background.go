package components

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type Background struct {
	Color color.NRGBA

	PaddingLeft   unit.Dp
	PaddingTop    unit.Dp
	PaddingRight  unit.Dp
	PaddingBottom unit.Dp
}

func NewBackground(color color.NRGBA, paddingLeft, paddingTop, paddingRight, paddingBottom unit.Dp) *Background {
	bg := &Background{
		Color:         color,
		PaddingLeft:   paddingLeft,
		PaddingTop:    paddingTop,
		PaddingRight:  paddingRight,
		PaddingBottom: paddingBottom,
	}
	return bg
}

func (p *Background) Layout(gtx layout.Context, widget func(layout.Context) layout.Dimensions) layout.Dimensions {
	cgtx := gtx

	paddingTop := gtx.Dp(p.PaddingTop)
	paddingBottom := gtx.Dp(p.PaddingBottom)
	paddingLeft := gtx.Dp(p.PaddingLeft)
	paddingRight := gtx.Dp(p.PaddingRight)

	mcs := cgtx.Constraints
	mcs.Max.X -= +paddingLeft + paddingRight
	if mcs.Max.X < 0 {
		paddingLeft = 0
		paddingRight = 0
		mcs.Max.X = 0
	}
	if mcs.Min.X > mcs.Max.X {
		mcs.Min.X = mcs.Max.X
	}
	mcs.Max.Y -= paddingTop + paddingBottom
	if mcs.Max.Y < 0 {
		paddingTop = 0
		paddingBottom = 0
		mcs.Max.Y = 0
	}
	if mcs.Min.Y > mcs.Max.Y {
		mcs.Min.Y = mcs.Max.Y
	}
	cgtx.Constraints = mcs

	macro := op.Record(gtx.Ops)
	trans := op.Offset(image.Pt(paddingLeft, paddingTop)).Push(cgtx.Ops)
	dims := widget(cgtx)
	trans.Pop()
	call := macro.Stop()

	rec := clip.Rect{
		Min: image.Pt(0, 0),
		Max: image.Pt(dims.Size.X+paddingLeft+paddingRight, dims.Size.Y+paddingTop+paddingBottom),
	}
	paint.FillShape(gtx.Ops, p.Color, rec.Op())

	call.Add(gtx.Ops)
	return layout.Dimensions{
		Size: image.Pt(dims.Size.X+paddingLeft+paddingRight, dims.Size.Y+paddingTop+paddingBottom),
	}
}
