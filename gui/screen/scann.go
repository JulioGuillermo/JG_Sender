package screen

import (
	"fmt"
	"image"
	"image/color"
	"net/netip"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/notify"
	"gioui.org/x/outlay"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/connection"
	"github.com/julioguillermo/jg_sender/gui/components"
)

type SNSource interface {
	GetSubnets() []*netip.Prefix
}

type Scanner struct {
	Notification map[string][]notify.Notification

	scanner  *connection.Scanner
	conf     *config.Config
	src      SNSource
	scan     widget.Clickable
	list     widget.List
	devices  []*found
	anim     outlay.Animation
	progress float64
	win      *app.Window

	loading_anim_show outlay.Animation
	loading_anim      *components.LoadingAnim
	loading_visible   bool
	loading_ctl       bool

	card   *components.Card
	appbar *component.AppBar
	modal  *component.ModalLayer

	layoutH layout.Flex
	layoutV layout.Flex

	OnOpen func(string)
}

type found struct {
	open widget.Clickable
	Anim outlay.Animation
}

func NewScannerScreen(th *material.Theme, conf *config.Config, src SNSource, w *app.Window) *Scanner {
	sn := &Scanner{
		scanner:  connection.NewScanner(conf),
		src:      src,
		conf:     conf,
		progress: -1,
		win:      w,
		card:     components.NewSimpleCard(conf.BGColor, 20, 10, 10),
	}
	sn.loading_anim = components.NewLoadingAnim(20, 5, 5, time.Second, conf.BGPrimaryColor)
	sn.loading_anim.Reset()

	modal := component.NewModal()
	sn.modal = modal
	appbar := component.NewAppBar(modal)
	appbar.Title = "Scanner"
	appbar.SetActions([]component.AppBarAction{{
		OverflowAction: component.OverflowAction{
			Name: "Scanner ON/OFF",
			Tag:  &sn.scan,
		},
		Layout: func(gtx layout.Context, bg, fg color.NRGBA) layout.Dimensions {
			bls := material.ButtonLayout(th, &sn.scan)
			bls.CornerRadius = ScreenBarHeight / 2
			return bls.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
							return components.NewIcon(th, gtx, config.ICClose, conf.FGPrimaryColor, ScreenBarHeight-20)
						}),
					)
				}
				return components.NewIcon(th, gtx, config.ICUpdate, conf.FGPrimaryColor, ScreenBarHeight)
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
	for _, e := range p.appbar.Events(gtx) {
		t, ok := e.(component.AppBarOverflowActionClicked)
		if ok && t.Tag == &p.scan {
			if p.scanner.Running {
				p.scanner.Stop()
			} else {
				connection.InvalidateDevices()
				go p.scanner.ScannAll(p.src.GetSubnets())
			}
		}
	}
	if p.scan.Clicked() {
		if p.scanner.Running {
			p.scanner.Stop()
		} else {
			p.loading_anim.Reset()
			connection.InvalidateDevices()
			go p.scanner.ScannAll(p.src.GetSubnets())
		}
	}

	if p.scanner.Running && !p.loading_ctl {
		p.loading_anim_show.Start(gtx.Now)
		p.loading_anim_show.Duration = p.conf.AnimTime()
		p.loading_visible = true
		p.loading_ctl = true
	} else if !p.scanner.Running && p.loading_ctl {
		p.loading_ctl = false
		p.loading_anim_show.Start(gtx.Now)
		p.loading_anim_show.Duration = p.conf.AnimTime()
	} else if !p.scanner.Running && !p.loading_ctl && !p.loading_anim_show.Animating(gtx) {
		p.loading_visible = false
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

	d := p.layoutV.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.appbar.Layout(gtx, th, "Connections", "...")
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.loading_visible {
				p.loading_anim.Color = conf.BGPrimaryColor
				d := layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						probar := material.ProgressBar(th, float32(p.progress))
						return probar.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(3).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return p.loading_anim.Layout(gtx)
						})
					}),
				)
				progress := p.loading_anim_show.Progress(gtx)
				if !p.loading_ctl {
					progress = 1 - progress
				}
				d.Size.Y = int(progress * float32(d.Size.Y))
				return d
			}
			return layout.Dimensions{}
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				len(connection.Devices),
				func(gtx layout.Context, index int) layout.Dimensions {
					return p.render(th, gtx, w, conf, index)
				},
			)
		}),
	)
	p.modal.Layout(gtx, th)
	return d
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

	for len(p.devices) < len(connection.Devices) {
		p.devices = append(p.devices, &found{})
	}
	for len(p.devices) > len(connection.Devices) {
		p.devices = p.devices[:len(p.devices)-1]
	}
	device := p.devices[index]
	connDev := connection.Devices[index]
	if device.open.Clicked() {
		p.Open(connDev.ID)
	}

	animPro := device.Anim.Progress(gtx)
	if animPro < 1 {
		gtx.Constraints.Max.X = int(animPro * float32(gtx.Constraints.Max.X))
	}

	d := device.open.Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			return p.card.Layout(
				gtx,
				conf,
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutH.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return components.NewIcon(th, gtx, p.GetOSIcon(connDev.OS), conf.BGPrimaryColor, 60)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return p.layoutV.Layout(
								gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									name := material.Label(th, title_size, connDev.Name)
									name.Color = conf.BGPrimaryColor
									name.Font.Weight = text.Bold
									d := p.layoutH.Layout(
										gtx,
										layout.Flexed(1, name.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											if connDev.Not == 0 {
												return layout.Dimensions{}
											}
											lab := material.Label(th, th.TextSize*0.5, fmt.Sprint(connDev.Not))
											lab.Color = conf.FGPrimaryColor

											macro := op.Record(gtx.Ops)
											dim := layout.UniformInset(3).Layout(gtx, lab.Layout)
											call := macro.Stop()

											r := dim.Size.X
											if r > dim.Size.Y {
												r = dim.Size.Y
											}
											rec := clip.UniformRRect(image.Rect(0, 0, dim.Size.X, dim.Size.Y), r/2)
											paint.FillShape(gtx.Ops, conf.DangerColor, rec.Op(gtx.Ops))

											call.Add(gtx.Ops)
											return dim
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
										func(gtx layout.Context) layout.Dimensions {
											return p.layoutH.Layout(
												gtx,
												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
													ic := config.ICOffline
													if connDev.Online {
														ic = config.ICOnline
													}
													return components.NewIcon(th, gtx, ic, conf.FGColor, 25)
												}),
												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
													return material.Label(th, info_size, connDev.OS+" - "+connDev.Addr.String()).Layout(gtx)
												}),
											)
										},
									)
								}),
							)
						}),
					)
				},
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

func (p *Scanner) OnFound() {
	p.win.Invalidate()
}

func (p *Scanner) OnProgress(pro float64) {
	p.progress = pro
	p.win.Invalidate()
}

func (p *Scanner) Open(ID string) {
	if p.OnOpen != nil {
		p.OnOpen(ID)
	}
}
