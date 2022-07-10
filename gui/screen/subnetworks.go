package screen

import (
	"image"
	"net/netip"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/connection"
	"github.com/julioguillermo/jg_sender/gui/components"
)

type Subnetworks struct {
	conf    *config.Config
	add     widget.Clickable
	list    widget.List
	subnets []*subnet
	anim    float32
}

type subnet struct {
	edit     *components.TextInput
	delete   widget.Clickable
	anim     float32
	removing bool
	dim      layout.Dimensions
}

func NewSubnetworksScreen(conf *config.Config) *Subnetworks {
	sn := &Subnetworks{
		conf: conf,
		anim: 1,
	}
	sn.list.List.Axis = layout.Vertical

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
		anim: 0,
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
		p.list.Position.Offset = p.list.Position.Length
	}

	if p.anim < 1 {
		p.anim += conf.AnimSpeed(gtx)
		if p.anim > 1 {
			p.anim = 1
		}
		op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
		gtx.Constraints.Max.Y = int(p.anim * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(p.anim * float32(gtx.Constraints.Max.X))
	}
	gtx.Constraints.Min = gtx.Constraints.Max

	rec := clip.Rect{
		Min: image.Pt(0, 0),
		Max: gtx.Constraints.Max,
	}
	paint.FillShape(gtx.Ops, conf.ScreenColor, rec.Op())

	for i, sn := range p.subnets {
		if sn.removing {
			if sn.anim == 0 {
				p.subnets = append(p.subnets[:i], p.subnets[i+1:]...)
			}
		} else {
			if sn.delete.Clicked() {
				sn.removing = true
			}
		}
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			rec := clip.Rect{
				Min: image.Pt(0, 0),
				Max: image.Pt(gtx.Constraints.Max.X, gtx.Dp(ScreenBarHeight)),
			}
			paint.FillShape(gtx.Ops, conf.BGPrimaryColor, rec.Op())

			return layout.Flex{
				Axis:      layout.Horizontal,
				Spacing:   layout.SpaceBetween,
				Alignment: layout.Middle,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.Label(th, gtx.Metric.DpToSp(ScreenBarHeight-TitleMargin), "Subnetworks")
					title.Color = conf.FGPrimaryColor
					return layout.Inset{
						Left: 10,
					}.Layout(
						gtx,
						title.Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &p.add, func(gtx layout.Context) layout.Dimensions {
						return components.NewIcon(th, gtx, config.ICNewSubnet, conf.FGPrimaryColor, ScreenBarHeight)
					})
				}),
			)
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
	if subnet.removing {
		if subnet.anim > 0 {
			subnet.anim -= conf.AnimSpeed(gtx)
			if subnet.anim < 0 {
				subnet.anim = 0
			}
			op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
			gtx.Constraints.Max.X = int(subnet.anim * float32(gtx.Constraints.Max.X))
		}
	} else {
		if subnet.anim < 1 {
			subnet.anim += conf.AnimSpeed(gtx)
			if subnet.anim > 1 {
				subnet.anim = 1
			}
			p.list.Position.Offset = p.list.Position.Length
			op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
			gtx.Constraints.Max.X = int(subnet.anim * float32(gtx.Constraints.Max.X))
		}
	}

	d := layout.UniformInset(10).Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			card := clip.UniformRRect(image.Rect(0, 0, gtx.Constraints.Max.X, subnet.dim.Size.Y), gtx.Dp(20))
			paint.FillShape(gtx.Ops, conf.BGColor, card.Op(gtx.Ops))

			subnet.dim = layout.UniformInset(5).Layout(
				gtx,
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
			return subnet.dim
		},
	)
	d.Size.Y = int(subnet.anim * float32(d.Size.Y))
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
	p.anim = 0
}

func (p *Subnetworks) Stopped() bool {
	return p.anim == 1
}
