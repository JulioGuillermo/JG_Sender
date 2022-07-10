package dialog

import (
	"image"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
)

type Dialog struct {
	anim     float32
	removing bool
	cancel   widget.Clickable
	Widget   func(*material.Theme, layout.Context, *app.Window, *config.Config) layout.Dimensions
}

func (p *Dialog) SetWidget(widget func(*material.Theme, layout.Context, *app.Window, *config.Config) layout.Dimensions) {
	p.Widget = widget
	p.anim = 0
	p.removing = false
}

func (p *Dialog) RemoveWidget() {
	p.removing = true
}

func (p *Dialog) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if p.Widget == nil {
		return layout.Dimensions{}
	}

	if p.removing {
		if p.anim > 0 {
			p.anim -= conf.AnimSpeed(gtx)
			if p.anim <= 0 {
				p.anim = 0
				p.Widget = nil
				return layout.Dimensions{}
			}
			op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
		}
		gtx.Constraints.Max.Y = int(p.anim * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(p.anim * float32(gtx.Constraints.Max.X))
	} else {
		if p.anim < 1 {
			p.anim += conf.AnimSpeed(gtx)
			if p.anim > 1 {
				p.anim = 1
			}
			op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
		}
		gtx.Constraints.Max.Y = int(p.anim * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(p.anim * float32(gtx.Constraints.Max.X))
	}

	return layout.Stack{
		Alignment: layout.Center,
	}.Layout(
		gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return p.cancel.Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					rec := clip.Rect{
						Min: image.Pt(0, 0),
						Max: gtx.Constraints.Max,
					}
					paint.FillShape(gtx.Ops, conf.Shadow, rec.Op())
					return layout.Dimensions{
						Size:     image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
						Baseline: 0,
					}
				},
			)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return p.Widget(th, gtx, w, conf)
		}),
	)
}
