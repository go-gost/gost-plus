package http

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"net"
	"strings"

	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/theme"
	ui_widget "github.com/go-gost/gost.plus/ui/widget"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type C = layout.Context
type D = layout.Dimensions

type httpPage struct {
	router *page.Router
	modal  *component.ModalLayer

	btnBack     widget.Clickable
	btnState    widget.Clickable
	btnDelete   widget.Clickable
	btnEdit     widget.Clickable
	btnSave     widget.Clickable
	btnFavorite widget.Clickable

	list layout.List

	wgID         widget.Clickable
	wgEntrypoint widget.Clickable

	name     component.TextField
	endpoint component.TextField

	rewriteHost widget.Bool
	hostname    component.TextField

	basicAuth widget.Bool
	username  component.TextField
	password  component.TextField

	enableTLS widget.Bool

	btnPasswordVisible widget.Clickable
	passwordVisible    bool

	id   string
	edit bool

	delDialog ui_widget.Dialog
}

func NewPage(r *page.Router) page.Page {
	return &httpPage{
		router: r,
		modal:  component.NewModal(),
		list: layout.List{
			// NOTE: the list must be vertical
			Axis: layout.Vertical,
		},
		name: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		endpoint: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		hostname: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		username: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		password: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		delDialog: ui_widget.Dialog{
			Title: i18n.Get(i18n.DeleteTunnel),
		},
	}
}

func (p *httpPage) Init(opts ...page.PageOption) {
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
	p.endpoint.Clear()

	p.rewriteHost.Value = false
	p.hostname.Clear()

	p.basicAuth.Value = false
	p.username.Clear()
	p.password.Clear()
	p.passwordVisible = false

	s := tunnel.Get(p.id)
	if s != nil {
		sopts := s.Options()
		p.name.SetText(sopts.Name)
		p.endpoint.SetText(sopts.Endpoint)
		if sopts.Hostname != "" {
			p.rewriteHost.Value = true
			p.hostname.SetText(sopts.Hostname)
		}
		if sopts.Username != "" {
			p.basicAuth.Value = true
			p.username.SetText(sopts.Username)
			p.password.SetText(sopts.Password)
		}
		p.enableTLS.Value = sopts.EnableTLS
	}
}

func (p *httpPage) Layout(gtx C) D {
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
						title := material.H6(th, "HTTP")
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx C) D {
						tun := tunnel.Get(p.id)
						if tun == nil {
							return D{}
						}

						if p.btnFavorite.Clicked(gtx) {
							tun.Favorite(!tun.IsFavorite())
							tunnel.SaveConfig()
						}

						btn := material.IconButton(th, &p.btnFavorite, icons.IconFavorite, "Favorite")

						if tun.IsFavorite() {
							btn.Color = color.NRGBA(colornames.Red500)
						} else {
							btn.Color = th.Fg
						}
						btn.Background = th.Bg

						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx C) D {
						tun := tunnel.Get(p.id)
						if tun == nil {
							return D{}
						}

						if p.btnState.Clicked(gtx) {
							p.onoff()
						}

						if !tun.IsClosed() {
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

func (p *httpPage) layout(gtx C, th *material.Theme) D {
	src := gtx.Source

	if !p.edit {
		gtx = gtx.Disabled()
	}

	tun := tunnel.Get(p.id)

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
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if tun == nil {
						return layout.Dimensions{}
					}

					gtx.Source = src

					copied := false
					if p.wgID.Clicked(gtx) {
						copied = true
						gtx.Execute(clipboard.WriteCmd{
							Data: io.NopCloser(bytes.NewBufferString(tun.ID())),
						})
					}

					return layout.Inset{
						Top:    8,
						Bottom: 8,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return p.wgID.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									label := material.Body1(th, tun.ID())
									label.Font.Weight = font.SemiBold
									return label.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: 8}.Layout),
								layout.Rigid(func(gtx C) D {
									c := color.NRGBA(colornames.Blue500)
									if copied {
										c = color.NRGBA(colornames.Green500)
									}
									return icons.IconCopy.Layout(gtx, c)
								}),
							)
						})
					})
				}),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if tun == nil {
						return layout.Dimensions{}
					}

					gtx.Source = src

					copied := false
					if p.wgEntrypoint.Clicked(gtx) {
						copied = true
						gtx.Execute(clipboard.WriteCmd{
							Data: io.NopCloser(bytes.NewBufferString(tun.Entrypoint())),
						})
					}

					return layout.Inset{
						Top:    8,
						Bottom: 8,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return p.wgEntrypoint.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									label := material.Body1(th, tun.Entrypoint())
									label.Font.Weight = font.SemiBold
									return label.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: 8}.Layout),
								layout.Rigid(func(gtx C) D {
									c := color.NRGBA(colornames.Blue500)
									if copied {
										c = color.NRGBA(colornames.Green500)
									}
									return icons.IconCopy.Layout(gtx, c)
								}),
							)
						})
					})
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),

				layout.Rigid(func(gtx C) D {
					return material.Body1(th, i18n.Get(i18n.Name)).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return p.name.Layout(gtx, th, "")
				}),
				layout.Rigid(layout.Spacer{Height: 16}.Layout),
				layout.Rigid(func(gtx C) D {
					return material.Body1(th, i18n.Get(i18n.Endpoint)).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if err := func() error {
						addr := strings.TrimSpace(p.endpoint.Text())
						if addr == "" {
							return nil
						}
						if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
							return fmt.Errorf(i18n.Get(i18n.ErrInvalidAddr))
						}
						return nil
					}(); err != nil {
						p.endpoint.SetError(err.Error())
					} else {
						p.endpoint.ClearError()
					}

					return p.endpoint.Layout(gtx, th, i18n.Get(i18n.Address))
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),

				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    8,
						Bottom: 8,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Spacing: layout.SpaceBetween,
						}.Layout(gtx,
							layout.Flexed(1, material.Body1(th, i18n.Get(i18n.BasicAuth)).Layout),
							layout.Rigid(material.Switch(th, &p.basicAuth, "Basic auth").Layout),
						)
					})
				}),

				layout.Rigid(func(gtx C) D {
					if !p.basicAuth.Value {
						p.username.Clear()
						return D{}
					}
					return p.username.Layout(gtx, th, i18n.Get(i18n.Username))
				}),
				layout.Rigid(func(gtx C) D {
					if !p.basicAuth.Value {
						p.password.Clear()
						return D{}
					}

					{
						gtx := gtx
						gtx.Source = src

						if p.btnPasswordVisible.Clicked(gtx) {
							p.passwordVisible = !p.passwordVisible
						}

						if p.passwordVisible {
							p.password.Suffix = func(gtx C) D {
								return p.btnPasswordVisible.Layout(gtx, func(gtx C) D {
									return icons.IconVisibility.Layout(gtx, color.NRGBA(colornames.Grey500))
								})
							}
							p.password.Mask = 0
						} else {
							p.password.Suffix = func(gtx C) D {
								return p.btnPasswordVisible.Layout(gtx, func(gtx C) D {
									return icons.IconVisibilityOff.Layout(gtx, color.NRGBA(colornames.Grey500))
								})
							}
							p.password.Mask = '*'
						}
					}

					return layout.Inset{
						Bottom: 8,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return p.password.Layout(gtx, th, i18n.Get(i18n.Password))
					})
				}),

				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    8,
						Bottom: 8,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Spacing: layout.SpaceBetween,
						}.Layout(gtx,
							layout.Flexed(1, material.Body1(th, i18n.Get(i18n.CustomHostname)).Layout),
							layout.Rigid(material.Switch(th, &p.rewriteHost, "Hostname").Layout),
						)
					})
				}),
				layout.Rigid(func(gtx C) D {
					if !p.rewriteHost.Value {
						p.hostname.SetText("")
						return layout.Dimensions{}
					}

					return layout.Inset{
						Bottom: 8,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return p.hostname.Layout(gtx, th, i18n.Get(i18n.Hostname))
					})
				}),

				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: 8, Bottom: 8}.Layout(gtx, func(gtx C) D {
						return layout.Flex{
							Spacing: layout.SpaceBetween,
						}.Layout(gtx,
							layout.Flexed(1, material.Body1(th, i18n.Get(i18n.EnableTLS)).Layout),
							layout.Rigid(material.Switch(th, &p.enableTLS, "enable TLS").Layout),
						)
					})
				}),
				layout.Rigid(layout.Spacer{Height: 8}.Layout),
			)
		})
	})
}

