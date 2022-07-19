package screen

import (
	"image"
	"image/color"
	"net/netip"
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
	"github.com/julioguillermo/jg_sender/connection"
	"github.com/julioguillermo/jg_sender/gui/components"
)

type Subnetworks struct {
	conf    *config.Config
	add     widget.Clickable
	list    widget.List
	subnets []*subnet
	anim    outlay.Animation

	appbar *component.AppBar
	card   *components.Card
}

type subnet struct {
	edit     *components.TextInput
	delete   widget.Clickable
	anim     outlay.Animation
	removing bool
}

func NewSubnetworksScreen(th *material.Theme, conf *config.Config) *Subnetworks {
	sn := &Subnetworks{
		conf: conf,
		card: components.NewSimpleCard(conf.BGColor, 20, 10, 5),
	}
	sn.list.List.Axis = layout.Vertical

	modal := component.NewModal()
	appbar := component.NewAppBar(modal)
	appbar.Title = "Subnetworks"
	appbar.SetActions([]component.AppBarAction{{
		Layout: func(gtx layout.Context, bg, fg color.NRGBA) layout.Dimensions {
			return material.Clickable(gtx, &sn.add, func(gtx layout.Context) layout.Dimensions {
				return components.NewIcon(th, gtx, config.ICNewSubnet, conf.FGPrimaryColor, ScreenBarHeight)
			})
		},
	}}, []component.OverflowAction{})
	sn.appbar = appbar

	subnets := connection.GetIPS()
	for _, s := range subnets {
		sn.New(s.String())
	}

	return sn
}

func (p *subnet) Validator(s string) bool {
	_, e := netip.ParsePrefix(s)
	return e == nil
}

func (p *Subnetworks) New(sn string) *subnet {
	subnet := &subnet{
		anim: outlay.Animation{
			Duration:  p.conf.AnimTime(),
			StartTime: time.Now(),
		},
	}
	subnet.edit = components.NewTextInput("Subnetwork", false)
	subnet.edit.SetText(sn)
	subnet.edit.Validator = subnet.Validator
	p.subnets = append(p.subnets, subnet)
	return subnet
}

func (p *Subnetworks) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if p.add.Clicked() {
		p.New("")
	}

	animPro := p.anim.Progress(gtx)
	if animPro < 1 {
		gtx.Constraints.Max.Y = int(animPro * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(animPro * float32(gtx.Constraints.Max.X))
	}
	gtx.Constraints.Min = gtx.Constraints.Max

	for i, sn := range p.subnets {
		if sn.removing {
			if !sn.anim.Animating(gtx) {
				p.subnets = append(p.subnets[:i], p.subnets[i+1:]...)
			}
		} else {
			if sn.delete.Clicked() {
				sn.removing = true
				sn.anim.Start(time.Now())
			}
		}
	}

	rec := clip.Rect{
		Min: image.Pt(0, 0),
		Max: gtx.Constraints.Max,
	}
	paint.FillShape(gtx.Ops, conf.ScreenColor, rec.Op())

	p.card.Color = conf.BGColor
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.appbar.Layout(gtx, th, "Subnetworks", "...")
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				len(p.subnets),
				func(gtx layout.Context, index int) layout.Dimensions {
					return p.render(th, gtx, w, conf, index)
				},
			)
		}),
	)
}

func (p *Subnetworks) render(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, index int) layout.Dimensions {
	subnet := p.subnets[index]
	animPro := subnet.anim.Progress(gtx)
	if animPro < 1 {
		if subnet.removing {
			animPro = 1 - animPro
		} else {
			p.list.Position.Offset = p.list.Position.Length
		}
		gtx.Constraints.Max.X = int(animPro * float32(gtx.Constraints.Max.X))
	}

	d := p.card.Layout(
		gtx,
		conf,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{
				Alignment: layout.NE,
			}.Layout(
				gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(10).Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							return subnet.edit.Layout(th, gtx, w, conf)
						},
					)
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					return subnet.delete.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return components.NewIcon(th, gtx, config.ICDelete, conf.DangerColor, 30)
					})
				}),
			)
		},
	)
	d.Size.Y = int(animPro * float32(d.Size.Y))
	return d
}

func (p *Subnetworks) GetSubnets() []*netip.Prefix {
	subnets := []*netip.Prefix{}
	for _, sn := range p.subnets {
		pre, err := netip.ParsePrefix(sn.edit.Text())
		if err == nil {
			subnets = append(subnets, &pre)
		}
	}
	return subnets
}

func (p *Subnetworks) InAnim() {
	p.anim.Duration = p.conf.AnimTime()
	p.anim.Start(time.Now())
}

func (p *Subnetworks) Stopped(gtx layout.Context) bool {
	return !p.anim.Animating(gtx)
}
