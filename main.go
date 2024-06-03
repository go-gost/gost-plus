package main

import (
	"context"
	"fmt"
	"log/slog"
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
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/theme"
	"github.com/go-gost/gost.plus/ui/widget"
	_ "github.com/go-gost/gost.plus/winres"
)

func main() {
	Init()

	go func() {
		if err := run(); err != nil {
			logger.Default().Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run() error {
	ui := ui.NewUI()

	go handleEvent(ui)

	w := ui.Window()
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

func handleEvent(ui *ui.UI) {
	for {
		select {
		case e := <-ui.Router().Event():
			switch e.ID {
			case page.EventThemeChanged:
				slog.Debug("theme changed", "event", e.ID)
				ui.Window().Option(app.StatusColor(theme.Current().Material.Bg))
			}

		case e := <-runner.Event():
			switch e.TaskID {
			case runner.TaskUpdateStats:
				ui.Window().Invalidate()

			default:
				if e.Err != nil {
					slog.Error(fmt.Sprintf("task: %s", e.Err), "task", e.TaskID)
					ui.Router().Notify(widget.Message{
						Type:    widget.Error,
						Content: e.Err.Error(),
					})
				}
			}
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
