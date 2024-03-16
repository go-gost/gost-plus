package home

import (
	"fmt"
	"image/color"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost.plus/ui/icons"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type Widget interface {
	Layout(gtx C, th *material.Theme) D
}

type serviceWidget struct {
	list layout.List
	nav  widget.Clickable
}

func (p *serviceWidget) Layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Left:  10,
				Right: 10,
			}.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Spacing:   layout.SpaceBetween,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return material.H6(th, "Services").Layout(gtx)
					}),
					layout.Flexed(1, layout.Spacer{}.Layout),
					layout.Rigid(func(gtx C) D {
						btn := material.IconButtonStyle{
							Color:       color.NRGBA(colornames.Grey800),
							Icon:        icons.IconNavArrowForward,
							Button:      &p.nav,
							Size:        24,
							Inset:       layout.UniformInset(5),
							Description: "forward",
						}
						return btn.Layout(gtx)

					}),
				)
			})
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return p.list.Layout(gtx, 10, func(gtx C, index int) D {
				return layout.Inset{
					Top:    5,
					Bottom: 5,
					Right:  10,
					Left:   10,
				}.Layout(gtx, func(gtx C) D {
					return component.SurfaceStyle{
						Theme: th,
						ShadowStyle: component.ShadowStyle{
							CornerRadius: 15,
						},
						Fill: color.NRGBA(colornames.BlueGrey50),
					}.Layout(gtx, func(gtx C) D {
						return layout.Inset{
							Top:    10,
							Bottom: 10,
							Left:   10,
							Right:  10,
						}.Layout(gtx, material.H4(th, fmt.Sprintf("Service-%d\n\n", index)).Layout)
					})
				})
			})
		}),
	)
}
