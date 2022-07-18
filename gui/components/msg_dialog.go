package components

import (
	"image"
	"net/netip"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/connection"
)

type MSGDialog struct {
	addr  *netip.Addr
	name  string
	entry widget.Editor

	send  widget.Clickable
	close widget.Clickable

	err error

	newInboxItem func(InboxItemWidget, bool)
}

func NewMSGDialog(addr *netip.Addr, name string, newInboxItem func(InboxItemWidget, bool)) *MSGDialog {
	diag := &MSGDialog{
		addr:         addr,
		name:         name,
		newInboxItem: newInboxItem,
	}
	diag.entry.Focus()
	return diag
}

func (p *MSGDialog) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if gtx.Constraints.Max.X > gtx.Dp(300) {
		gtx.Constraints.Max.X = gtx.Dp(300)
	}
	if gtx.Constraints.Max.Y > gtx.Dp(200) {
		gtx.Constraints.Max.Y = gtx.Dp(200)
	}

	dim := layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(
		gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					to := material.Label(th, 20, "MSG to: "+p.name)
					to.Color = conf.BGPrimaryColor
					return to.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					to := material.Label(th, 13, p.addr.String())
					to.Color = conf.FGColor
					return to.Layout(gtx)
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
					gtx.Constraints.Min = gtx.Constraints.Max
					d := material.Editor(th, &p.entry, "").Layout(gtx)
					rec := clip.Rect{
						Min: image.Pt(0, d.Size.Y),
						Max: image.Pt(d.Size.X, d.Size.Y+gtx.Dp(1)),
					}
					paint.FillShape(gtx.Ops, conf.BGPrimaryColor, rec.Op())
					return d
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceBetween,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.close.Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							return NewIcon(th, gtx, config.ICClose, conf.DangerColor, 30)
						},
					)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{
						Size: image.Pt(0, gtx.Constraints.Max.Y),
					}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.send.Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							return NewIcon(th, gtx, config.ICSend, conf.BGPrimaryColor, 30)
						},
					)
				}),
			)
		}),
	)

	if p.send.Clicked() {
		connection.SendMSG(conf, p.addr, p.entry.Text(), func(err error) {
			p.err = err
		})
		if p.err == nil {
			if p.newInboxItem != nil {
				p.newInboxItem(NewMSG(p.addr.String(), p.name, p.entry.Text()), false)
				w.Invalidate()
			}
			conf.CloseDialog()
		}
	}
	if p.close.Clicked() {
		conf.CloseDialog()
	}

	return dim
}
