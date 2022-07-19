package components

import (
	"image"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/outlay"
	"github.com/julioguillermo/jg_sender/config"
)

type TextInput struct {
	Edit      widget.Editor
	Validator func(string) bool
	Hint      string
	ctl       byte
	Anim      outlay.Animation
	anim      float32
	animTL    float32
	Height    int

	changed bool
	valid   bool
}

func NewTextInput(hint string, multiline bool) *TextInput {
	input := &TextInput{
		Hint:      hint,
		Height:    50,
		Validator: nil,
		ctl:       255,
	}
	input.Edit.SingleLine = !multiline
	return input
}

func (p *TextInput) SetText(t string) {
	p.Edit.SetText(t)
}

func (p *TextInput) Text() string {
	return p.Edit.Text()
}

func (p *TextInput) Changed() bool {
	return p.changed
}

func (p *TextInput) Valid() bool {
	return p.valid
}

func (p *TextInput) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	p.changed = len(p.Edit.Events()) > 0
	if p.changed {
		if p.Validator == nil {
			p.valid = true
		} else {
			p.valid = p.Validator(p.Edit.Text())
		}
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	if p.Edit.Focused() {
		if p.ctl != 0 {
			p.Anim.Duration = conf.AnimTime()
			p.Anim.Start(time.Now())
			p.ctl = 0
		}
		p.anim = p.Anim.Progress(gtx)

		if p.animTL > 0 {
			p.animTL = 1 - p.anim
		}
	} else {
		if p.ctl != 1 {
			p.Anim.Duration = conf.AnimTime()
			p.Anim.Start(time.Now())
			p.ctl = 1
		}
		p.anim = 1 - p.Anim.Progress(gtx)

		if len(p.Edit.Text()) == 0 {
			if p.animTL < 1 {
				p.animTL = 1 - p.anim
			}
		} else {
			if p.animTL > 0 {
				p.animTL = p.anim
			}
		}
	}

	item_color := ColorTransition(conf.FGColor, conf.BGPrimaryColor, p.anim)

	space_between := unit.Dp(5)
	hint_size := th.TextSize * 3 / 4
	top_margin := gtx.Metric.SpToDp(hint_size) + space_between

	d := layout.Stack{
		Alignment: layout.NW,
	}.Layout(
		gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(
				gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top: top_margin,
					}.Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							e := material.Editor(th, &p.Edit, "")
							e.Color = item_color
							return e.Layout(gtx)
						},
					)
				}),
			)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			d := layout.Inset{
				Top: unit.Dp(float32(top_margin) * p.animTL),
			}.Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					lab := material.Label(th, (th.TextSize-hint_size)*unit.Sp(p.animTL)+hint_size, p.Hint)
					lab.Color = conf.BGPrimaryColor
					return lab.Layout(gtx)
				},
			)
			d.Size.Y = gtx.Sp(hint_size)
			return d
		}),
	)
	d.Size.Y += gtx.Dp(2)
	rec := clip.Rect{
		Min: image.Pt(0, d.Size.Y-gtx.Dp(2)),
		Max: d.Size,
	}
	if !p.valid {
		paint.FillShape(gtx.Ops, conf.DangerColor, rec.Op())
	} else {
		paint.FillShape(gtx.Ops, conf.FGColor, rec.Op())

		rec.Max.X = int(float32(gtx.Constraints.Max.X) * p.anim)
		paint.FillShape(gtx.Ops, conf.BGPrimaryColor, rec.Op())
	}
	return d
}
