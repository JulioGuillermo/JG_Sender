package screen

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
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

type ConfigUI struct {
	Conf *config.Config

	Name        *components.TextInput
	Inbox       *components.TextInput
	Connections *components.TextInput
	Timeout     *components.TextInput
	BufSize     *components.TextInput
	AnimTime    *components.TextInput

	reset     widget.Clickable
	openInbox widget.Clickable
	list      widget.List

	appbar *component.AppBar
	card   *components.Card

	anim outlay.Animation

	// colors clickables
	bg         widget.Clickable
	fg         widget.Clickable
	primarybg  widget.Clickable
	primaryfg  widget.Clickable
	screen     widget.Clickable
	secondary  widget.Clickable
	important  widget.Clickable
	androidbar widget.Clickable
	recived    widget.Clickable
	sended     widget.Clickable
}

func CheckNum(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func NewConfigScreen(c *config.Config) *ConfigUI {
	conf := &ConfigUI{
		Conf: c,

		Name:        components.NewTextInput("Name", false),
		Inbox:       components.NewTextInput("Inbox", false),
		Connections: components.NewTextInput("Connections", false),
		Timeout:     components.NewTextInput("Timeout (ms)", false),
		BufSize:     components.NewTextInput("Buffer size", false),
		AnimTime:    components.NewTextInput("Animation time (ms)", false),

		card: components.NewSimpleCard(c.BGColor, 20, 10, 10),
	}

	modal := component.NewModal()
	appbar := component.NewAppBar(modal)
	appbar.Title = "Config"
	conf.appbar = appbar

	conf.Name.Validator = func(s string) bool {
		return s != ""
	}
	conf.Inbox.Validator = func(s string) bool {
		if s != "" {
			inf, err := os.Stat(s)
			return err == nil && inf.IsDir()
		}
		return false
	}
	conf.Connections.Validator = func(s string) bool {
		if !CheckNum(s) {
			return false
		}
		conns, err := strconv.ParseUint(s, 10, 64)
		return err == nil && conns > 1
	}
	conf.Timeout.Validator = func(s string) bool {
		if !CheckNum(s) {
			return false
		}
		timeout, err := strconv.ParseUint(s, 10, 64)
		return err == nil && timeout > 1
	}
	conf.BufSize.Validator = func(s string) bool {
		if !CheckNum(s) {
			return false
		}
		bsize, err := strconv.ParseUint(s, 10, 64)
		return err == nil && bsize > 1
	}
	conf.AnimTime.Validator = func(s string) bool {
		if !CheckNum(s) {
			return false
		}
		_, err := strconv.ParseUint(s, 10, 64)
		return err == nil
	}
	conf.list.List.Axis = layout.Vertical

	conf.Load()

	return conf
}

func (p *ConfigUI) Load() {
	p.Name.SetText(p.Conf.Name())
	p.Inbox.SetText(p.Conf.Inbox())
	p.Connections.SetText(fmt.Sprint(p.Conf.Connections()))
	p.Timeout.SetText(fmt.Sprint(p.Conf.Timeout()))
	p.BufSize.SetText(fmt.Sprint(p.Conf.BufSize()))
	p.AnimTime.SetText(fmt.Sprint(p.Conf.C_AnimTime))
}

func (p *ConfigUI) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if p.reset.Clicked() {
		p.Conf.Reset()
		p.Conf.Save()
		p.Load()
	} else if p.openInbox.Clicked() {
		dirDiag := components.NewDirDialog(conf.Inbox(), func(s string) {
			conf.SetInbox(s)
			p.Inbox.SetText(s)
		})
		conf.OpenDialog(dirDiag.Layout)
	} else if p.Name.Changed() && p.Name.Valid() {
		p.Conf.SetName(p.Name.Text())
	} else if p.Inbox.Changed() && p.Inbox.Valid() {
		p.Conf.SetInbox(p.Inbox.Text())
	} else if p.Connections.Changed() {
		if CheckNum(p.Connections.Text()) {
			conns, err := strconv.ParseUint(p.Connections.Text(), 10, 64)
			if err == nil && conns > 1 {
				p.Conf.SetConnections(conns)
			}
		}
	} else if p.Timeout.Changed() {
		if CheckNum(p.Timeout.Text()) {
			timeout, err := strconv.ParseUint(p.Timeout.Text(), 10, 64)
			if err == nil && timeout > 1 {
				p.Conf.SetTimeout(timeout)
			}
		}
	} else if p.BufSize.Changed() {
		if CheckNum(p.BufSize.Text()) {
			bufsize, err := strconv.ParseUint(p.BufSize.Text(), 10, 64)
			if err == nil && bufsize > 1 {
				p.Conf.SetBufSize(bufsize)
			}
		}
	} else if p.AnimTime.Changed() {
		if CheckNum(p.AnimTime.Text()) {
			atime, err := strconv.ParseUint(p.AnimTime.Text(), 10, 64)
			if err == nil {
				p.Conf.SetAnimTime(atime)
			}
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
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEnd,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.appbar.Layout(gtx, th, "Config", "...")
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				1,
				func(gtx layout.Context, index int) layout.Dimensions {
					return p.card.Layout(
						gtx,
						conf,
						func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(
								gtx,
								// General Config
								p.GetConfigItem(th, w, conf, p.Name.Layout),
								p.GetConfigItem(th, w, conf, func(t *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
									return layout.Flex{
										Axis:      layout.Horizontal,
										Alignment: layout.End,
									}.Layout(
										gtx,
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											return p.Inbox.Layout(t, gtx, w, conf)
										}),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return material.Clickable(gtx, &p.openInbox, func(gtx layout.Context) layout.Dimensions {
												return layout.Flex{
													Axis:      layout.Horizontal,
													Alignment: layout.Middle,
												}.Layout(
													gtx,
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														return components.NewIcon(th, gtx, config.ICOpenDir, conf.FGColor, 40)
													}),
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														return material.Label(th, th.TextSize, "Select").Layout(gtx)
													}),
												)
											})
										}),
									)
								}),
								p.GetConfigItem(th, w, conf, p.Connections.Layout),
								p.GetConfigItem(th, w, conf, p.Timeout.Layout),
								p.GetConfigItem(th, w, conf, p.BufSize.Layout),
								p.GetConfigItem(th, w, conf, p.AnimTime.Layout),

								// Theme config
								// Main colors
								p.RenderColor(th, w, conf, conf.BGColor, "Background color", &p.bg, func(n color.NRGBA) {
									conf.BGColor = n
									conf.Save()
									th.Bg = n
									w.Invalidate()
								}),
								p.RenderColor(th, w, conf, conf.FGColor, "Text color", &p.fg, func(n color.NRGBA) {
									conf.FGColor = n
									conf.Save()
									th.Fg = n
									w.Invalidate()
								}),
								p.RenderColor(th, w, conf, conf.BGPrimaryColor, "Primary color", &p.primarybg, func(n color.NRGBA) {
									conf.BGPrimaryColor = n
									conf.Save()
									th.ContrastBg = n
									w.Invalidate()
								}),
								p.RenderColor(th, w, conf, conf.FGPrimaryColor, "Primary contrast color", &p.primaryfg, func(n color.NRGBA) {
									conf.FGPrimaryColor = n
									conf.Save()
									th.ContrastFg = n
									w.Invalidate()
								}),
								// Screen colors
								p.RenderColor(th, w, conf, conf.Shadow, "Secondary color", &p.secondary, func(n color.NRGBA) {
									n.A = 100
									conf.Shadow = n
									conf.Save()
									w.Invalidate()
								}),
								p.RenderColor(th, w, conf, conf.ScreenColor, "Screen color", &p.screen, func(n color.NRGBA) {
									conf.ScreenColor = n
									conf.Save()
									w.Invalidate()
								}),
								// Others
								p.RenderColor(th, w, conf, conf.DangerColor, "Important color", &p.important, func(n color.NRGBA) {
									conf.DangerColor = n
									conf.Save()
									w.Invalidate()
								}),
								p.RenderColor(th, w, conf, conf.SendedColor, "Sended color", &p.sended, func(n color.NRGBA) {
									conf.SendedColor = n
									conf.Save()
									w.Invalidate()
								}),
								p.RenderColor(th, w, conf, conf.RecivedColor, "Recived color", &p.recived, func(n color.NRGBA) {
									conf.RecivedColor = n
									conf.Save()
									w.Invalidate()
								}),

								// Platform settings
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if conf.OS() == "android" {
										return layout.Inset{
											Top: 10,
										}.Layout(
											gtx,
											func(gtx layout.Context) layout.Dimensions {
												return layout.Flex{
													Axis: layout.Vertical,
												}.Layout(
													gtx,
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														lab := material.Label(th, th.TextSize*1.3, "Android settings")
														lab.Color = conf.BGPrimaryColor
														return lab.Layout(gtx)
													}),
													p.RenderColor(th, w, conf, conf.AndroidBarColor, "Android bar color", &p.androidbar, func(n color.NRGBA) {
														conf.AndroidBarColor = n
														conf.Save()
														w.Option(app.StatusColor(n))
														w.Invalidate()
													}),

													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														lab := material.Label(th, th.TextSize, "Android storage permission")
														lab.Color = conf.DangerColor
														return lab.Layout(gtx)
													}),
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														lab := material.Label(th, th.TextSize, "Please allow the app to read and write internal and external storages.")
														return lab.Layout(gtx)
													}),
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														lab := material.Label(th, th.TextSize*0.7, "Settings >> Aplications >> JG_Sender >> Permission")
														return lab.Layout(gtx)
													}),
												)
											},
										)
									}
									return layout.Dimensions{}
								}),

								// Reset Config
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Inset{
										Top: 10,
									}.Layout(
										gtx,
										func(gtx layout.Context) layout.Dimensions {
											return layout.Stack{
												Alignment: layout.E,
											}.Layout(
												gtx,
												layout.Stacked(func(gtx layout.Context) layout.Dimensions {
													return material.Clickable(gtx, &p.reset, func(gtx layout.Context) layout.Dimensions {
														return layout.UniformInset(5).Layout(
															gtx,
															func(gtx layout.Context) layout.Dimensions {
																return layout.Flex{
																	Axis:      layout.Horizontal,
																	Alignment: layout.Middle,
																}.Layout(
																	gtx,
																	layout.Rigid(func(gtx layout.Context) layout.Dimensions {
																		return components.NewIcon(th, gtx, config.ICReset, conf.DangerColor, gtx.Metric.SpToDp(20))
																	}),
																	layout.Rigid(func(gtx layout.Context) layout.Dimensions {
																		lab := material.Label(th, th.TextSize, "Reset settings")
																		lab.Color = conf.DangerColor
																		return lab.Layout(gtx)
																	}),
																)
															},
														)
													})
												}),
											)
										},
									)
								}),
							)
						},
					)
				},
			)
		}),
	)
}

