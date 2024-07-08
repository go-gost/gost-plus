package ui

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	gio_theme "gioui.org/x/pref/theme"
	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/ui/fonts"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/page/entrypoint"
	tcp_ep "github.com/go-gost/gost.plus/ui/page/entrypoint/tcp"
	udp_ep "github.com/go-gost/gost.plus/ui/page/entrypoint/udp"
	"github.com/go-gost/gost.plus/ui/page/home"
	"github.com/go-gost/gost.plus/ui/page/settings"
	"github.com/go-gost/gost.plus/ui/page/tunnel"
	"github.com/go-gost/gost.plus/ui/page/tunnel/file"
	"github.com/go-gost/gost.plus/ui/page/tunnel/http"
	"github.com/go-gost/gost.plus/ui/page/tunnel/tcp"
	"github.com/go-gost/gost.plus/ui/page/tunnel/udp"
	"github.com/go-gost/gost.plus/ui/theme"
)

type C = layout.Context
type D = layout.Dimensions

type UI struct {
	w      *app.Window
	router *page.Router
}

func NewUI() *UI {
	if settings := config.Get().Settings; settings != nil {
		switch settings.Theme {
		case theme.Dark:
			theme.UseDark()
		default:
			theme.UseLight()
		}
		i18n.Set(settings.Lang)
	}

	th := material.NewTheme()
	// th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Shaper = text.NewShaper(text.WithCollection(fonts.Collection()))
	th.Palette = theme.Current().Material

	w := &app.Window{}
	w.Option(
		app.Title("GOST"),
		app.MinSize(800, 600),
		app.StatusColor(th.Bg),
	)

	router := page.NewRouter(w, th)
	router.Register(page.PageHome, home.NewPage(router))
	router.Register(page.PageTunnel, tunnel.NewPage(router))
	router.Register(page.PageTunnelFile, file.NewPage(router))
	router.Register(page.PageTunnelHTTP, http.NewPage(router))
	router.Register(page.PageTunnelTCP, tcp.NewPage(router))
	router.Register(page.PageTunnelUDP, udp.NewPage(router))
	router.Register(page.PageEntrypoint, entrypoint.NewPage(router))
	router.Register(page.PageEntrypointTCP, tcp_ep.NewPage(router))
	router.Register(page.PageEntrypointUDP, udp_ep.NewPage(router))
	router.Register(page.PageSettings, settings.NewPage(router))

	router.Goto(page.Route{
		Path: page.PageHome,
	})

	return &UI{
		w:      w,
		router: router,
	}
}

func (ui *UI) Layout(gtx C) D {
	if settings := config.Get().Settings; settings != nil {
		if settings.Theme != theme.Dark && settings.Theme != theme.Light {
			if dark, _ := gio_theme.IsDarkMode(); dark {
				if theme.Current().Name == theme.Light {
					theme.UseDark()
				}
			} else {
				if theme.Current().Name == theme.Dark {
					theme.UseLight()
				}
			}
		}
	}
	return ui.router.Layout(gtx)
}

func (ui *UI) Window() *app.Window {
	return ui.w
}

func (ui *UI) Router() *page.Router {
	return ui.router
}
