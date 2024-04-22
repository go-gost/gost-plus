package main

import (
	"context"
	_ "net"
	"os"
	"time"

	"gioui.org/app"
	_ "gioui.org/app/permission/storage"
	"gioui.org/op"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/runner"
	"github.com/go-gost/gost.plus/runner/task"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	"github.com/go-gost/gost.plus/ui"
	_ "github.com/go-gost/gost.plus/winres"
)

func main() {
	Init()

	go func() {
		var w app.Window
		w.Option(app.Title("GOST+"))
		w.Option(app.MinSize(800, 600))
		err := run(&w)
		if err != nil {
			logger.Default().Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	go func() {
		for e := range runner.Event() {
			if e.TaskID == runner.TaskUpdateStats {
				w.Invalidate()
			}
		}
	}()

	ui := ui.NewUI()
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			tunnel.SaveConfig()
			entrypoint.SaveConfig()
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			ui.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func Init() {
	config.Init()
	tunnel.LoadConfig()
	entrypoint.LoadConfig()

	runner.Exec(context.Background(), task.UpdateStats(),
		runner.WithAync(true),
		runner.WithInterval(time.Second),
		runner.WithCancel(true),
	)
}
