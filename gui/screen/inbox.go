package screen

import (
	"image"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/outlay"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/gui/components"
)

type Inbox struct {
	conf *config.Config
	anim outlay.Animation

	appbar *component.AppBar

	list  widget.List
	items []*components.InboxItem
}

func NewInboxScreen(conf *config.Config) *Inbox {
	inbox := &Inbox{
		conf: conf,
	}

	modal := component.NewModal()
	appbar := component.NewAppBar(modal)
	appbar.Title = "Inbox"
	inbox.appbar = appbar

	inbox.list.List.Axis = layout.Vertical
	return inbox
}

func (p *Inbox) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	animPro := p.anim.Progress(gtx)
	if animPro < 1 {
		gtx.Constraints.Max.Y = int(animPro * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(animPro * float32(gtx.Constraints.Max.X))
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
			return p.appbar.Layout(gtx, th, "Inbox", "...")
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
	p.items = append(p.items, components.NewInboxItem(p.conf, item, in))
}

func (p *Inbox) InAnim() {
	p.anim.Duration = p.conf.AnimTime()
	p.anim.Start(time.Now())
}

func (p *Inbox) Stopped(gtx layout.Context) bool {
	return !p.anim.Animating(gtx)
}
