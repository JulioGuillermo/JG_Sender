package main

import (
	"log"
	"os"

	"gioui.org/app"
	_ "gioui.org/app/permission/storage"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/connection"
	"github.com/julioguillermo/jg_sender/font"
	"github.com/julioguillermo/jg_sender/gui/components"
	"github.com/julioguillermo/jg_sender/gui/dialog"
	"github.com/julioguillermo/jg_sender/gui/screen"
)

func main() {
	os.Setenv("LANG", "en_US.utf8")
	go func() {
		th := material.NewTheme(font.JGFonts())
		conf := config.NewConfig(th)

		w := app.NewWindow(
			app.Title("JG Sender"),
			app.Size(400, 600),
			app.StatusColor(conf.AndroidBarColor),
		)

		err := run(th, w, conf)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(th *material.Theme, w *app.Window, conf *config.Config) error {
	server := connection.InitServer(conf)

	th.TextSize = unit.Sp(20)

	config_screen := screen.NewConfigScreen(conf)
	subnet_screen := screen.NewSubnetworksScreen(conf)
	inbox_screen := screen.NewInboxScreen(conf)
	scanner_screen := screen.NewScannerScreen(conf, subnet_screen, w, inbox_screen.NewInbox)

	tabs := components.NewTab(
		0,
		components.NewTabItem("Subnetworks", config.ICSubNetworks),
		components.NewTabItem("Connections", config.ICConnections),
		components.NewTabItem("Inbox", config.ICInbox),
		components.NewTabItem("Config", config.ICConfig),
	)

	var ops op.Ops

	var (
		new screen.Screen
		old screen.Screen
		dlg dialog.Dialog
	)

	conf.SetDialogOpener(dlg.SetWidget)
	conf.SetDialogCloser(func() {
		dlg.RemoveWidget()
	})

	notify := func() {
		if tabs.ScreenIndex() != 2 {
			tabs.Notify(2, true)
		}
	}

	server.OnMSG = func(addr, name, msg string) {
		inbox_screen.NewInbox(components.NewMSG(addr, name, msg), true)
		notify()
		w.Invalidate()
	}
	server.OnFile = func(addr, user, file string, onCancel func()) (func(uint64, uint64, uint64), func(error)) {
		file_item := components.NewInboxFile(addr, user, file, w, onCancel)
		inbox_screen.NewInbox(file_item, true)
		notify()
		w.Invalidate()
		return file_item.SetProgress, file_item.SetError
	}

	bottom_size := unit.Dp(float32(tabs.Height) * 0.75)
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			layout.Stack{
				Alignment: layout.S,
			}.Layout(
				gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Bottom: bottom_size,
					}.Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(
								gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									if tabs.Changed() || new == nil {
										old = new
										switch tabs.ScreenIndex() {
										case 0:
											new = subnet_screen
										case 1:
											new = scanner_screen
										case 2:
											tabs.Notify(2, false)
											new = inbox_screen
										default:
											new = config_screen
										}
										if old != new {
											new.InAnim()
										}
									}
									if old == nil || new.Stopped() {
										return new.Layout(th, gtx, w, conf)
									}
									return layout.Stack{
										Alignment: layout.S,
									}.Layout(
										gtx,
										layout.Stacked(func(gtx layout.Context) layout.Dimensions {
											return old.Layout(th, gtx, w, conf)
										}),
										layout.Stacked(func(gtx layout.Context) layout.Dimensions {
											return new.Layout(th, gtx, w, conf)
										}),
									)
								}),
							)
						},
					)
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					return tabs.Layout(th, gtx, w, conf)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return dlg.Layout(th, gtx, w, conf)
				}),
			)

			e.Frame(gtx.Ops)
		}
	}
}
