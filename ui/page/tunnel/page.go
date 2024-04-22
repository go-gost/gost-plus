package tunnel

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/theme"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type C = layout.Context
type D = layout.Dimensions

type navButton struct {
	btn  widget.Clickable
	name string
	desc i18n.Key
	path page.PagePath
}

type tunnelPage struct {
	router *page.Router
	list   widget.List

	navs []navButton

	btnBack widget.Clickable
}

func NewPage(r *page.Router) page.Page {
	return &tunnelPage{
		router: r,
		list: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		navs: []navButton{
			{
				name: "File",
				desc: i18n.FileTunnelDesc,
				path: page.PageTunnelFile,
			},
			{
				name: "HTTP",
				desc: i18n.HTTPTunnelDesc,
				path: page.PageTunnelHTTP,
			},
			{
				name: "TCP",
				desc: i18n.TCPTunnelDesc,
				path: page.PageTunnelTCP,
			},
			{
				name: "UDP",
				desc: i18n.UDPTunnelDesc,
				path: page.PageTunnelUDP,
			},
		},
	}
}

func (p *tunnelPage) Init(opts ...page.PageOption) {}

func (p *tunnelPage) Layout(gtx C) D {
	if p.btnBack.Clicked(gtx) {
		p.router.Back()
	}

	th := p.router.Theme

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// header
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Top:    8,
				Bottom: 8,
				Left:   8,
				Right:  8,
			}.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						btn := material.IconButton(th, &p.btnBack, icons.IconBack, "Back")
						btn.Color = th.Fg
						btn.Background = th.Bg
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx C) D {
						title := material.H6(th, i18n.Tunnel.Value())
						// title.Font.Weight = font.Bold
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
				)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return p.list.Layout(gtx, len(p.navs), func(gtx C, index int) D {
				if p.navs[index].btn.Clicked(gtx) {
					p.router.Goto(page.Route{
						Path: p.navs[index].path,
					})
				}

				return layout.Inset{
					Top:    8,
					Bottom: 8,
					Left:   8,
					Right:  8,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.ButtonLayoutStyle{
						Background:   theme.Current().ListBg,
						CornerRadius: 12,
						Button:       &p.navs[index].btn,
					}.Layout(gtx, func(gtx C) D {
						return layout.UniformInset(16).Layout(gtx, func(gtx C) D {
							return layout.Flex{
								Alignment: layout.Middle,
								Spacing:   layout.SpaceBetween,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx C) D {
									return layout.Flex{
										Axis: layout.Vertical,
									}.Layout(gtx,
										layout.Rigid(material.Body1(th, p.navs[index].name).Layout),
										layout.Rigid(layout.Spacer{Height: 8}.Layout),
										layout.Rigid(material.Body2(th, p.navs[index].desc.Value()).Layout),
									)
								}),
								layout.Rigid(layout.Spacer{Width: 8}.Layout),
								layout.Rigid(func(gtx C) D {
									return icons.IconForward.Layout(gtx, color.NRGBA(colornames.Grey500))
								}),
							)
						})
					})
				})
			})
		}),
	)
}
