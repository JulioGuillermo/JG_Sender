package screen

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
	"github.com/julioguillermo/jg_sender/gui/components"
)

type Inbox struct {
	conf *config.Config
	anim float32

	list  widget.List
	items []*components.InboxItem
}

func NewInboxScreen(conf *config.Config) *Inbox {
	inbox := &Inbox{
		conf: conf,
	}
	inbox.list.List.Axis = layout.Vertical
	return inbox
}

func (p *Inbox) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if p.anim < 1 {
		p.anim += components.AnimSpeed(gtx) / 2
		if p.anim > 1 {
			p.anim = 1
		}
		op.InvalidateOp{At: gtx.Now.Add(components.Time(gtx))}.Add(gtx.Ops)
		gtx.Constraints.Max.Y = int(p.anim * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(p.anim * float32(gtx.Constraints.Max.X))
	}
	gtx.Constraints.Min = gtx.Constraints.Max

	rec := clip.Rect{
		Min: image.Pt(0, 0),
		Max: gtx.Constraints.Max,
	}
	paint.FillShape(gtx.Ops, conf.ScreenColor, rec.Op())

	p.list.ScrollToEnd = !p.list.Position.BeforeEnd
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEnd,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			rec := clip.Rect{
				Min: image.Pt(0, 0),
				Max: image.Pt(gtx.Constraints.Max.X, gtx.Dp(ScreenBarHeight)),
			}
			paint.FillShape(gtx.Ops, conf.BGPrimaryColor, rec.Op())

			title := material.Label(th, gtx.Metric.DpToSp(ScreenBarHeight-TitleMargin), "Inbox")
			title.Color = conf.FGPrimaryColor
			d := layout.Inset{
				Left: 10,
			}.Layout(
				gtx,
				title.Layout,
			)
			d.Size.Y = gtx.Dp(ScreenBarHeight)
			return d
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				len(p.items),
				func(gtx layout.Context, index int) layout.Dimensions {
					d := p.items[index].Layout(th, gtx, w, conf)
					return d
				},
			)
		}),
	)
}

func (p *Inbox) NewInbox(item components.InboxItemWidget, in bool) {
	p.items = append(p.items, components.NewInboxItem(item, in))
}

func (p *Inbox) InAnim() {
	p.anim = 0
}

func (p *Inbox) Stopped() bool {
	return p.anim == 1
}
