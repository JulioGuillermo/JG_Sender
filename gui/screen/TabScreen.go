package screen

import (
	"image"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/outlay"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/gui/components"
)

type TabScreen struct {
	Conf *config.Config

	index  int
	title  string
	change bool
	anim   outlay.Animation

	old Screen
	new Screen

	Height int

	items []*TabItem
}

type TabItem struct {
	Title  string
	Icon   rune
	Screen Screen

	clickable    widget.Clickable
	anim         float32
	notification bool
}

type ItemRender struct {
	parent *TabScreen
	screen int
}

func NewTabScreen(conf *config.Config) *TabScreen {
	tabScreen := &TabScreen{
		Conf:   conf,
		Height: 100,
	}
	return tabScreen
}

func (p *TabScreen) Push(title string, icon rune, screen Screen) {
	p.items = append(p.items, &TabItem{
		Title:  title,
		Icon:   icon,
		Screen: screen,
	})
}

func (p *TabScreen) Layout(th *material.Theme, gtx layout.Context, w *app.Window) layout.Dimensions {
	return layout.Stack{
		Alignment: layout.S,
	}.Layout(
		gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Bottom: unit.Dp(0.75 * float32(p.Height)),
			}.Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					if p.Changed() || p.new == nil {
						p.old = p.new
						p.new = p.Screen()
						if p.old != p.new {
							p.new.InAnim()
						}
					}
					if p.old == nil || p.new.Stopped(gtx) {
						return p.new.Layout(th, gtx, w, p.Conf)
					}
					return layout.Stack{
						Alignment: layout.S,
					}.Layout(
						gtx,
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							return p.old.Layout(th, gtx, w, p.Conf)
						}),
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							return p.new.Layout(th, gtx, w, p.Conf)
						}),
					)
				},
			)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			p.change = false
			children := []layout.FlexChild{}

			for i := range p.items {
				render := ItemRender{p, i}
				children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Stack{
						Alignment: layout.S,
					}.Layout(
						gtx,
						layout.Expanded(func(gtx layout.Context) layout.Dimensions {
							return render.parent.items[render.screen].Layout(th, gtx, w, p.Conf, render.screen, render.parent)
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
			paint.FillShape(gtx.Ops, p.Conf.BGPrimaryColor, rec.Op())

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
		}),
	)
}

func (p *TabItem) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, index int, parent *TabScreen) layout.Dimensions {
	if p.clickable.Clicked() {
		parent.change = true
		parent.index = index
		parent.title = p.Title
		parent.anim.Duration = conf.AnimTime()
		parent.anim.Start(time.Now())
	}

	if parent.index == index {
		p.anim = parent.anim.Progress(gtx)
	} else {
		if p.anim > 0 {
			p.anim = 1 - parent.anim.Progress(gtx)
		}
	}

	trans := uint8(p.anim * 255)
	size := int(p.anim*float32(gtx.Constraints.Max.Y)/4 + float32(gtx.Constraints.Max.Y)*3/4)
	bmargin := gtx.Metric.Dp(5)

	bg_color := conf.FGPrimaryColor
	bg_color.A = trans
	item_color := components.ColorTransition(conf.FGPrimaryColor, conf.BGPrimaryColor, p.anim)

	return p.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
				return components.NewIcon(th, gtx, p.Icon, item_color, unit.Dp(float32(size)/gtx.Metric.PxPerDp))
			},
		)
	})
}

func (p *TabScreen) Changed() bool {
	return p.change
}

func (p *TabScreen) Screen() Screen {
	return p.items[p.index].Screen
}

func (p *TabScreen) ScreenIndex() int {
	return p.index
}

func (p *TabScreen) Notify(s int, n bool) {
	if s < 0 || s >= len(p.items) {
		return
	}
	p.items[s].notification = n
}

func (p *TabScreen) Change(s int) {
	if s < 0 || s >= len(p.items) {
		return
	}
	p.change = true
	p.index = s
	p.title = p.items[s].Title
}
