package page

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/go-gost/gost-plus/ui/icons"
	"github.com/go-gost/gost-plus/version"
)

type aboutPage struct {
	router *Router

	list layout.List
}

func NewAboutPage(r *Router) Page {
	return &aboutPage{
		router: r,
	}
}

func (p *aboutPage) Init(opts ...PageOption) {
	p.router.bar.SetActions(nil, nil)
	p.router.bar.Title = "About"
	p.router.bar.NavigationIcon = icons.IconBack
}

func (p *aboutPage) Layout(gtx C, th *material.Theme) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		return p.list.Layout(gtx, 1, func(gtx C, _ int) D {
			return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						if icons.IconApp == nil {
							return D{}
						}
						return icons.IconApp.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: 10}.Layout),
					layout.Rigid(func(gtx C) D {
						label := material.H6(th, "GOST.PLUS")
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: 10}.Layout),
					layout.Rigid(func(gtx C) D {
						return material.Body1(th, version.Version).Layout(gtx)
					}),
				)
			})
		})
	})
}
