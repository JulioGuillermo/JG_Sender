package components

import (
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
)

const (
	ColorSize   = 30
	ColorMargin = 0
)

type ColorIntensityItem struct {
	colorIntensity color.NRGBA
	clickable      widget.Clickable
}

type ColorItem struct {
	intensities []*ColorIntensityItem
	list        widget.List
}

type ColorDialog struct {
	close  widget.Clickable
	colors []*ColorItem
	list   widget.List

	OnSelect func(color.NRGBA)
}

func IntToRGB(i float32, c color.NRGBA) color.NRGBA {
	i *= 2
	if i > 1 {
		return ColorTransition(c, color.NRGBA{255, 255, 255, 255}, i-1)
	}
	return ColorTransition(color.NRGBA{0, 0, 0, 255}, c, i)
}

func NewIntensity(r, g, b uint8, i float32) *ColorIntensityItem {
	cii := &ColorIntensityItem{}

	cii.colorIntensity = color.NRGBA{
		R: r,
		G: g,
		B: b,
		A: 255,
	}
	cii.colorIntensity = IntToRGB(i, cii.colorIntensity)

	return cii
}

func NewColorItem(canIntensities int, r, g, b uint8) *ColorItem {
	ci := &ColorItem{
		intensities: make([]*ColorIntensityItem, canIntensities+1),
	}
	ci.list.Axis = layout.Horizontal

	for i := 0; i <= canIntensities; i++ {
		ci.intensities[i] = NewIntensity(r, g, b, 1-float32(i)/float32(canIntensities))
	}

	return ci
}

func NewColorDialog(canColors, canIntensities int) *ColorDialog {
	cd := &ColorDialog{
		colors: []*ColorItem{NewColorItem(canIntensities, 128, 128, 128)},
	}
	cd.list.Axis = layout.Vertical

	inc := 256.0 * 6 / float32(canColors)
	r := float32(255)
	g := float32(0)
	b := float32(0)
	// R-G
	// G up
	for g < 255 {
		cd.colors = append(cd.colors, NewColorItem(canIntensities, uint8(r), uint8(g), uint8(b)))
		g += inc
	}
	g = 255
	// R down
	for r > 0 {
		cd.colors = append(cd.colors, NewColorItem(canIntensities, uint8(r), uint8(g), uint8(b)))
		r -= inc
	}
	r = 0

	// G-B
	// B up
	for b < 255 {
		cd.colors = append(cd.colors, NewColorItem(canIntensities, uint8(r), uint8(g), uint8(b)))
		b += inc
	}
	b = 255
	// G down
	for g > 0 {
		cd.colors = append(cd.colors, NewColorItem(canIntensities, uint8(r), uint8(g), uint8(b)))
		g -= inc
	}
	g = 0

	// B-R
	// R up
	for r < 255 {
		cd.colors = append(cd.colors, NewColorItem(canIntensities, uint8(r), uint8(g), uint8(b)))
		r += inc
	}
	r = 255
	// B down
	for b > 0 {
		cd.colors = append(cd.colors, NewColorItem(canIntensities, uint8(r), uint8(g), uint8(b)))
		b -= inc
	}

	return cd
}

func (p *ColorDialog) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions {
	if gtx.Constraints.Max.X > gtx.Dp(500) {
		gtx.Constraints.Max.X = gtx.Dp(500)
	}
	if gtx.Constraints.Max.Y > gtx.Dp(600) {
		gtx.Constraints.Max.Y = gtx.Dp(600)
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
					title := material.Label(th, 20, "Select a color")
					title.Color = conf.BGPrimaryColor
					return title.Layout(gtx)
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
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &p.list).Layout(
				gtx,
				len(p.colors),
				func(gtx layout.Context, index int) layout.Dimensions {
					return p.colors[index].Layout(th, gtx, w, conf, p.OnSelect)
				},
			)
		}),
	)

	if p.close.Clicked() {
		conf.CloseDialog()
	}

	return dim
}

func (p *ColorItem) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, onSelect func(color.NRGBA)) layout.Dimensions {
	return material.List(th, &p.list).Layout(
		gtx,
		len(p.intensities),
		func(gtx layout.Context, index int) layout.Dimensions {
			return p.intensities[index].Layout(th, gtx, w, conf, onSelect)
		},
	)
}

func (p *ColorIntensityItem) Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config, onSelect func(color.NRGBA)) layout.Dimensions {
	if p.clickable.Clicked() && onSelect != nil {
		onSelect(p.colorIntensity)
	}
	return layout.UniformInset(ColorMargin).Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			return p.clickable.Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					size := image.Pt(gtx.Dp(ColorSize), gtx.Dp(ColorSize))

					rec := clip.Rect{
						Min: image.Pt(0, 0),
						Max: size,
					}
					paint.FillShape(gtx.Ops, color.NRGBA{255, 255, 255, 255}, rec.Op())
					rec.Min.X++
					rec.Min.Y++
					rec.Max.X--
					rec.Max.Y--
					paint.FillShape(gtx.Ops, color.NRGBA{0, 0, 0, 255}, rec.Op())
					rec.Min.X++
					rec.Min.Y++
					rec.Max.X--
					rec.Max.Y--
					paint.FillShape(gtx.Ops, p.colorIntensity, rec.Op())

					return layout.Dimensions{Size: size}
				},
			)
		},
	)
}
