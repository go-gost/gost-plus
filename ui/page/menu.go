package page

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost-plus/ui/icons"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type menuPage struct {
	router *Router

	list widget.List

	wgFile widget.Clickable
	wgHTTP widget.Clickable
	wgTCP  widget.Clickable
	wgUDP  widget.Clickable
}

func NewMenuPage(r *Router) Page {
	return &menuPage{
		router: r,
		list: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
	}
}

func (p *menuPage) Init(opts ...PageOption) {
	p.router.bar.SetActions(nil, nil)
	p.router.bar.Title = "Add"
	p.router.bar.NavigationIcon = icons.IconBack
}

func (p *menuPage) Layout(gtx C, th *material.Theme) D {
	if clicked := func() bool {
		if p.wgFile.Clicked(gtx) {
			p.router.SwitchTo(Route{Path: PageNewFile})
			return true
		}
		if p.wgHTTP.Clicked(gtx) {
			p.router.SwitchTo(Route{Path: PageNewHTTP})
			return true
		}
		if p.wgTCP.Clicked(gtx) {
			p.router.SwitchTo(Route{Path: PageNewTCP})
			return true
		}
		if p.wgUDP.Clicked(gtx) {
			p.router.SwitchTo(Route{Path: PageNewUDP})
			return true
		}

		return false
	}(); clicked {
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	return p.list.List.Layout(gtx, 1, func(gtx C, _ int) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						label := material.H6(th, "Services")
						label.Font.Weight = font.Bold
						return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, label.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, func(gtx C) D {
							return component.Surface(th).Layout(gtx, func(gtx C) D {
								return p.wgFile.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
										return p.layoutCard(gtx, th, "File", "Expose local files to public network")
									})
								})
							})
						})
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, func(gtx C) D {
							return component.Surface(th).Layout(gtx, func(gtx C) D {
								return p.wgHTTP.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
										return p.layoutCard(gtx, th, "HTTP", "Expose local HTTP service to public network")
									})
								})
							})
						})
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, func(gtx C) D {
							return component.Surface(th).Layout(gtx, func(gtx C) D {
								return p.wgTCP.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
										return p.layoutCard(gtx, th, "TCP", "Expose local TCP service to public network")
									})
								})
							})
						})
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, func(gtx C) D {
							return component.Surface(th).Layout(gtx, func(gtx C) D {
								return p.wgUDP.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
										return p.layoutCard(gtx, th, "UDP", "Expose local UDP service to public network")
									})
								})
							})
						})
					}),
				)
			})
		})
	})
}

func (p *menuPage) layoutCard(gtx C, th *material.Theme, name, desc string) D {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Spacing:   layout.SpaceBetween,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					title := material.Body1(th, name)
					title.Font.Weight = font.Bold
					return title.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: 10}.Layout),
				layout.Rigid(func(gtx C) D {
					title := material.Body1(th, desc)
					return title.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Width: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return icons.IconForward.Layout(gtx, color.NRGBA(colornames.Grey500))
		}),
	)
}
