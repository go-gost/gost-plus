package widget

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
)

type Menu struct {
	List     layout.List
	Items    []MenuItem
	Title    i18n.Key
	Selected func(index int)
	btnAdd   widget.Clickable
	ShowAdd  bool
	Multiple bool
}

func (p *Menu) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    16,
			Bottom: 16,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return component.SurfaceStyle{
				Theme: th,
				ShadowStyle: component.ShadowStyle{
					CornerRadius: 28,
				},
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    16,
					Bottom: 16,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    8,
								Bottom: 8,
								Left:   24,
								Right:  24,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Alignment: layout.Middle,
								}.Layout(gtx,
									layout.Flexed(1, material.H6(th, p.Title.Value()).Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										if !p.ShowAdd {
											return layout.Dimensions{}
										}
										btn := material.IconButton(th, &p.btnAdd, icons.IconAdd, "Add")
										btn.Background = th.Bg
										btn.Color = th.Fg
										btn.Inset = layout.UniformInset(0)
										return btn.Layout(gtx)
									}),
								)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return p.List.Layout(gtx, len(p.Items), func(gtx layout.Context, index int) layout.Dimensions {
									if p.Items[index].state.Clicked(gtx) {
										if p.Multiple {
											p.Items[index].Selected = !p.Items[index].Selected
										} else {
											for i := range p.Items {
												p.Items[i].Selected = false
											}
											p.Items[index].Selected = true
										}
										if p.Selected != nil {
											p.Selected(index)
										}
									}
									return p.Items[index].Layout(gtx, th)
								})
							})
						}),
					)
				})
			})
		})
	})
}

type MenuItem struct {
	state    widget.Clickable
	Key      i18n.Key
	Value    string
	Selected bool
}

func (p *MenuItem) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return material.ButtonLayoutStyle{
		Background: th.Bg,
		Button:     &p.state,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    8,
			Bottom: 8,
			Left:   16,
			Right:  16,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Spacing:   layout.SpaceBetween,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(1, material.Body2(th, p.Key.Value()).Layout),
				layout.Rigid(layout.Spacer{Width: 8}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.Selected {
						gtx.Constraints.Min.X = gtx.Dp(16)
						return icons.IconDone.Layout(gtx, th.Fg)
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}
