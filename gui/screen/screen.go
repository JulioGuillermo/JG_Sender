package screen

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
)

type Screen interface {
	Layout(th *material.Theme, gtx layout.Context, w *app.Window, conf *config.Config) layout.Dimensions
	InAnim()
	Stopped() bool
}
