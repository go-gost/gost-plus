package main

import (
	_ "net"
	"os"

	"gioui.org/app"
	_ "gioui.org/app/permission/storage"
	"gioui.org/io/key"
	"gioui.org/op"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	"github.com/go-gost/gost.plus/ui"
	_ "github.com/go-gost/gost.plus/winres"
)

func main() {
	Init()

	go func() {
		w := app.NewWindow(
			app.Title("GOST+"),
			app.MinSize(800, 600),
		)
		err := run(w)
		if err != nil {
			logger.Default().Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	ui := ui.NewUI()
	var ops op.Ops
	for {
		switch e := w.NextEvent().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			ui.Layout(gtx)
			e.Frame(gtx.Ops)
		case key.Event:
			if e.Name == key.NameBack {
				return nil
			}
		}
	}
}

func Init() {
	config.Init()
	tunnel.LoadConfig()
	entrypoint.LoadConfig()
}
