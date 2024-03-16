package ui

import (
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/config"
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
	}

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = theme.Current().Material

	router := page.NewRouter(th)
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
		router: router,
	}
}

func (ui *UI) Layout(gtx C) D {
	return ui.router.Layout(gtx)
}
