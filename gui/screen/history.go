package screen

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
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

const (
	BigMargin     = 50
	SmallMargin   = 20
	CornerRadious = 20
)

type History struct {
	UserID       string
	Notification map[string][]notify.Notification

	conf *config.Config
	win  *app.Window

	appbar *component.AppBar

	list     widget.List
	items    []*InboxItem
	entry    widget.Editor
	send     widget.Clickable
	openFile widget.Clickable
	card     *components.Card

	anim    outlay.Animation
	visible bool
	closing bool

	SendMSG       func(string, string)
	SendRes       func(string, []string)
	ContinueTrans func(string, *connection.Transfer)
}

type InboxItem struct {
	anim      outlay.Animation
	clickable widget.Clickable
}

func NewHistoryScreen(th *material.Theme, conf *config.Config, w *app.Window) *History {
	history := &History{
		conf: conf,
		win:  w,
		card: components.NewSimpleCard(conf.BGColor, 28, 10, 3),
	}

	modal := component.NewModal()
	appbar := component.NewAppBar(modal)
	appbar.Title = "History"
	appbar.NavigationLayout = func(gtx layout.Context) layout.Dimensions {
		bls := material.ButtonLayout(th, &appbar.NavigationButton)
		bls.CornerRadius = 25
		return bls.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return components.NewIcon(th, gtx, config.ICBack, conf.FGPrimaryColor, ScreenBarHeight)
		})
	}
	history.appbar = appbar

	history.list.List.Axis = layout.Vertical
	return history
}

func formatSize(size float64) string {
	const (
		b  = 1024.0
		kb = b * b
		mb = b * kb
	)
	switch {
	case size > mb:
		size /= mb
		return fmt.Sprintf("%.2f GB", size)
	case size > kb:
		size /= kb
		return fmt.Sprintf("%.2f MB", size)
	case size > b:
		size /= b
		return fmt.Sprintf("%.2f KB", size)
	default:
		return fmt.Sprintf("%.2f B", size)
	}
}

func (p *History) Update(UserID string) {
	if p.visible && p.UserID == UserID {
		p.win.Invalidate()
	}
}

func (p *History) Open(id string) {
	p.closing = false
	p.visible = true
	p.anim.Duration = p.conf.AnimTime()
	p.anim.Start(time.Now())
	p.UserID = id
	if p.Notification != nil {
		nots := p.Notification[id]
		if len(nots) > 0 {
			for _, n := range nots {
				n.Cancel()
			}
			p.Notification[id] = []notify.Notification{}
		}
	}
}

func (p *History) Close() {
	p.closing = true
	p.anim.Duration = p.conf.AnimTime()
	p.anim.Start(time.Now())
}

func (p *History) Visibility() bool {
	return p.visible
}

