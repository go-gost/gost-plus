package page

import (
	"fmt"
	"image/color"
	"strings"
	"sync/atomic"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost-plus/tunnel"
	"github.com/go-gost/gost-plus/ui/icons"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type tunnelState struct {
	editor widget.Clickable
}

type tunnelPage struct {
	router *Router

	list layout.List

	wgFavorite widget.Clickable
	wgAdd      widget.Clickable

	tunnels  map[int]*tunnelState
	favorite atomic.Bool
}

func NewTunnelPage(r *Router) Page {
	return &tunnelPage{
		router: r,
		list: layout.List{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		},
		tunnels: make(map[int]*tunnelState),
	}
}

func (p *tunnelPage) Init(opts ...PageOption) {
	p.router.bar.SetActions(
		[]component.AppBarAction{
			{
				OverflowAction: component.OverflowAction{
					Name: "Favorite",
					Tag:  &p.wgFavorite,
				},
				Layout: func(gtx C, bg, fg color.NRGBA) D {
					if p.wgFavorite.Clicked(gtx) {
						p.favorite.Store(!p.favorite.Load())
					}

					btn := component.SimpleIconButton(bg, fg, &p.wgFavorite, icons.IconFavorite)
					if p.favorite.Load() {
						btn.Color = color.NRGBA(colornames.Red500)
					} else {
						btn.Color = fg
					}
					return btn.Layout(gtx)
				},
			},
			{
				OverflowAction: component.OverflowAction{
					Name: "Add",
					Tag:  &p.wgAdd,
				},
				Layout: func(gtx C, bg, fg color.NRGBA) D {
					if p.wgAdd.Clicked(gtx) {
						p.router.SwitchTo(Route{Path: PageMenu})
					}
					return component.SimpleIconButton(bg, fg, &p.wgAdd, icons.IconAdd).Layout(gtx)
				},
			},
		},
		[]component.OverflowAction{
			{
				Name: "About",
				Tag:  OverflowActionAbout,
			},
		},
	)

	p.router.bar.Title = "Tunnels"
	p.router.bar.NavigationIcon = icons.IconHome
}

func (p *tunnelPage) Layout(gtx C, th *material.Theme) D {
	favorite := p.favorite.Load()
	// gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return p.list.Layout(gtx, tunnel.TunnelCount(), func(gtx C, index int) D {
		s := tunnel.GetTunnel(index)
		if s == nil {
			delete(p.tunnels, index)
			return D{}
		}

		if p.tunnels[index] == nil {
			p.tunnels[index] = &tunnelState{}
		}

		if favorite && !s.IsFavorite() {
			return D{}
		}

		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
				surface := component.Surface(th)
				return surface.Layout(gtx, func(gtx C) D {
					state := p.tunnels[index]
					if state.editor.Clicked(gtx) {
						switch s.Type() {
						case tunnel.FileTunnel:
							p.router.SwitchTo(Route{Path: PageEditFile, ID: s.ID()})
						case tunnel.HTTPTunnel:
							p.router.SwitchTo(Route{Path: PageEditHTTP, ID: s.ID()})
						case tunnel.TCPTunnel:
							p.router.SwitchTo(Route{Path: PageEditTCP, ID: s.ID()})
						case tunnel.UDPTunnel:
							p.router.SwitchTo(Route{Path: PageEditUDP, ID: s.ID()})
						}
						op.InvalidateOp{}.Add(gtx.Ops)
					}
					return state.editor.Layout(gtx, func(gtx C) D {
						return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
							return p.layoutTunnel(gtx, th, s)
						})
					})
				})
			})
		})
	})
}

func (p *tunnelPage) layoutTunnel(gtx C, th *material.Theme, s tunnel.Tunnel) D {
	return layout.Flex{
		Alignment: layout.Middle,
		Spacing:   layout.SpaceBetween,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					label := material.Body1(th, s.ID())
					label.Font.Weight = font.Bold
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Type: %s", strings.ToUpper(s.Type()))).Layout),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Name: %s", s.Name())).Layout),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Endpoint: %s", s.Endpoint())).Layout),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Entrypoint: %s", s.Entrypoint())).Layout),
			)
		}),
		layout.Rigid(layout.Spacer{Width: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			c := colornames.Green500
			if s.IsClosed() {
				c = colornames.Grey500
			}
			return icons.IconTunnelState.Layout(gtx, color.NRGBA(c))
		}),
	)
}
