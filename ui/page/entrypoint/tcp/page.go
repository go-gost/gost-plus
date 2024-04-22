package tcp

import (
	"fmt"
	"image/color"
	"net"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/theme"
	ui_widget "github.com/go-gost/gost.plus/ui/widget"
	"github.com/google/uuid"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type C = layout.Context
type D = layout.Dimensions

type tcpPage struct {
	router *page.Router

	btnBack     widget.Clickable
	btnState    widget.Clickable
	btnDelete   widget.Clickable
	btnEdit     widget.Clickable
	btnSave     widget.Clickable
	btnFavorite widget.Clickable

	list layout.List

	tunnelID   component.TextField
	name       component.TextField
	entrypoint component.TextField

	id   string
	edit bool

	delDialog ui_widget.Dialog
}

func NewPage(r *page.Router) page.Page {
	return &tcpPage{
		router: r,
		list: layout.List{
			// NOTE: the list must be vertical
			Axis: layout.Vertical,
		},
		tunnelID: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		name: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		entrypoint: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		delDialog: ui_widget.Dialog{
			Title: i18n.DeleteEntrypoint,
		},
	}
}

func (p *tcpPage) Init(opts ...page.PageOption) {
	var options page.PageOptions
	for _, opt := range opts {
		opt(&options)
	}

	p.id = options.ID

	if p.id != "" {
		p.edit = false
	} else {
		p.edit = true
	}

	p.tunnelID.Clear()
	p.name.Clear()
	p.entrypoint.Clear()

	s := entrypoint.Get(p.id)
	if s != nil {
		sopts := s.Options()
		p.tunnelID.SetText(sopts.ID)
		p.name.SetText(sopts.Name)
		p.entrypoint.SetText(sopts.Endpoint)
	}
}

func (p *tcpPage) Layout(gtx C) D {
	if p.btnBack.Clicked(gtx) {
		p.router.Back()
	}
	if p.btnEdit.Clicked(gtx) {
		p.edit = true
	}

	if p.btnSave.Clicked(gtx) {
		if p.id == "" {
			p.create()
		} else {
			p.update()
		}
		p.router.Back()
	}

	if p.btnDelete.Clicked(gtx) {
		p.delDialog.Clicked = func(ok bool) {
			if ok {
				p.delete()
				p.router.Back()
			}
			p.router.HideModal(gtx)
		}
		p.router.ShowModal(gtx, func(gtx page.C, th *material.Theme) page.D {
			return p.delDialog.Layout(gtx, th)
		})
	}

	th := p.router.Theme

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
					layout.Rigid(func(gtx C) D {
						btn := material.IconButton(th, &p.btnBack, icons.IconBack, "Back")
						btn.Color = th.Fg
						btn.Background = th.Bg
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Flexed(1, func(gtx C) D {
						title := material.H6(th, "TCP")
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx C) D {
						ep := entrypoint.Get(p.id)
						if ep == nil {
							return D{}
						}

						if p.btnFavorite.Clicked(gtx) {
							ep.Favorite(!ep.IsFavorite())
							entrypoint.SaveConfig()
						}

						btn := material.IconButton(th, &p.btnFavorite, icons.IconFavorite, "Favorite")

						if ep.IsFavorite() {
							btn.Color = color.NRGBA(colornames.Red500)
						} else {
							btn.Color = th.Fg
						}
						btn.Background = th.Bg

						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx C) D {
						ep := entrypoint.Get(p.id)
						if ep == nil {
							return D{}
						}

						if p.btnState.Clicked(gtx) {
							p.onoff()
						}

						if !ep.IsClosed() {
							btn := material.IconButton(th, &p.btnState, icons.IconStop, "Stop")

							btn.Color = th.Fg
							btn.Background = th.Bg
							return btn.Layout(gtx)
						}

						btn := material.IconButton(th, &p.btnState, icons.IconStart, "Start")

						btn.Color = th.Fg
						btn.Background = th.Bg
						return btn.Layout(gtx)

					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx C) D {
						if p.id == "" {
							return D{}
						}
						btn := material.IconButton(th, &p.btnDelete, icons.IconDelete, "Delete")

						btn.Color = th.Fg
						btn.Background = th.Bg
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx C) D {
						if p.edit {
							btn := material.IconButton(th, &p.btnSave, icons.IconDone, "Done")
							btn.Color = th.Fg
							btn.Background = th.Bg
							return btn.Layout(gtx)
						} else {
							btn := material.IconButton(th, &p.btnEdit, icons.IconEdit, "Edit")
							btn.Color = th.Fg
							btn.Background = th.Bg
							return btn.Layout(gtx)
						}
					}),
				)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return p.list.Layout(gtx, 1, func(gtx C, _ int) D {
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

func (p *tcpPage) layout(gtx C, th *material.Theme) D {
	if !p.edit {
		gtx = gtx.Disabled()
	}

	return component.SurfaceStyle{
		Theme: th,
		ShadowStyle: component.ShadowStyle{
			CornerRadius: 12,
		},
		Fill: theme.Current().ContentSurfaceBg,
	}.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(16).Layout(gtx, func(gtx C) D {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return material.Body1(th, i18n.TunnelID.Value()).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if err := func() error {
						tid := strings.TrimSpace(p.tunnelID.Text())
						if tid == "" {
							return nil
						}
						if _, err := uuid.Parse(tid); err != nil {
							return fmt.Errorf(i18n.ErrInvalidTunnelID.Value())
						}
						return nil
					}(); err != nil {
						p.tunnelID.SetError(err.Error())
					} else {
						p.tunnelID.ClearError()
					}

					if p.id != "" {
						gtx = gtx.Disabled()
					}
					return p.tunnelID.Layout(gtx, th, "")
				}),
				layout.Rigid(layout.Spacer{Height: 16}.Layout),

				layout.Rigid(func(gtx C) D {
					return material.Body1(th, i18n.Name.Value()).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return p.name.Layout(gtx, th, "")
				}),
				layout.Rigid(layout.Spacer{Height: 16}.Layout),

				layout.Rigid(func(gtx C) D {
					return material.Body1(th, i18n.Entrypoint.Value()).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if err := func() error {
						addr := strings.TrimSpace(p.entrypoint.Text())
						if addr == "" {
							return nil
						}
						if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
							return fmt.Errorf(i18n.ErrInvalidAddr.Value())
						}
						return nil
					}(); err != nil {
						p.entrypoint.SetError(err.Error())
					} else {
						p.entrypoint.ClearError()
					}

					return p.entrypoint.Layout(gtx, th, i18n.Address.Value())
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),
			)
		})
	})
}

