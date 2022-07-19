package components

import (
	"image"
	"image/color"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"gioui.org/x/outlay"
	"github.com/julioguillermo/jg_sender/config"
)

const (
	BigMargin     = 50
	SmallMargin   = 20
	CornerRadious = 20
)

type InboxItemWidget interface {
	Layout(*material.Theme, layout.Context, *app.Window, *config.Config) layout.Dimensions
}

type InboxItem struct {
	anim outlay.Animation

	In     bool
	Widget InboxItemWidget
}

func NewInboxItem(conf *config.Config, widget InboxItemWidget, in bool) *InboxItem {
	return &InboxItem{
		Widget: widget,
		In:     in,
		anim: outlay.Animation{
			Duration:  conf.AnimTime(),
			StartTime: time.Now(),
		},
	}
}

func (p *InboxItem) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	animPro := p.anim.Progress(gtx)

	marg := layout.UniformInset(10)
	if p.In {
		marg.Right = BigMargin
		marg.Left = SmallMargin
	} else {
		marg.Left = BigMargin
		marg.Right = SmallMargin
	}

	d := marg.Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dim := layout.UniformInset(10).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					return p.Widget.Layout(th, gtx, w, conf)
				},
			)
			call := macro.Stop()

			var col color.NRGBA
			var tails_x float32
			tails_y := float32(dim.Size.Y)
			tails_w := float32(gtx.Dp(SmallMargin))
			tails_h := float32(gtx.Dp(CornerRadious))
			if p.In {
				col = conf.RecivedColor
				tails_x = -float32(gtx.Dp(SmallMargin))
			} else {
				col = conf.SendedColor
				tails_x = float32(dim.Size.X + gtx.Dp(SmallMargin))
				tails_w = -tails_w
			}

			var ops op.Ops
			var tails clip.Path
			tails.Begin(&ops)
			tails.MoveTo(f32.Pt(tails_x+tails_w, tails_y-tails_h))
			tails.QuadTo(f32.Pt(tails_x+tails_w, tails_y), f32.Pt(tails_x, tails_y))
			tails.QuadTo(f32.Pt(tails_x+tails_w, tails_y), f32.Pt(tails_x+tails_w*2, tails_y-tails_h/2))
			tails.Close()

			stack := clip.Outline{Path: tails.End()}.Op().Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			stack.Pop()

			rec := clip.UniformRRect(image.Rect(0, 0, dim.Size.X, dim.Size.Y), gtx.Dp(CornerRadious))
			paint.FillShape(gtx.Ops, col, rec.Op(gtx.Ops))

			call.Add(gtx.Ops)
			return dim
		},
	)

	if animPro < 1 {
		return layout.Dimensions{
			Size: image.Pt(d.Size.X, int(float32(d.Size.Y)*animPro)),
		}
	}

	return d
}
