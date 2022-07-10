package components

import (
	"image"
	"image/color"
	"net/netip"
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

type ExplorerDialog struct {
	explorer *connection.Explorer
	addr     *netip.Addr
	name     string
	dim      layout.Dimensions

	close widget.Clickable

	newInboxItem func(InboxItemWidget, bool)

	err error

	storages widget.Clickable
	dirUp    widget.Clickable
	download widget.Clickable

	dirlist  widget.List
	list     widget.List
	dir      string
	elements []*storage.Element

	prog float32
}

func NewExplorerDialog(addr *netip.Addr, name string, newInboxItem func(InboxItemWidget, bool)) *ExplorerDialog {
	diag := &ExplorerDialog{
		addr:         addr,
		name:         name,
		newInboxItem: newInboxItem,
		dir:          "",
		explorer:     connection.NewExplorer(),
		elements:     []*storage.Element{},
	}
	diag.dirlist.Axis = layout.Horizontal
	diag.list.List.Axis = layout.Vertical
	diag.explorer.Explore(diag.addr, diag.dir, diag.onError, diag.onPath)
	return diag
}

func (p *ExplorerDialog) onError(err error) {
	p.err = err
}

func (p *ExplorerDialog) clear() {
	p.elements = []*storage.Element{}
}

func (p *ExplorerDialog) onPath(isDir bool, pth string, n uint64, t uint64) {
	if t == 0 {
		p.prog = 1
	} else {
		p.prog = float32(n) / float32(t)
	}
	p.elements = append(p.elements, &storage.Element{
		IsDir: isDir,
		Path:  pth,
		Name:  path.Base(pth),
		Anim:  0,
	})
}

func (p *ExplorerDialog) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if gtx.Constraints.Max.X > gtx.Dp(400) {
		gtx.Constraints.Max.X = gtx.Dp(400)
	}
	if gtx.Constraints.Max.Y > gtx.Dp(500) {
		gtx.Constraints.Max.Y = gtx.Dp(500)
	}
	rec := clip.UniformRRect(image.Rect(0, 0, p.dim.Size.X, p.dim.Size.Y), gtx.Dp(20))
	paint.FillShape(gtx.Ops, conf.BGColor, rec.Op(gtx.Ops))

	selected := false
	if p.storages.Clicked() {
		p.dir = ""
		p.clear()
		p.explorer.Explore(p.addr, p.dir, p.onError, p.onPath)
	} else if p.dirUp.Clicked() {
		p.dir = path.Dir(p.dir)
		p.clear()
		p.explorer.Explore(p.addr, p.dir, p.onError, p.onPath)
	} else {
		for _, element := range p.elements {
			if element.Clickable.Clicked() {
				if element.IsDir {
					p.dir = element.Path
					p.clear()
					p.explorer.Explore(p.addr, p.dir, p.onError, p.onPath)
				} else {
					p.explorer.Download(p.addr, []string{element.Path}, func(err error) {
						p.onError(err)
						w.Invalidate()
					}, func() {
						conf.CloseDialog()
					})
				}
			} else if element.Selected.Value {
				selected = true
			}
		}
		if selected && p.download.Clicked() {
			res := []string{}
			n := ""
			for _, e := range p.elements {
				if e.Selected.Value {
					res = append(res, e.Path)
					n += e.Name + "\n"
				}
			}
			p.explorer.Download(p.addr, res, func(err error) {
				p.onError(err)
				w.Invalidate()
			}, func() {
				conf.CloseDialog()
			})
		}
	}

	p.dim = layout.UniformInset(10).Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(
						gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							to := material.Label(th, 20, "Exploring: "+p.name)
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
					to := material.Label(th, 13, p.addr.String())
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
									lab := material.Label(th, 13, p.dir)
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
					to := material.Label(th, 13, p.err.Error())
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
							return material.Clickable(gtx, &p.download, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Axis:      layout.Horizontal,
									Alignment: layout.Middle,
								}.Layout(
									gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return NewIcon(th, gtx, config.ICDOWNLOAD, col, 40)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										lab := material.Label(th, 20, "Download")
										lab.Color = col
										return lab.Layout(gtx)
									}),
								)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.prog == 1 {
								return layout.Dimensions{}
							}
							bar := material.ProgressBar(th, p.prog)
							bar.Color = conf.BGPrimaryColor
							bar.TrackColor = conf.Shadow
							return bar.Layout(gtx)
						}),
					)
				}),
			)
		},
	)

	if p.close.Clicked() {
		conf.CloseDialog()
	}

	return p.dim
}

func (p *ExplorerDialog) render(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, index int) layout.Dimensions {
	const (
		title_size = 15
		path_size  = 10
	)
	element := p.elements[index]

	if element.Anim < 1 {
		element.Anim += AnimSpeed(gtx)
		if element.Anim > 1 {
			element.Anim = 1
		}
		op.InvalidateOp{At: gtx.Now.Add(Time(gtx))}.Add(gtx.Ops)
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