func (p *tcpPage) create() error {
	defer entrypoint.SaveConfig()

	ep := entrypoint.NewTCPEntryPoint(
		tunnel.IDOption(strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))),
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.EndpointOption(strings.TrimSpace(p.entrypoint.Text())),
	)

	entrypoint.Add(ep)

	if err := ep.Run(); err != nil {
		ep.Close()
		return err
	}

	return nil
}

func (p *tcpPage) update(opts ...tunnel.Option) tunnel.Tunnel {
	defer entrypoint.SaveConfig()

	if t := entrypoint.Get(p.id); t != nil {
		t.Close()
	}

	if opts == nil {
		opts = []tunnel.Option{
			tunnel.IDOption(strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))),
			tunnel.NameOption(strings.TrimSpace(p.name.Text())),
			tunnel.EndpointOption(strings.TrimSpace(p.entrypoint.Text())),
		}
	}
	ep := entrypoint.NewTCPEntryPoint(opts...)

	entrypoint.Set(ep)

	if err := ep.Run(); err != nil {
		ep.Close()
		logger.Default().Error(err)
	}

	return ep
}

func (p *tcpPage) onoff() {
	ep := entrypoint.Get(p.id)
	if ep == nil {
		return
	}

	if ep.IsClosed() {
		opts := ep.Options()
		p.update(
			tunnel.IDOption(opts.ID),
			tunnel.NameOption(opts.Name),
			tunnel.EndpointOption(opts.Endpoint),
			tunnel.CreatedAtOption(opts.CreatedAt),
		)
	} else {
		ep.Close()
	}
	entrypoint.SaveConfig()
}

func (p *tcpPage) delete() {
	entrypoint.Delete(p.id)
	entrypoint.SaveConfig()
}
