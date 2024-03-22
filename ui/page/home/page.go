package home

import (
	"image/color"
	"sync/atomic"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/page/home/list"
	ui_widget "github.com/go-gost/gost.plus/ui/widget"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type C = layout.Context
type D = layout.Dimensions

type navPage struct {
	list list.List
	path page.PagePath
}

type homePage struct {
	router *page.Router

	btnFavorite widget.Clickable
	favorite    atomic.Bool

	nav         *ui_widget.Nav
	pages       []navPage
	btnAdd      widget.Clickable
	btnSettings widget.Clickable
}

func NewPage(r *page.Router) page.Page {
	return &homePage{
		router: r,
		pages: []navPage{
			{
				list: list.Tunnel(r),
				path: page.PageTunnel,
			},
			{
				list: list.Entrypoint(r),
				path: page.PageEntrypoint,
			},
		},
	}
}

func (p *homePage) Init(opts ...page.PageOption) {
	p.nav = ui_widget.NewNav(
		ui_widget.NewNavButton(i18n.Tunnel),
		ui_widget.NewNavButton(i18n.Entrypoint),
	)
}

func (p *homePage) Layout(gtx C) D {
	if p.btnAdd.Clicked(gtx) {
		p.router.Goto(page.Route{Path: p.pages[p.nav.Current()].path})
	}

	th := p.router.Theme

	return layout.Stack{
		Alignment: layout.SE,
	}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
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
							Spacing:   layout.SpaceBetween,
							Alignment: layout.Middle,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = gtx.Dp(50)
								return icons.IconApp.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: 8}.Layout),
							layout.Flexed(1, func(gtx C) D {
								label := material.H6(th, "GOST+")
								label.Font.Weight = font.SemiBold
								return label.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: 8}.Layout),

							layout.Rigid(func(gtx C) D {
								if p.btnFavorite.Clicked(gtx) {
									p.favorite.Store(!p.favorite.Load())
								}

								btn := material.IconButton(th, &p.btnFavorite, icons.IconFavorite, "Favorite")

								if p.favorite.Load() {
									btn.Color = color.NRGBA(colornames.Red500)
								} else {
									btn.Color = th.Fg
								}
								btn.Background = th.Bg

								return btn.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: 8}.Layout),
							layout.Rigid(func(gtx C) D {
								if p.btnSettings.Clicked(gtx) {
									p.router.Goto(page.Route{
										Path: page.PageSettings,
									})
								}

								btn := material.IconButton(th, &p.btnSettings, icons.IconSettings, "Settings")
								btn.Color = th.Fg
								btn.Background = th.Bg
								return btn.Layout(gtx)
							}),
						)
					})
				}),
				// nav
				layout.Rigid(func(gtx C) D {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Top:    4,
							Bottom: 4,
						}.Layout(gtx, func(gtx C) D {
							return p.nav.Layout(gtx, th)
						})
					})
				}),
				// list
				layout.Flexed(1, func(gtx C) D {
					current := p.nav.Current()
					if current < 0 || current >= len(p.pages) {
						current = 0
					}
					pg := p.pages[current]
					if pg.list == nil {
						return D{
							Size: gtx.Constraints.Max,
						}
					}
					pg.list.Filter(list.Filter{
						Favorite: p.favorite.Load(),
					})
					return pg.list.Layout(gtx, th)
				}),
			)
		}),
		layout.Stacked(func(gtx C) D {
			return layout.Inset{
				Top:    16,
				Bottom: 16,
				Left:   16,
				Right:  16,
			}.Layout(gtx, func(gtx C) D {
				btn := material.IconButton(th, &p.btnAdd, icons.IconAdd, "Add")
				btn.Inset = layout.UniformInset(16)

				return btn.Layout(gtx)
			})
		}),
	)
}
