package widget

import (
	"sync"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
)

type Menu struct {
	Title     i18n.Key
	Options   []MenuOption
	OnClick   func(ok bool)
	list      material.ListStyle
	btnAdd    widget.Clickable
	OnAdd     func()
	Multiple  bool
	btnCancel widget.Clickable
	btnOK     widget.Clickable
	once      sync.Once
}

func (p *Menu) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	p.once.Do(func() {
		p.list = material.List(th, &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		})
		p.list.AnchorStrategy = material.Overlay
	})

	if p.btnAdd.Clicked(gtx) && p.OnAdd != nil {
		p.OnAdd()
	}

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
								Left:  24,
								Right: 24,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Alignment: layout.Middle,
								}.Layout(gtx,
									layout.Flexed(1, material.H6(th, p.Title.Value()).Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										if p.OnAdd == nil {
											return layout.Dimensions{}
										}
										btn := material.IconButton(th, &p.btnAdd, icons.IconAdd, "Add")
										btn.Background = th.Bg
										btn.Color = th.Fg
										// btn.Inset = layout.UniformInset(8)
										return btn.Layout(gtx)
									}),
								)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Max.Y -= gtx.Dp(80)

							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return p.list.Layout(gtx, len(p.Options), func(gtx layout.Context, index int) layout.Dimensions {
									if p.Options[index].state.Clicked(gtx) {
										p.Options[index].Selected = !p.Options[index].Selected

										if !p.Multiple {
											for i := range p.Options {
												if i != index {
													p.Options[i].Selected = false
												}
											}
										}
									}

									return p.Options[index].Layout(gtx, th)
								})
							})
						}),

						layout.Rigid(layout.Spacer{Height: 8}.Layout),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Left:  24,
								Right: 24,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Spacing:   layout.SpaceBetween,
									Alignment: layout.Middle,
								}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return layout.Spacer{Width: 8}.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										if p.btnCancel.Clicked(gtx) && p.OnClick != nil {
											p.OnClick(false)
										}

										return material.ButtonLayoutStyle{
											Background:   th.Bg,
											CornerRadius: 18,
											Button:       &p.btnCancel,
										}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{
												Top:    8,
												Bottom: 8,
												Left:   20,
												Right:  20,
											}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												label := material.Body1(th, i18n.Cancel.Value())
												label.Color = th.Fg
												return label.Layout(gtx)
											})

										})
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return layout.Spacer{Width: 8}.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										if p.btnOK.Clicked(gtx) && p.OnClick != nil {
											p.OnClick(true)
										}

										return material.ButtonLayoutStyle{
											Background:   th.Bg,
											CornerRadius: 18,
											Button:       &p.btnOK,
										}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{
												Top:    8,
												Bottom: 8,
												Left:   20,
												Right:  20,
											}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												label := material.Body1(th, i18n.OK.Value())
												label.Color = th.Fg
												return label.Layout(gtx)
											})

										})
									}),
								)
							})
						}),
					)
				})
			})
		})
	})
}

type MenuOption struct {
	state    widget.Clickable
	Key      i18n.Key
	Name     string
	Value    string
	Selected bool
}

func (p *MenuOption) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return material.ButtonLayoutStyle{
		Background: th.Bg,
		Button:     &p.state,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    12,
			Bottom: 12,
			Left:   24,
			Right:  24,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Spacing:   layout.SpaceBetween,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					name := p.Name
					if name == "" {
						name = p.Key.Value()
					}
					if name == "" {
						name = p.Value
					}
					return material.Body2(th, name).Layout(gtx)
				}),
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