func (p *History) Layout(th *material.Theme, gtx layout.Context) layout.Dimensions {
	if p.appbar.NavigationButton.Clicked() {
		p.Close()
	}
	animPro := p.anim.Progress(gtx)
	if p.closing {
		animPro = 1 - animPro
		if animPro == 0 {
			p.visible = false
		}
	}
	if !p.visible {
		return layout.Dimensions{}
	}

	if p.send.Clicked() && p.SendMSG != nil && p.entry.Text() != "" {
		go p.SendMSG(p.UserID, p.entry.Text())
		p.entry.SetText("")
	} else if p.openFile.Clicked() {
		diag := components.NewFileDialog(p.UserID, p.SendRes)
		p.conf.OpenDialog(diag.Layout)
	}

	if animPro < 1 {
		gtx.Constraints.Max.Y = int(animPro * float32(gtx.Constraints.Max.Y))
		gtx.Constraints.Max.X = int(animPro * float32(gtx.Constraints.Max.X))
	}
	gtx.Constraints.Min = gtx.Constraints.Max

	rec := clip.Rect{
		Min: image.Pt(0, 0),
		Max: gtx.Constraints.Max,
	}
	paint.FillShape(gtx.Ops, p.conf.ScreenColor, rec.Op())

	History := connection.GetUserHistory(p.UserID)

	dev := connection.GetDevice(p.UserID)
	if dev != nil {
		p.appbar.Title = dev.Name
	}

	for len(p.items) < len(History) {
		p.items = append(p.items, &InboxItem{
			anim: outlay.Animation{
				Duration:  p.conf.AnimTime(),
				StartTime: time.Now(),
			},
		})
	}
	for len(p.items) > len(History) {
		p.items = p.items[:len(p.items)-1]
	}

	p.list.ScrollToEnd = !p.list.Position.BeforeEnd
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEnd,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.appbar.Layout(gtx, th, "History", "...")
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				len(History),
				func(gtx layout.Context, index int) layout.Dimensions {
					element := History[index]
					item := p.items[index]
					return p.items[index].Layout(th, gtx, p.win, p.conf, element.In, func(gtx layout.Context) layout.Dimensions {
						if element.File != nil {
							return p.renderFile(th, gtx, element, &item.clickable, func() {
								element.File.Canceled = true
							})
						}
						return p.renderMSG(th, gtx, element)
					})
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.card.Layout(gtx, p.conf, func(gtx layout.Context) layout.Dimensions {
				var minH int
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.End,
				}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						bls := material.ButtonLayout(th, &p.openFile)
						bls.CornerRadius = 25
						bls.Background = p.conf.BGColor
						dim := bls.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(10).Layout(
								gtx,
								func(gtx layout.Context) layout.Dimensions {
									return components.NewIcon(th, gtx, config.ICOpenFile, p.conf.BGPrimaryColor, 30)
								},
							)
						})
						minH = dim.Size.Y
						return dim
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.Y = gtx.Dp(200)
						e := material.Editor(th, &p.entry, "MSG")

						macro := op.Record(gtx.Ops)
						dim := e.Layout(gtx)
						call := macro.Stop()

						offset := 0
						if dim.Size.Y < minH {
							offset = (minH - dim.Size.Y) / 2
							dim.Size.Y = minH
						}

						offsetTrans := op.Offset(image.Pt(0, offset)).Push(gtx.Ops)
						call.Add(gtx.Ops)
						offsetTrans.Pop()
						return dim
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						bls := material.ButtonLayout(th, &p.send)
						bls.CornerRadius = 25
						bls.Background = p.conf.BGColor
						return bls.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(10).Layout(
								gtx,
								func(gtx layout.Context) layout.Dimensions {
									return components.NewIcon(th, gtx, config.ICSend, p.conf.BGPrimaryColor, 30)
								},
							)
						})
					}),
				)
			})
		}),
	)
}

func GetFiles(files []*connection.Element) string {
	fs := ""
	for _, f := range files {
		fs += "- " + f.Name + "\n"
	}
	if len(fs) > 0 {
		return fs[:len(fs)-1]
	}
	return fs
}

func FormatTime(t time.Time) string {
	y1, m1, d1 := t.Date()
	y2, m2, d2 := time.Now().Date()
	if y1 == y2 && m1 == m2 && d1 == d2 {
		h, m, _ := t.Clock()
		return fmt.Sprintf("%d:%d", h, m)
	}
	return fmt.Sprintf("%d/%d/%d", d1, m1, y1)
}

