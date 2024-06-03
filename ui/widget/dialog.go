package widget

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost.plus/ui/i18n"
)

type Dialog struct {
	Title     i18n.Key
	Body      string
	Widget    func(gtx layout.Context, th *material.Theme) layout.Dimensions
	Clicked   func(ok bool)
	btnCancel widget.Clickable
	btnOK     widget.Clickable
}

func (p *Dialog) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
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
					Left:   24,
					Right:  24,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.Title == "" {
								return layout.Dimensions{}
							}
							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return material.H6(th, p.Title.Value()).Layout(gtx)
							})
						}),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.Body == "" {
								return layout.Dimensions{}
							}
							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return material.Body1(th, p.Body).Layout(gtx)
							})
						}),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.Widget == nil {
								return layout.Dimensions{}
							}
							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return p.Widget(gtx, th)
							})
						}),

						layout.Rigid(layout.Spacer{Height: 8}.Layout),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Spacing:   layout.SpaceBetween,
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.Spacer{Width: 8}.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if p.btnCancel.Clicked(gtx) && p.Clicked != nil {
										p.Clicked(false)
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
									if p.btnOK.Clicked(gtx) && p.Clicked != nil {
										p.Clicked(true)
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
						}),
					)
				})
			})
		})
	})
}

type InputDialog struct {
	Surface   component.SurfaceStyle
	Title     string
	Body      string
	Hint      string
	Input     component.TextField
	Clicked   func(ok bool)
	btnCancel widget.Clickable
	btnOK     widget.Clickable
}

func (p *InputDialog) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if gtx.Constraints.Max.X > gtx.Dp(800) {
		gtx.Constraints.Max.X = gtx.Dp(800)
	}
	gtx.Constraints.Max.X = gtx.Constraints.Max.X * 2 / 3

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    16,
			Bottom: 16,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.Surface.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    16,
					Bottom: 16,
					Left:   24,
					Right:  24,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.Title == "" {
								return layout.Dimensions{}
							}
							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return material.H6(th, p.Title).Layout(gtx)
							})
						}),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.Body == "" {
								return layout.Dimensions{}
							}

							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return material.Body1(th, p.Body).Layout(gtx)
							})
						}),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    8,
								Bottom: 8,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return p.Input.Layout(gtx, th, p.Hint)
							})
						}),
						layout.Rigid(layout.Spacer{Height: 8}.Layout),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Spacing:   layout.SpaceBetween,
								Alignment: layout.End,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.Spacer{Width: 0}.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if p.btnCancel.Clicked(gtx) && p.Clicked != nil {
										p.Clicked(false)
									}

									return material.ButtonLayoutStyle{
										Background:   th.Bg,
										CornerRadius: 20,
										Button:       &p.btnCancel,
									}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return layout.Inset{
											Top:    8,
											Bottom: 8,
											Left:   24,
											Right:  24,
										}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											label := material.Body1(th, "Cancel")
											label.Color = th.Fg
											return label.Layout(gtx)
										})

									})
								}),
								layout.Rigid(layout.Spacer{Width: 8}.Layout),

								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if p.btnOK.Clicked(gtx) && p.Clicked != nil {
										p.Clicked(true)
									}

									return material.ButtonLayoutStyle{
										Background:   th.Bg,
										CornerRadius: 20,
										Button:       &p.btnOK,
									}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return layout.Inset{
											Top:    8,
											Bottom: 8,
											Left:   24,
											Right:  24,
										}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											label := material.Body1(th, "OK")
											label.Color = th.Fg
											return label.Layout(gtx)
										})

									})
								}),
							)
						}),
					)
				})

			})
		})
	})
}
