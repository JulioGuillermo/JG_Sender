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
	"gioui.org/x/notify"
	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/connection"
	"github.com/julioguillermo/jg_sender/font"
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
	th.TextSize = unit.Sp(20)

	server := connection.InitServer(conf)
	notifications := map[string][]notify.Notification{}

	history := screen.NewHistoryScreen(th, conf, w)
	config_screen := screen.NewConfigScreen(conf)
	subnet_screen := screen.NewSubnetworksScreen(th, conf)
	scanner_screen := screen.NewScannerScreen(th, conf, subnet_screen, w)

	scanner_screen.OnOpen = history.Open
	scanner_screen.Notification = notifications

	history.Notification = notifications
	history.SendMSG = server.SendMSG
	history.SendRes = server.SendResources
	history.ContinueTrans = server.ContinueTrans

	server.UpdateHistory = history.Update
	server.Notify = func(UserID, title, txt string) {
		if !history.Visibility() || history.UserID != UserID {
			n, err := notify.Push(title, txt)
			if err == nil {
				if notifications[UserID] == nil {
					notifications[UserID] = []notify.Notification{n}
				} else {
					notifications[UserID] = append(notifications[UserID], n)
				}
			}
		}
		w.Invalidate()
	}

	tabs := screen.NewTabScreen(conf)
	tabs.Push("Subnetworks", config.ICSubNetworks, subnet_screen)
	tabs.Push("Scanner", config.ICConnections, scanner_screen)
	tabs.Push("Config", config.ICConfig, config_screen)

	var ops op.Ops

	var (
		dlg dialog.Dialog
	)
	dlg.Conf = conf

	conf.SetDialogOpener(dlg.SetWidget)
	conf.SetDialogCloser(func() {
		dlg.RemoveWidget()
	})

	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			layout.Stack{
				Alignment: layout.Center,
			}.Layout(
				gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return tabs.Layout(th, gtx, w)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return history.Layout(th, gtx)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return dlg.Layout(th, gtx, w, conf)
				}),
			)

			e.Frame(gtx.Ops)
		}
	}
}
