package components

import (
	"image"
	"image/color"
	"path"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/config/storage"
	"github.com/julioguillermo/jg_sender/connection"
)

type FileDialog struct {
	userid  string
	SendRes func(string, []string)

	close widget.Clickable

	err error

	storages widget.Clickable
	dirUp    widget.Clickable
	send     widget.Clickable

	dirlist  widget.List
	list     widget.List
	dir      string
	elements []*storage.Element
}

func NewFileDialog(userID string, sendRes func(userID string, resources []string)) *FileDialog {
	diag := &FileDialog{
		userid:  userID,
		SendRes: sendRes,
		dir:     "",
	}
	diag.dirlist.Axis = layout.Horizontal
	diag.list.List.Axis = layout.Vertical
	diag.elements, diag.err = storage.Explore(diag.dir)
	return diag
}

func (p *FileDialog) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	device := connection.GetDevice(p.userid)
	addr := ""
	name := "Unknown"
	if device != nil {
		name = device.Name
		addr = device.Addr.String()
	}

	if gtx.Constraints.Max.X > gtx.Dp(400) {
		gtx.Constraints.Max.X = gtx.Dp(400)
	}
	if gtx.Constraints.Max.Y > gtx.Dp(500) {
		gtx.Constraints.Max.Y = gtx.Dp(500)
	}

	selected := false
	if p.storages.Clicked() {
		p.dir = ""
		p.elements, p.err = storage.Explore(p.dir)
	} else if p.dirUp.Clicked() {
		p.dir = path.Dir(p.dir)
		p.elements, p.err = storage.Explore(p.dir)
	} else {
		for _, element := range p.elements {
			if element.Clickable.Clicked() {
				if element.IsDir {
					p.dir = element.Path
					p.elements, p.err = storage.Explore(p.dir)
				} else {
					if p.SendRes != nil {
						go p.SendRes(p.userid, []string{element.Path})
					}
					conf.CloseDialog()
					w.Invalidate()
				}
			} else if element.Selected.Value {
				selected = true
			}
		}
		if selected && p.send.Clicked() {
			res := []string{}
			n := ""
			for _, e := range p.elements {
				if e.Selected.Value {
					res = append(res, e.Path)
					n += e.Name + "\n"
				}
			}
			if p.SendRes != nil {
				go p.SendRes(p.userid, res)
			}
			conf.CloseDialog()
			w.Invalidate()
		}
	}

	dim := layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(
				gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					to := material.Label(th, th.TextSize, "File to: "+name)
					to.Color = conf.BGPrimaryColor
					return to.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.close.Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							return NewIcon(th, gtx, config.ICClose, conf.DangerColor, 30)
						},
					)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			to := material.Label(th, th.TextSize*0.7, addr)
			to.Color = conf.FGColor
			return to.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &p.storages, func(gtx layout.Context) layout.Dimensions {
						return NewIcon(th, gtx, config.ICStorages, conf.FGColor, 30)
					})
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					list := material.List(th, &p.dirlist)
					list.Indicator.MinorWidth = 0
					list.Indicator.CornerRadius = 0
					list.Indicator.MajorMinLen = 0
					return list.Layout(
						gtx,
						1,
						func(gtx layout.Context, index int) layout.Dimensions {
							lab := material.Label(th, th.TextSize*0.7, p.dir)
							return lab.Layout(gtx)
						},
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &p.dirUp, func(gtx layout.Context) layout.Dimensions {
						return NewIcon(th, gtx, config.ICDirUp, conf.FGColor, 30)
					})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.err == nil {
				return layout.Dimensions{
					Size: image.Pt(0, 0),
				}
			}
			to := material.Label(th, th.TextSize*0.7, p.err.Error())
			to.Color = conf.DangerColor
			return to.Layout(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				len(p.elements),
				func(gtx layout.Context, index int) layout.Dimensions {
					return p.render(th, gtx, w, conf, index)
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(
				gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{
						Size: image.Pt(gtx.Constraints.Max.X, 0),
					}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					var col color.NRGBA
					if selected {
						col = conf.BGPrimaryColor
					} else {
						col = conf.Shadow
					}
					return material.Clickable(gtx, &p.send, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(
							gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return NewIcon(th, gtx, config.ICSend, col, 40)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lab := material.Label(th, th.TextSize, "Send")
								lab.Color = col
								return lab.Layout(gtx)
							}),
						)
					})
				}),
			)
		}),
	)

	if p.close.Clicked() {
		conf.CloseDialog()
	}

	return dim
}

func (p *FileDialog) render(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, index int) layout.Dimensions {
	const (
		title_size = 15
		path_size  = 10
	)
	element := p.elements[index]

	if element.Anim < 1 {
		element.Anim += conf.AnimSpeed(gtx)
		if element.Anim > 1 {
			element.Anim = 1
		}
		op.InvalidateOp{At: gtx.Now.Add(conf.Time(gtx))}.Add(gtx.Ops)
	}

	d := layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(th, &element.Selected, "").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, &element.Clickable, func(gtx layout.Context) layout.Dimensions {
				card := clip.Rect{
					Min: image.Pt(0, element.Dim.Size.Y-gtx.Dp(2)),
					Max: image.Pt(element.Dim.Size.X, element.Dim.Size.Y),
				}
				paint.FillShape(gtx.Ops, conf.Shadow, card.Op())

				element.Dim = layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						var ic rune
						if p.dir == "" {
							ic = config.ICStorage
						} else if element.IsDir {
							ic = config.ICDir
						} else {
							ic = config.ICFile
						}
						return NewIcon(th, gtx, ic, conf.BGPrimaryColor, title_size*3)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								name := material.Label(th, title_size, element.Name)
								name.Color = conf.BGPrimaryColor
								name.Font.Weight = text.Bold
								return name.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lab := material.Label(th, path_size, element.Path)
								return lab.Layout(gtx)
							}),
						)
					}),
				)
				return element.Dim
			})
		}),
	)

	return d
}
