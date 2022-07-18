package components

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/julioguillermo/jg_sender/config"
)

type Card struct {
	Color  color.NRGBA
	Radius unit.Dp

	MarginLeft   unit.Dp
	MarginTop    unit.Dp
	MarginRight  unit.Dp
	MarginBottom unit.Dp

	PaddingLeft   unit.Dp
	PaddingTop    unit.Dp
	PaddingRight  unit.Dp
	PaddingBottom unit.Dp
}

func NewCard(color color.NRGBA, radius, marginLeft, marginTop, marginRight, marginBottom, paddingLeft, paddingTop, paddingRight, paddingBottom unit.Dp) *Card {
	card := &Card{
		Color:  color,
		Radius: radius,

		MarginLeft:   marginLeft,
		MarginTop:    marginTop,
		MarginRight:  marginRight,
		MarginBottom: marginBottom,

		PaddingLeft:   paddingLeft,
		PaddingTop:    paddingTop,
		PaddingRight:  paddingRight,
		PaddingBottom: paddingBottom,
	}
	return card
}

func NewSimpleCard(color color.NRGBA, radius, margin, padding unit.Dp) *Card {
	return NewCard(color, radius, margin, margin, margin, margin, padding, padding, padding, padding)
}

func (p *Card) Layout(gtx layout.Context, conf *config.Config, widget func(layout.Context) layout.Dimensions) layout.Dimensions {
	cgtx := gtx

	marginTop := gtx.Dp(p.MarginTop)
	marginBottom := gtx.Dp(p.MarginBottom)
	marginLeft := gtx.Dp(p.MarginLeft)
	marginRight := gtx.Dp(p.MarginRight)

	paddingTop := gtx.Dp(p.PaddingTop)
	paddingBottom := gtx.Dp(p.PaddingBottom)
	paddingLeft := gtx.Dp(p.PaddingLeft)
	paddingRight := gtx.Dp(p.PaddingRight)

	mcs := cgtx.Constraints
	mcs.Max.X -= marginLeft + marginRight + paddingLeft + paddingRight
	if mcs.Max.X < 0 {
		marginLeft = 0
		marginRight = 0
		paddingLeft = 0
		paddingRight = 0
		mcs.Max.X = 0
	}
	if mcs.Min.X > mcs.Max.X {
		mcs.Min.X = mcs.Max.X
	}
	mcs.Max.Y -= marginTop + marginBottom + paddingTop + paddingBottom
	if mcs.Max.Y < 0 {
		marginTop = 0
		marginBottom = 0
		paddingTop = 0
		paddingBottom = 0
		mcs.Max.Y = 0
	}
	if mcs.Min.Y > mcs.Max.Y {
		mcs.Min.Y = mcs.Max.Y
	}
	cgtx.Constraints = mcs

	macro := op.Record(cgtx.Ops)
	trans := op.Offset(image.Pt(marginLeft+paddingLeft, marginTop+paddingTop)).Push(cgtx.Ops)
	dims := widget(cgtx)
	trans.Pop()
	call := macro.Stop()

	trans = op.Offset(image.Pt(marginLeft, marginTop)).Push(gtx.Ops)
	rec := clip.UniformRRect(image.Rect(0, 0, dims.Size.X+paddingLeft+paddingRight, dims.Size.Y+paddingTop+paddingBottom), gtx.Dp(p.Radius))
	paint.FillShape(gtx.Ops, p.Color, rec.Op(gtx.Ops))
	trans.Pop()

	call.Add(gtx.Ops)
	return layout.Dimensions{
		Size: image.Pt(dims.Size.X+paddingLeft+paddingRight+marginLeft+marginRight, dims.Size.Y+paddingTop+paddingBottom+marginTop+marginBottom),
	}
}
