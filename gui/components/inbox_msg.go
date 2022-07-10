package components

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
)

type InboxMSG struct {
	addr string
	name string
	msg  string
}

func NewMSG(addr, name, msg string) *InboxMSG {
	return &InboxMSG{
		addr: addr,
		name: name,
		msg:  msg,
	}
}

func (p *InboxMSG) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := material.Label(th, 20, p.name)
			title.Font.Weight = text.Bold
			title.Color = conf.BGPrimaryColor
			return title.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			addr := material.Label(th, 13, p.addr)
			addr.Color = conf.FGColor
			return addr.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			msg := material.Label(th, 20, p.msg)
			msg.Color = conf.FGColor
			return msg.Layout(gtx)
		}),
	)
}