func (p *httpPage) create() error {
	defer tunnel.SaveConfig()

	var username, password string
	if p.basicAuth.Value {
		username = strings.TrimSpace(p.username.Text())
		password = strings.TrimSpace(p.password.Text())
	}
	var hostname string
	if p.rewriteHost.Value {
		hostname = strings.TrimSpace(p.hostname.Text())
	}
	tun := tunnel.NewHTTPTunnel(
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.EndpointOption(strings.TrimSpace(p.endpoint.Text())),
		tunnel.UsernameOption(username),
		tunnel.PasswordOption(password),
		tunnel.HostnameOption(hostname),
		tunnel.EnableTLSOption(p.enableTLS.Value),
	)

	tunnel.Add(tun)

	if err := tun.Run(); err != nil {
		tun.Close()
		return err
	}

	return nil
}

func (p *httpPage) update(opts ...tunnel.Option) tunnel.Tunnel {
	defer tunnel.SaveConfig()

	if t := tunnel.Get(p.id); t != nil {
		t.Close()
	}

	if opts == nil {
		var username, password string
		if p.basicAuth.Value {
			username = strings.TrimSpace(p.username.Text())
			password = strings.TrimSpace(p.password.Text())
		}
		var hostname string
		if p.rewriteHost.Value {
			hostname = strings.TrimSpace(p.hostname.Text())
		}
		opts = []tunnel.Option{
			tunnel.NameOption(strings.TrimSpace(p.name.Text())),
			tunnel.IDOption(p.id),
			tunnel.EndpointOption(strings.TrimSpace(p.endpoint.Text())),
			tunnel.UsernameOption(username),
			tunnel.PasswordOption(password),
			tunnel.HostnameOption(hostname),
			tunnel.EnableTLSOption(p.enableTLS.Value),
		}
	}
	tun := tunnel.NewHTTPTunnel(opts...)

	tunnel.Set(tun)

	if err := tun.Run(); err != nil {
		tun.Close()
		logger.Default().Error(err)
	}

	return tun
}

func (p *httpPage) onoff() {
	tun := tunnel.Get(p.id)
	if tun == nil {
		return
	}

	if tun.IsClosed() {
		opts := tun.Options()
		p.update(
			tunnel.NameOption(opts.Name),
			tunnel.IDOption(opts.ID),
			tunnel.EndpointOption(opts.Endpoint),
			tunnel.UsernameOption(opts.Username),
			tunnel.PasswordOption(opts.Password),
			tunnel.HostnameOption(opts.Hostname),
			tunnel.EnableTLSOption(opts.EnableTLS),
		)
	} else {
		tun.Close()
	}
	tunnel.SaveConfig()
}

func (p *httpPage) delete() {
	tunnel.Delete(p.id)
	tunnel.SaveConfig()
}