func (p *History) renderFile(th *material.Theme, gtx layout.Context, element *connection.Transfer, clickable *widget.Clickable, onCancel func()) layout.Dimensions {
	device := connection.GetDevice(element.UserID)
	canContinue := device != nil && p.ContinueTrans != nil && !element.In
	if clickable.Clicked() {
		if element.Error == nil && !element.File.Canceled {
			element.File.Canceled = true
		} else if canContinue {
			go p.ContinueTrans(element.UserID, element)
		}
	}
	progress := float32(element.File.TransBytes) / float32(element.File.TotalBytes)
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			name := material.Label(th, th.TextSize, GetFiles(element.File.Files))
			name.Font.Weight = text.Bold
			return name.Layout(gtx)
			/*l := material.List(th, &p.list)
			l.Indicator.MajorMinLen = 0
			l.Indicator.MinorWidth = 0
			l.Indicator.CornerRadius = 0
			return l.Layout(
				gtx,
				0,
				func(gtx layout.Context, index int) layout.Dimensions {
					name := material.Label(th, th.TextSize, GetFiles(element.File.Files))
					name.Font.Weight = text.Bold
					return name.Layout(gtx)
				},
			)*/
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = gtx.Dp(30)
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if element.File.TransBytes == element.File.TotalBytes || ((element.Error != nil || element.File.Canceled) && !canContinue) {
						return layout.Dimensions{}
					}
					return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if element.Error != nil || element.File.Canceled {
							return components.NewIcon(th, gtx, config.ICReset, p.conf.DangerColor, 30)
						}
						return components.NewIcon(th, gtx, config.ICClose, p.conf.DangerColor, 30)
					})
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if element.Error != nil {
						errLab := material.Label(th, th.TextSize, element.Error.Error())
						errLab.Color = p.conf.DangerColor
						return errLab.Layout(gtx)
					}
					if element.File.Canceled {
						errLab := material.Label(th, th.TextSize, "Canceled")
						errLab.Color = p.conf.DangerColor
						return errLab.Layout(gtx)
					}
					var txt string
					if element.File.TransBytes == element.File.TotalBytes {
						txt = "Completed"
					} else {
						txt = fmt.Sprintf("[%d / %d] %.0f %%", element.File.Index, len(element.File.Files), progress*100)
					}
					lab := material.Label(th, th.TextSize, txt)
					lab.Color = p.conf.BGPrimaryColor
					return lab.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lab := material.Label(th, th.TextSize, fmt.Sprintf("%s / %s", formatSize(float64(element.File.TransBytes)), formatSize(float64(element.File.TotalBytes))))
					lab.Color = p.conf.BGPrimaryColor
					return lab.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			bar := material.ProgressBar(th, progress)
			bar.Color = p.conf.BGPrimaryColor
			bar.TrackColor = p.conf.Shadow
			return bar.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceStart,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.Label(th, th.TextSize*0.7, FormatTime(element.DateTime))
					return title.Layout(gtx)
				}),
			)
		}),
	)
}

func (p *History) renderMSG(th *material.Theme, gtx layout.Context, element *connection.Transfer) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if element.Error == nil {
				return layout.Dimensions{}
			}
			err := material.Label(th, th.TextSize*0.7, "Error")
			err.Color = p.conf.DangerColor
			return err.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			msg := material.Label(th, th.TextSize, element.MSG)
			return msg.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceStart,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					date := material.Label(th, th.TextSize*0.7, FormatTime(element.DateTime))
					return date.Layout(gtx)
				}),
			)
		}),
	)
}

func (p *InboxItem) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, In bool, widget func(layout.Context) layout.Dimensions) layout.Dimensions {
	marg := layout.UniformInset(10)
	if In {
		marg.Right = BigMargin
		marg.Left = SmallMargin
	} else {
		marg.Left = BigMargin
		marg.Right = SmallMargin
	}

	animPro := p.anim.Progress(gtx)
	return marg.Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dim := layout.UniformInset(10).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					return widget(gtx)
				},
			)
			call := macro.Stop()

			var col color.NRGBA
			var tails_x float32
			tails_y := float32(dim.Size.Y)
			tails_w := float32(gtx.Dp(SmallMargin))
			tails_h := float32(gtx.Dp(CornerRadious))
			if In {
				col = conf.RecivedColor
				tails_x = -float32(gtx.Dp(SmallMargin))
			} else {
				col = conf.SendedColor
				tails_x = float32(dim.Size.X + gtx.Dp(SmallMargin))
				tails_w = -tails_w
			}

			var ops op.Ops
			var tails clip.Path
			tails.Begin(&ops)
			tails.MoveTo(f32.Pt(tails_x+tails_w, tails_y-tails_h))
			tails.QuadTo(f32.Pt(tails_x+tails_w, tails_y), f32.Pt(tails_x, tails_y))
			tails.QuadTo(f32.Pt(tails_x+tails_w, tails_y), f32.Pt(tails_x+tails_w*2, tails_y-tails_h/2))
			tails.Close()

			stack := clip.Outline{Path: tails.End()}.Op().Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			stack.Pop()

			rec := clip.UniformRRect(image.Rect(0, 0, dim.Size.X, dim.Size.Y), gtx.Dp(CornerRadious))
			paint.FillShape(gtx.Ops, col, rec.Op(gtx.Ops))

			call.Add(gtx.Ops)
			if animPro < 1 {
				dim.Size.Y = int(animPro * float32(dim.Size.Y))
			}
			return dim
		},
	)
}
