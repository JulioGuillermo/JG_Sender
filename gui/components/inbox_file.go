package components

import (
	"errors"
	"fmt"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
)

type InboxFile struct {
	addr string
	name string
	file string

	list widget.List

	size  uint64
	pro   uint64
	nfile uint64
	err   error

	w *app.Window

	cancel   widget.Clickable
	onCancel func()
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

func NewInboxFile(addr, name, file string, w *app.Window, onCancel func()) *InboxFile {
	inboxFile := &InboxFile{
		addr: addr,
		name: name,
		file: file,

		err:   nil,
		size:  0,
		pro:   0,
		nfile: 0,

		w: w,

		onCancel: onCancel,
	}
	inboxFile.list.Axis = layout.Horizontal
	return inboxFile
}

func (p *InboxFile) SetProgress(n, pro, size uint64) {
	p.size = size
	p.pro = pro
	p.nfile = n
	if p.w != nil {
		p.w.Invalidate()
	}
}

func IsErr(e error, ts ...error) bool {
	for _, t := range ts {
		if errors.Is(e, t) {
			return true
		}
	}
	return false
}

func (p *InboxFile) SetError(err error) {
	p.err = errors.New("Canceled")
	//p.err = err
	if p.w != nil {
		p.w.Invalidate()
	}
}

func (p *InboxFile) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if p.cancel.Clicked() && p.onCancel != nil {
		p.onCancel()
	}
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
					title := material.Label(th, 20, p.name)
					title.Font.Weight = text.Bold
					title.Color = conf.BGPrimaryColor
					return title.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.pro == p.size || p.err != nil {
						return layout.Dimensions{}
					}
					return p.cancel.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return NewIcon(th, gtx, config.ICClose, conf.DangerColor, 30)
					})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			addr := material.Label(th, 13, p.addr)
			addr.Color = conf.FGColor
			return addr.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.List(th, &p.list)
			l.Indicator.MajorMinLen = 0
			l.Indicator.MinorWidth = 0
			l.Indicator.CornerRadius = 0
			return l.Layout(
				gtx,
				1,
				func(gtx layout.Context, index int) layout.Dimensions {
					name := material.Label(th, 20, p.file[:len(p.file)-1])
					name.Color = conf.FGColor
					name.Font.Weight = text.Bold
					return name.Layout(gtx)
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(
				gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if p.err != nil {
						errLab := material.Label(th, 20, p.err.Error())
						errLab.Color = conf.DangerColor
						return errLab.Layout(gtx)
					}
					var txt string
					if p.pro == p.size {
						txt = "Completed"
					} else {
						txt = fmt.Sprintf("[%d] %.0f %%", p.nfile, float32(p.pro)/float32(p.size)*100)
					}
					lab := material.Label(th, 20, txt)
					lab.Color = conf.BGPrimaryColor
					return lab.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lab := material.Label(th, 20, fmt.Sprintf("%s / %s", formatSize(float64(p.pro)), formatSize(float64(p.size))))
					lab.Color = conf.BGPrimaryColor
					return lab.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			bar := material.ProgressBar(th, float32(p.pro)/float32(p.size))
			bar.Color = conf.BGPrimaryColor
			bar.TrackColor = conf.Shadow
			return bar.Layout(gtx)
		}),
	)
}
