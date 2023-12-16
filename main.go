package main

import (
	"log"
	_ "net"
	"os"

	"gioui.org/app"
	_ "gioui.org/app/permission/storage"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/go-gost/gost-plus/config"
	"github.com/go-gost/gost-plus/tunnel"
	"github.com/go-gost/gost-plus/tunnel/entrypoint"
	"github.com/go-gost/gost-plus/ui"
)

func main() {
	config.Init()
	tunnel.LoadConfig()
	entrypoint.LoadConfig()

	go func() {
		w := app.NewWindow(
			app.Title("GOST.PLUS"),
			app.MinSize(800, 600),
		)
		err := run(w)
		if err != nil {
			log.Fatal(err)
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
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			ui.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
