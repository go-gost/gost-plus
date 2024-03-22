package list

import (
	"fmt"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/theme"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type entrypointState struct {
	btn widget.Clickable
}

type entrypointList struct {
	router *page.Router
	list   layout.List
	states []entrypointState
	filter Filter
}

func (l *entrypointList) Filter(f Filter) {
	l.filter = f
}

func Entrypoint(r *page.Router) List {
	return &entrypointList{
		router: r,
		list: layout.List{
			Axis: layout.Vertical,
		},
		states: make([]entrypointState, 16),
	}
}

func (l *entrypointList) Layout(gtx C, th *material.Theme) D {
	tn := entrypoint.Count()
	if tn > len(l.states) {
		states := l.states
		l.states = make([]entrypointState, tn)
		copy(l.states, states)
	}

	return l.list.Layout(gtx, tn, func(gtx C, index int) D {
		t := entrypoint.GetIndex(index)
		if t == nil {
			return D{}
		}

		if l.filter.Favorite && !t.IsFavorite() {
			return D{}
		}

		if l.states[index].btn.Clicked(gtx) {
			var path page.PagePath
			switch t.Type() {
			case entrypoint.TCPEntryPoint:
				path = page.PageEntrypointTCP
			case entrypoint.UDPEntryPoint:
				path = page.PageEntrypointUDP
			}
			l.router.Goto(page.Route{
				Path: path,
				ID:   t.ID(),
			})
		}

		return layout.Inset{
			Top:    8,
			Bottom: 8,
			Left:   8,
			Right:  8,
		}.Layout(gtx, func(gtx C) D {
			return material.ButtonLayoutStyle{
				Background:   theme.Current().ListBg,
				CornerRadius: 12,
				Button:       &l.states[index].btn,
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
								layout.Rigid(func(gtx C) D {
									label := material.Body1(th, t.ID())
									label.Font.Weight = font.SemiBold
									return label.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Height: 8}.Layout),
								layout.Rigid(material.Body2(th, fmt.Sprintf("%s: %s", i18n.Get(i18n.Type), strings.ToUpper(t.Type()))).Layout),
								layout.Rigid(layout.Spacer{Height: 8}.Layout),
								layout.Rigid(material.Body2(th, fmt.Sprintf("%s: %s", i18n.Get(i18n.Name), t.Name())).Layout),
								layout.Rigid(layout.Spacer{Height: 8}.Layout),
								layout.Rigid(material.Body2(th, fmt.Sprintf("%s: %s", i18n.Get(i18n.Endpoint), t.Endpoint())).Layout),
								layout.Rigid(layout.Spacer{Height: 8}.Layout),
								layout.Rigid(material.Body2(th, fmt.Sprintf("%s: %s", i18n.Get(i18n.Entrypoint), t.Entrypoint())).Layout),
							)
						}),
						layout.Rigid(layout.Spacer{Width: 8}.Layout),
						layout.Rigid(func(gtx C) D {
							gtx.Constraints.Min.X = gtx.Dp(12)

							c := colornames.GreenA700
							if t.Err() != nil {
								c = colornames.Red600
							}
							if t.IsClosed() {
								c = colornames.Grey600
							}
							return icons.IconCircle.Layout(gtx, color.NRGBA(c))
						}),
					)
				})
			})
		})
	})
}