func (p *ConfigUI) GetConfigItem(th *material.Theme, w *app.Window, conf *config.Config, widget func(*material.Theme, layout.Context, *app.Window, *config.Config) layout.Dimensions) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    10,
			Bottom: 10,
		}.Layout(
			gtx,
			func(gtx layout.Context) layout.Dimensions {
				return widget(th, gtx, w, conf)
			},
		)
	})
}

func (p *ConfigUI) RenderColor(th *material.Theme, w *app.Window, conf *config.Config, c color.NRGBA, name string, click *widget.Clickable, onSelect func(color.NRGBA)) layout.FlexChild {
	if click.Clicked() {
		cd := components.NewColorDialog(30, 15)
		cd.OnSelect = func(n color.NRGBA) {
			conf.CloseDialog()
			if onSelect != nil {
				onSelect(n)
			}
			w.Invalidate()
		}
		conf.OpenDialog(cd.Layout)
	}
	return p.GetConfigItem(th, w, conf, func(t *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
		d := layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.End,
		}.Layout(
			gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return material.Label(th, th.TextSize, name).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return components.NewColorBox(gtx, click, c)
			}),
		)
		d.Size.Y += gtx.Dp(2)
		rec := clip.Rect{
			Min: image.Pt(0, d.Size.Y-gtx.Dp(2)),
			Max: d.Size,
		}
		paint.FillShape(gtx.Ops, conf.FGColor, rec.Op())
		return d
	})
}

func (p *ConfigUI) InAnim() {
	p.anim.Duration = p.Conf.AnimTime()
	p.anim.Start(time.Now())
}

func (p *ConfigUI) Stopped(gtx layout.Context) bool {
	return !p.anim.Animating(gtx)
}
