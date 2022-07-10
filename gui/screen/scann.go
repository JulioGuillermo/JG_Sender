package screen

import (
	"image"
	"net/netip"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
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
	anim     float32
	progress float64
	win      *app.Window

	msg  *components.MSGDialog
	file *components.FileDialog

	newInbox func(components.InboxItemWidget, bool)
}

type found struct {
	name     string
	addr     *netip.Addr
	os       string
	anim     float32
	dim      layout.Dimensions
	SendMSG  widget.Clickable
	SendFile widget.Clickable
}

func NewScannerScreen(conf *config.Config, src SNSource, w *app.Window, newInbox func(components.InboxItemWidget, bool)) *Scanner {
	sn := &Scanner{
		scanner:  connection.NewScanner(conf),
		src:      src,
		conf:     conf,
		anim:     1,
		progress: -1,
		win:      w,
		newInbox: newInbox,
	}

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
					title := material.Label(th, gtx.Metric.DpToSp(ScreenBarHeight-TitleMargin), "Connections")
					title.Color = conf.FGPrimaryColor
					return layout.Inset{
						Left: 10,
					}.Layout(
						gtx,
						title.Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &p.scan, func(gtx layout.Context) layout.Dimensions {
						if p.scanner.Running {
							c := material.ProgressCircle(th, float32(p.progress))
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
				}),
			)
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

	if device.anim < 1 {
		device.anim += conf.AnimSpeed(gtx)
		if device.anim > 1 {
			device.anim = 1
		}
		op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
		gtx.Constraints.Max.X = int(device.anim * float32(gtx.Constraints.Max.X))
	}

	d := layout.UniformInset(10).Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			card := clip.UniformRRect(image.Rect(0, 0, gtx.Constraints.Max.X, device.dim.Size.Y+gtx.Dp(insets*2)), gtx.Dp(20))
			paint.FillShape(gtx.Ops, conf.BGColor, card.Op(gtx.Ops))

			return layout.UniformInset(insets).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Start,
					}.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Left: 5,
							}.Layout(
								gtx,
								func(gtx layout.Context) layout.Dimensions {
									return components.NewIcon(th, gtx, p.GetOSIcon(device.os), conf.BGPrimaryColor, unit.Dp(float32(device.dim.Size.Y)/gtx.Metric.PxPerDp))
								},
							)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							device.dim = layout.Flex{
								Axis: layout.Vertical,
							}.Layout(
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
		},
	)
	d.Size.Y = int(device.anim * float32(d.Size.Y))
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
	p.anim = 0
}

func (p *Scanner) Stopped() bool {
	return p.anim == 1
}

func (p *Scanner) OnFound(addr *netip.Addr, name, device string) {
	dev := &found{
		addr: addr,
		os:   device,
		name: name,
		anim: 0,
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
