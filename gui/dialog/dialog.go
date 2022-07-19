package dialog

import (
	"image"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/outlay"
	"github.com/julioguillermo/jg_sender/config"
)

type Dialog struct {
	Conf     *config.Config
	anim     outlay.Animation
	removing bool
	cancel   widget.Clickable
	Widget   func(*material.Theme, layout.Context, *app.Window, *config.Config) layout.Dimensions
}

func (p *Dialog) SetWidget(widget func(*material.Theme, layout.Context, *app.Window, *config.Config) layout.Dimensions) {
	p.Widget = widget
	p.anim.Duration = p.Conf.AnimTime()
	p.anim.Start(time.Now())
	p.removing = false
}

func (p *Dialog) RemoveWidget() {
	p.removing = true
}

func (p *Dialog) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if p.Widget == nil {
		return layout.Dimensions{}
	}

	animPro := p.anim.Progress(gtx)
	if animPro < 1 {
		if p.removing {
			animPro = 1 - animPro
			if animPro <= 0 {
				p.Widget = nil
				return layout.Dimensions{}
			}
		}
		gtx.Constraints.Max.Y = int(animPro * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(animPro * float32(gtx.Constraints.Max.X))
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
			macro := op.Record(gtx.Ops)
			dim := layout.UniformInset(10).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					return p.Widget(th, gtx, w, conf)
				},
			)
			call := macro.Stop()

			rec := clip.UniformRRect(image.Rect(0, 0, dim.Size.X, dim.Size.Y), gtx.Dp(20))
			paint.FillShape(gtx.Ops, conf.BGColor, rec.Op(gtx.Ops))

			call.Add(gtx.Ops)
			return dim
		}),
	)
}
