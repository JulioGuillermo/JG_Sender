package components

import (
	"image"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
)

type Tab struct {
	screen int
	title  string
	Items  []*TabItem
	Height int
	change bool
}

type TabItem struct {
	Clickable    *widget.Clickable
	Title        string
	Icon         rune
	anim         float32
	notification bool
}

type ItemRender struct {
	parent *Tab
	screen int
}

func NewTab(def int, items ...*TabItem) *Tab {
	items[def].anim = 1
	return &Tab{
		screen: def,
		Items:  items,
		Height: 100,
		change: false,
	}
}

func NewTabItem(title string, icon rune) *TabItem {
	return &TabItem{
		Clickable: &widget.Clickable{},
		Title:     title,
		Icon:      icon,
		anim:      0,
	}
}

func (p *Tab) Changed() bool {
	return p.change
}

func (p *Tab) Screen() (int, string) {
	return p.screen, p.title
}

func (p *Tab) ScreenIndex() int {
	return p.screen
}

func (p *Tab) Notify(s int, n bool) {
	if s < 0 || s >= len(p.Items) {
		return
	}
	p.Items[s].notification = n
}

func (p *Tab) Change(s int) {
	if s < 0 || s >= len(p.Items) {
		return
	}
	p.change = true
	p.screen = s
	p.title = p.Items[s].Title
}

func (p *Tab) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	p.change = false
	children := []layout.FlexChild{}

	for i := range p.Items {
		render := ItemRender{p, i}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{
				Alignment: layout.S,
			}.Layout(
				gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return render.parent.Items[render.screen].Layout(th, gtx, w, conf, render.screen, render.parent)
				}),
			)
		}))
	}

	size := int(float32(p.Height) * gtx.Metric.PxPerDp)
	tmargin := int(20 * gtx.Metric.PxPerDp)
	gtx.Constraints.Max.Y = size
	gtx.Constraints.Min.Y = size

	rec := clip.Rect{
		Min: image.Pt(0, size/4),
		Max: image.Pt(gtx.Constraints.Max.X, size),
	}
	paint.FillShape(gtx.Ops, conf.BGPrimaryColor, rec.Op())

	return layout.UniformInset(10).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.Y = size - tmargin
		gtx.Constraints.Min.Y = size - tmargin
		d := layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.End,
		}.Layout(
			gtx,
			children...,
		)
		d.Size.Y = size - tmargin
		return d
	})
}

func (p *TabItem) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, screen int, parent *Tab) layout.Dimensions {
	speed := conf.AnimSpeed(gtx)

	if p.Clickable.Clicked() {
		parent.change = true
		parent.screen = screen
		parent.title = p.Title
	}

	if parent.screen == screen {
		if p.anim < 1 {
			p.anim += speed
			op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
		}
		if p.anim > 1 {
			p.anim = 1
		}
	} else {
		if p.anim > 0 {
			p.anim -= speed
		}
		if p.anim < 0 {
			p.anim = 0
		}
	}

	trans := uint8(p.anim * 255)
	size := int(p.anim*float32(gtx.Constraints.Max.Y)/4 + float32(gtx.Constraints.Max.Y)*3/4)
	bmargin := gtx.Metric.Dp(5)

	bg_color := conf.FGPrimaryColor
	bg_color.A = trans
	item_color := ColorTransition(conf.FGPrimaryColor, conf.BGPrimaryColor, p.anim)

	return p.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		bg := clip.Ellipse{
			Min: image.Pt(0, 0),
			Max: image.Pt(size, size),
		}
		paint.FillShape(gtx.Ops, conf.BGPrimaryColor, bg.Op(gtx.Ops))

		bg.Min.X += bmargin
		bg.Min.Y += bmargin
		bg.Max.X -= bmargin
		bg.Max.Y -= bmargin
		paint.FillShape(gtx.Ops, bg_color, bg.Op(gtx.Ops))

		if p.notification {
			not := clip.Ellipse{
				Min: image.Pt(size/2-gtx.Dp(5), 0),
				Max: image.Pt(size/2+gtx.Dp(5), gtx.Dp(10)),
			}
			paint.FillShape(gtx.Ops, conf.DangerColor, not.Op(gtx.Ops))
		}

		return layout.UniformInset(0).Layout(
			gtx,
			func(gtx layout.Context) layout.Dimensions {
				return NewIcon(th, gtx, p.Icon, item_color, unit.Dp(float32(size)/gtx.Metric.PxPerDp))
			},
		)
	})
}
