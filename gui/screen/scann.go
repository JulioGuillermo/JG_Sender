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
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/outlay"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/connection"
	"github.com/julioguillermo/jg_sender/gui/components"
)

type SNSource interface {
	GetSubnets() []*netip.Prefix
}

type Scanner struct {
	scanner  *connection.Scanner
	conf     *config.Config
	src      SNSource
	scan     widget.Clickable
	list     widget.List
	devices  []*found
	anim     outlay.Animation
	progress float64
	win      *app.Window

	msg  *components.MSGDialog
	file *components.FileDialog

	newInbox func(components.InboxItemWidget, bool)

	card   *components.Card
	appbar *component.AppBar

	layoutH layout.Flex
	layoutV layout.Flex
}

type found struct {
	name     string
	addr     *netip.Addr
	os       string
	dim      layout.Dimensions
	SendMSG  widget.Clickable
	SendFile widget.Clickable
	Anim     outlay.Animation
}

func NewScannerScreen(th *material.Theme, conf *config.Config, src SNSource, w *app.Window, newInbox func(components.InboxItemWidget, bool)) *Scanner {
	sn := &Scanner{
		scanner:  connection.NewScanner(conf),
		src:      src,
		conf:     conf,
		progress: -1,
		win:      w,
		newInbox: newInbox,
		card:     components.NewSimpleCard(conf.BGColor, 20, 10, 10),
	}

	modal := component.NewModal()
	appbar := component.NewAppBar(modal)
	appbar.Title = "Scanner"
	appbar.SetActions([]component.AppBarAction{{
		Layout: func(gtx layout.Context, bg, fg color.NRGBA) layout.Dimensions {
			return material.Clickable(gtx, &sn.scan, func(gtx layout.Context) layout.Dimensions {
				if sn.scanner.Running {
					c := material.ProgressCircle(th, float32(sn.progress))
					c.Color = conf.FGPrimaryColor
					return layout.Stack{
						Alignment: layout.Center,
					}.Layout(
						gtx,
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							size := gtx.Dp(ScreenBarHeight)
							gtx.Constraints.Min.X = size
							gtx.Constraints.Min.Y = size
							gtx.Constraints.Max.X = size
							gtx.Constraints.Max.Y = size
							return layout.UniformInset(5).Layout(
								gtx,
								func(gtx layout.Context) layout.Dimensions {
									return c.Layout(gtx)
								},
							)
						}),
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							return components.NewIcon(th, gtx, config.ICScanStop, conf.FGPrimaryColor, ScreenBarHeight-20)
						}),
					)
				}
				return components.NewIcon(th, gtx, config.ICScan, conf.FGPrimaryColor, ScreenBarHeight)
			})
		},
	}}, []component.OverflowAction{})
	sn.appbar = appbar

	sn.layoutH.Alignment = layout.Middle
	sn.layoutH.Axis = layout.Horizontal
	sn.layoutV.Axis = layout.Vertical

	sn.list.List.Axis = layout.Vertical
	sn.scanner.Found = sn.OnFound
	sn.scanner.Progress = sn.OnProgress

	return sn
}

func (p *Scanner) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if p.scan.Clicked() {
		if p.scanner.Running {
			p.scanner.Stop()
		} else {
			p.devices = []*found{}
			go p.scanner.ScannAll(p.src.GetSubnets())
		}
	}

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

	p.card.Color = conf.BGColor

	return p.layoutV.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.appbar.Layout(gtx, th, "Connections", "...")
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				len(p.devices),
				func(gtx layout.Context, index int) layout.Dimensions {
					return p.render(th, gtx, w, conf, index)
				},
			)
		}),
	)
}

func (p *Scanner) render(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, index int) layout.Dimensions {
	const (
		// In sp
		title_size = 25
		info_size  = 17
		// In dp
		space_between = 10
		insets        = 10
		ext_space     = 20
	)

	device := p.devices[index]
	if device.SendMSG.Clicked() {
		p.SendMSG(device.addr, device.name)
	} else if device.SendFile.Clicked() {
		p.SendFile(device.addr, device.name)
	}

	animPro := device.Anim.Progress(gtx)
	if animPro < 1 {
		gtx.Constraints.Max.X = int(animPro * float32(gtx.Constraints.Max.X))
	}

	d := p.card.Layout(
		gtx,
		conf,
		func(gtx layout.Context) layout.Dimensions {
			return p.layoutH.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return components.NewIcon(th, gtx, p.GetOSIcon(device.os), conf.BGPrimaryColor, unit.Dp(float32(device.dim.Size.Y)/gtx.Metric.PxPerDp))
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					device.dim = p.layoutV.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							d := layout.Flex{
								Axis: layout.Horizontal,
							}.Layout(
								gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									name := material.Label(th, title_size, device.name)
									name.Color = conf.BGPrimaryColor
									name.Font.Weight = text.Bold
									return name.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return device.SendMSG.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return components.NewIcon(th, gtx, config.ICMSG, conf.BGPrimaryColor, 30)
									})
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return device.SendFile.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return components.NewIcon(th, gtx, config.ICFile, conf.BGPrimaryColor, 30)
									})
								}),
							)
							rec := clip.Rect{
								Min: image.Pt(0, d.Size.Y),
								Max: image.Pt(d.Size.X, d.Size.Y+gtx.Dp(2)),
							}
							paint.FillShape(gtx.Ops, conf.BGPrimaryColor, rec.Op())
							return d
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top: space_between,
							}.Layout(
								gtx,
								material.Label(th, info_size, device.os+" - "+device.addr.String()).Layout,
							)
						}),
					)
					return device.dim
				}),
			)
		},
	)
	d.Size.Y = int(animPro * float32(d.Size.Y))
	return d
}

func (p *Scanner) GetOSIcon(os string) rune {
	switch os {
	case "android":
		return config.ICAndroid
	case "ios":
		return config.ICApple
	case "windows":
		return config.ICWindows
	case "linux":
		return config.ICLinux
	}
	return config.ICUnknow
}

func (p *Scanner) InAnim() {
	p.anim.Duration = p.conf.AnimTime()
	p.anim.Start(time.Now())
}

func (p *Scanner) Stopped(gtx layout.Context) bool {
	return !p.anim.Animating(gtx)
}

func (p *Scanner) OnFound(addr *netip.Addr, name, device string) {
	dev := &found{
		addr: addr,
		os:   device,
		name: name,
		Anim: outlay.Animation{
			Duration:  p.conf.AnimTime(),
			StartTime: time.Now(),
		},
	}
	p.devices = append(p.devices, dev)
}

func (p *Scanner) OnProgress(pro float64) {
	p.progress = pro
	p.win.Invalidate()
}

func (p *Scanner) SendMSG(addr *netip.Addr, to string) {
	p.msg = components.NewMSGDialog(addr, to, p.newInbox)
	p.conf.OpenDialog(p.msg.Layout)
}

func (p *Scanner) SendFile(addr *netip.Addr, to string) {
	p.file = components.NewFileDialog(addr, to, p.newInbox)
	p.conf.OpenDialog(p.file.Layout)
}
