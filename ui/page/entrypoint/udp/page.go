package udp

import (
	"fmt"
	"image/color"
	"net"
	"strconv"
	"strings"
	"unicode"

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

type udpPage struct {
	router *page.Router
	modal  *component.ModalLayer

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

	keepalive widget.Bool
	ttl       component.TextField

	id   string
	edit bool

	delDialog ui_widget.Dialog
}

func NewPage(r *page.Router) page.Page {
	return &udpPage{
		router: r,
		modal:  component.NewModal(),
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
		ttl: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
			Suffix: func(gtx C) D {
				return material.Label(r.Theme, r.Theme.TextSize, "s").Layout(gtx)
			},
		},
		delDialog: ui_widget.Dialog{
			Title: i18n.Get(i18n.DeleteEntrypoint),
		},
	}
}

func (p *udpPage) Init(opts ...page.PageOption) {
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

	p.name.Clear()
	p.entrypoint.Clear()

	p.keepalive.Value = false
	p.ttl.Clear()

	s := entrypoint.Get(p.id)
	if s != nil {
		sopts := s.Options()
		p.tunnelID.SetText(sopts.ID)
		p.name.SetText(sopts.Name)
		p.entrypoint.SetText(sopts.Endpoint)
		p.keepalive.Value = sopts.Keepalive
		p.ttl.SetText(strconv.Itoa(sopts.TTL))
	}
}

func (p *udpPage) Layout(gtx C) D {
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
			p.modal.Disappear(gtx.Now)
		}
		p.modal.Widget = func(gtx layout.Context, th *material.Theme, anim *component.VisibilityAnimation) layout.Dimensions {
			return p.delDialog.Layout(gtx, th)
		}
		p.modal.Appear(gtx.Now)
	}

	th := p.router.Theme

	defer p.modal.Layout(gtx, th)

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
						title := material.H6(th, "UDP")
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

func (p *udpPage) layout(gtx C, th *material.Theme) D {
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
					return material.Body1(th, i18n.Get(i18n.TunnelID)).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if err := func() error {
						tid := strings.TrimSpace(p.tunnelID.Text())
						if tid == "" {
							return nil
						}
						if _, err := uuid.Parse(tid); err != nil {
							return fmt.Errorf(i18n.Get(i18n.ErrInvalidTunnelID))
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
					return material.Body1(th, i18n.Get(i18n.Name)).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return p.name.Layout(gtx, th, "")
				}),
				layout.Rigid(layout.Spacer{Height: 16}.Layout),

				layout.Rigid(func(gtx C) D {
					return material.Body1(th, i18n.Get(i18n.Entrypoint)).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if err := func() error {
						addr := strings.TrimSpace(p.entrypoint.Text())
						if addr == "" {
							return nil
						}
						if _, err := net.ResolveUDPAddr("udp", addr); err != nil {
							return fmt.Errorf(i18n.Get(i18n.ErrInvalidAddr))
						}
						return nil
					}(); err != nil {
						p.entrypoint.SetError(err.Error())
					} else {
						p.entrypoint.ClearError()
					}

					return p.entrypoint.Layout(gtx, th, i18n.Get(i18n.Address))
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),

				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    8,
						Bottom: 8,
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{
							Spacing: layout.SpaceBetween,
						}.Layout(gtx,
							layout.Flexed(1, material.Body1(th, i18n.Get(i18n.Keepalive)).Layout),
							layout.Rigid(material.Switch(th, &p.keepalive, "keepalive").Layout),
						)
					})
				}),
				layout.Rigid(func(gtx C) D {
					if !p.keepalive.Value {
						p.ttl.Clear()
						return layout.Dimensions{}
					}

					if err := func() string {
						for _, r := range p.ttl.Text() {
							if !unicode.IsDigit(r) {
								return i18n.Get(i18n.ErrDigitOnly)
							}
						}
						return ""
					}(); err != "" {
						p.ttl.SetError(err)
					} else {
						p.ttl.ClearError()
					}
					return p.ttl.Layout(gtx, th, "TTL")
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),
			)
		})
	})
}

func (p *udpPage) create() error {
	defer entrypoint.SaveConfig()

	ttl, _ := strconv.Atoi(strings.TrimSpace(p.ttl.Text()))

	ep := entrypoint.NewUDPEntryPoint(
		tunnel.IDOption(strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))),
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.EndpointOption(strings.TrimSpace(p.entrypoint.Text())),
		tunnel.KeepaliveOption(p.keepalive.Value),
		tunnel.TTLOption(ttl),
	)

	entrypoint.Add(ep)

	if err := ep.Run(); err != nil {
		ep.Close()
		return err
	}

	return nil
}

func (p *udpPage) update(opts ...tunnel.Option) tunnel.Tunnel {
	defer entrypoint.SaveConfig()

	if t := entrypoint.Get(p.id); t != nil {
		t.Close()
	}

	if opts == nil {
		ttl, _ := strconv.Atoi(strings.TrimSpace(p.ttl.Text()))

		opts = []tunnel.Option{
			tunnel.IDOption(strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))),
			tunnel.NameOption(strings.TrimSpace(p.name.Text())),
			tunnel.EndpointOption(strings.TrimSpace(p.entrypoint.Text())),
			tunnel.KeepaliveOption(p.keepalive.Value),
			tunnel.TTLOption(ttl),
		}
	}
	ep := entrypoint.NewUDPEntryPoint(opts...)

	entrypoint.Set(ep)

	if err := ep.Run(); err != nil {
		ep.Close()
		logger.Default().Error(err)
	}

	return ep
}

func (p *udpPage) onoff() {
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
			tunnel.KeepaliveOption(opts.Keepalive),
			tunnel.TTLOption(opts.TTL),
		)
	} else {
		ep.Close()
	}
	entrypoint.SaveConfig()
}

func (p *udpPage) delete() {
	entrypoint.Delete(p.id)
	entrypoint.SaveConfig()
}
