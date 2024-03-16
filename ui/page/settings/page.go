package settings

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/theme"
	ui_widget "github.com/go-gost/gost.plus/ui/widget"
	"github.com/go-gost/gost.plus/version"
)

type settingsPage struct {
	router *page.Router
	modal  *component.ModalLayer
	menu   ui_widget.Menu
	list   widget.List

	btnBack widget.Clickable

	lang  ui_widget.Selector
	theme ui_widget.Selector
}

func NewPage(r *page.Router) page.Page {
	return &settingsPage{
		router: r,
		modal:  component.NewModal(),
		menu: ui_widget.Menu{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		list: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		lang:  ui_widget.Selector{Title: "Language"},
		theme: ui_widget.Selector{Title: "Theme"},
	}
}

func (p *settingsPage) Init(opts ...page.PageOption) {
	settings := config.Get().Settings
	if settings == nil {
		settings = &config.Settings{}
	}
	if settings.Lang == "" {
		settings.Lang = "en_US"
	}
	if settings.Theme == "" {
		settings.Theme = theme.Light
	}

	p.lang.Clear()
	p.lang.Select(settings.Lang)

	p.theme.Clear()
	p.theme.Select(settings.Theme)
}

func (p *settingsPage) Layout(gtx layout.Context) layout.Dimensions {
	if p.btnBack.Clicked(gtx) {
		p.router.Back()
	}

	th := p.router.Theme

	defer p.modal.Layout(gtx, th)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    8,
				Bottom: 8,
				Left:   8,
				Right:  8,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.IconButton(th, &p.btnBack, icons.IconBack, "Back")
						btn.Color = th.Fg
						btn.Background = th.Bg
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H6(th, "Settings")
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
				)
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.list.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
				return layout.Inset{
					Top:    8,
					Bottom: 8,
					Left:   8,
					Right:  8,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return p.layout(gtx, th)
				})
			})
		}),
	)
}

func (p *settingsPage) layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return component.SurfaceStyle{
		Theme: th,
		ShadowStyle: component.ShadowStyle{
			CornerRadius: 12,
		},
		Fill: theme.Current().ContentSurfaceBg,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    8,
			Bottom: 8,
			Left:   8,
			Right:  8,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return icons.IconApp.Layout(gtx)
					})
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.H6(th, "GOST+")
					label.Font.Weight = font.Bold
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, version.Version).Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: 16}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.lang.Clicked(gtx) {
						p.showLangMenu(gtx)
					}
					return p.lang.Layout(gtx, th)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.theme.Clicked(gtx) {
						p.showThemeMenu(gtx)
					}
					return p.theme.Layout(gtx, th)
				}),
			)
		})
	})
}

func (p *settingsPage) showLangMenu(gtx layout.Context) {
	items := []ui_widget.MenuItem{
		{Key: "en_US", Value: "en_US"},
		{Key: "zh_CN", Value: "zh_CN"},
	}

	var found bool
	for i := range items {
		if found = p.lang.Any(items[i].Value); found {
			items[i].Selected = found
			break
		}
	}
	if !found {
		items[0].Selected = true
	}

	p.menu.Title = "Language"
	p.menu.Items = items
	p.menu.Selected = func(index int) {
		p.lang.Clear()
		p.lang.Select(p.menu.Items[index].Value)
		p.modal.Disappear(gtx.Now)

		cfg := config.Get()
		if cfg.Settings == nil {
			cfg.Settings = &config.Settings{}
		}
		cfg.Settings.Lang = p.lang.Value()

		config.Set(cfg)
		cfg.Write()
	}

	p.modal.Widget = func(gtx layout.Context, th *material.Theme, anim *component.VisibilityAnimation) layout.Dimensions {
		return p.menu.Layout(gtx, th)
	}
	p.modal.Appear(gtx.Now)
}

func (p *settingsPage) showThemeMenu(gtx layout.Context) {
	items := []ui_widget.MenuItem{
		{Key: "light", Value: theme.Light},
		{Key: "dark", Value: theme.Dark},
	}

	var found bool
	for i := range items {
		if found = p.theme.Any(items[i].Value); found {
			items[i].Selected = found
			break
		}
	}
	if !found {
		items[0].Selected = true
	}

	p.menu.Title = "Theme"
	p.menu.Items = items
	p.menu.Selected = func(index int) {
		p.theme.Clear()
		p.theme.Select(p.menu.Items[index].Value)
		p.modal.Disappear(gtx.Now)

		cfg := config.Get()
		if cfg.Settings == nil {
			cfg.Settings = &config.Settings{}
		}
		cfg.Settings.Theme = p.theme.Value()

		config.Set(cfg)
		cfg.Write()

		switch cfg.Settings.Theme {
		case theme.Dark:
			theme.UseDark()
		default:
			theme.UseLight()
		}
	}

	p.modal.Widget = func(gtx layout.Context, th *material.Theme, anim *component.VisibilityAnimation) layout.Dimensions {
		return p.menu.Layout(gtx, th)
	}
	p.modal.Appear(gtx.Now)
}
