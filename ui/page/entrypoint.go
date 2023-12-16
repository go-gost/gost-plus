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
	"github.com/go-gost/gost-plus/tunnel/entrypoint"
	"github.com/go-gost/gost-plus/ui/icons"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type entryPointState struct {
	editor widget.Clickable
}

type entryPointPage struct {
	router *Router

	list layout.List

	wgFavorite widget.Clickable
	wgAdd      widget.Clickable

	entryPoints map[int]*entryPointState
	favorite    atomic.Bool
}

func NewEntryPointPage(r *Router) Page {
	return &entryPointPage{
		router: r,
		list: layout.List{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		},
		entryPoints: make(map[int]*entryPointState),
	}
}

func (p *entryPointPage) Init(opts ...PageOption) {
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

	p.router.bar.Title = "EntryPoints"
	p.router.bar.NavigationIcon = icons.IconHome
}

func (p *entryPointPage) Layout(gtx C, th *material.Theme) D {
	favorite := p.favorite.Load()
	// gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return p.list.Layout(gtx, entrypoint.Count(), func(gtx C, index int) D {
		s := entrypoint.GetIndex(index)
		if s == nil {
			delete(p.entryPoints, index)
			return D{}
		}

		if p.entryPoints[index] == nil {
			p.entryPoints[index] = &entryPointState{}
		}

		if favorite && !s.IsFavorite() {
			return D{}
		}

		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
				return component.Surface(th).Layout(gtx, func(gtx C) D {
					state := p.entryPoints[index]
					if state.editor.Clicked(gtx) {
						switch s.Type() {
						case entrypoint.TCPEntryPoint:
							p.router.SwitchTo(Route{Path: PageEditTCPEntryPoint, ID: s.ID()})
						case entrypoint.UDPEntryPoint:
							p.router.SwitchTo(Route{Path: PageEditUDPEntryPoint, ID: s.ID()})
						}
						op.InvalidateOp{}.Add(gtx.Ops)
					}
					return material.Clickable(gtx, &state.editor, func(gtx C) D {
						return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
							return p.layout(gtx, th, s)
						})
					})
				})
			})
		})
	})
}

func (p *entryPointPage) layout(gtx C, th *material.Theme, ep entrypoint.EntryPoint) D {
	return layout.Flex{
		Alignment: layout.Middle,
		Spacing:   layout.SpaceBetween,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					label := material.Body1(th, ep.ID())
					label.Font.Weight = font.Bold
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Type: %s", strings.ToUpper(ep.Type()))).Layout),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Name: %s", ep.Name())).Layout),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Endpoint: %s", ep.Endpoint())).Layout),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(material.Body2(th, fmt.Sprintf("Entrypoint: %s", ep.Entrypoint())).Layout),
				layout.Rigid(layout.Spacer{Height: 5}.Layout),
				layout.Rigid(func(gtx C) D {
					if err := ep.Err(); !ep.IsClosed() && err != nil {
						label := material.Body2(th, err.Error())
						label.Color = color.NRGBA(colornames.Red500)
						return label.Layout(gtx)
					}
					return D{}
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Width: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			c := colornames.Green500
			if ep.Err() != nil {
				c = colornames.Red500
			}
			if ep.IsClosed() {
				c = colornames.Grey500
			}
			return icons.IconTunnelState.Layout(gtx, color.NRGBA(c))
		}),
	)
}
