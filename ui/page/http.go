package page

import (
	"fmt"
	"image/color"
	"log"
	"net"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost-plus/tunnel"
	"github.com/go-gost/gost-plus/ui/icons"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type httpAddPage struct {
	router *Router

	list   layout.List
	wgDone widget.Clickable

	name component.TextField
	addr component.TextField

	bHost    widget.Bool
	hostname component.TextField

	bBasicAuth widget.Bool
	username   component.TextField
	password   component.TextField

	bTLS widget.Bool
}

func NewHTTPAddPage(r *Router) Page {
	return &httpAddPage{
		router: r,
		list: layout.List{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		},
		name: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		addr: component.TextField{
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
	}
}

func (p *httpAddPage) Init(opts ...PageOption) {
	p.name.SetText("")
	p.addr.SetText("")
	p.hostname.SetText("")
	p.bBasicAuth.Value = false
	p.username.SetText("")
	p.password.SetText("")
	p.bTLS.Value = false

	p.router.bar.SetActions(
		[]component.AppBarAction{
			{
				OverflowAction: component.OverflowAction{
					Name: "Create",
					Tag:  &p.wgDone,
				},
				Layout: func(gtx C, bg, fg color.NRGBA) D {
					if p.wgDone.Clicked(gtx) {
						defer p.router.SwitchTo(Route{Path: PageTunnel})
						p.createTunnel()
						tunnel.SaveConfig()
					}
					return component.SimpleIconButton(bg, fg, &p.wgDone, icons.IconDone).Layout(gtx)
				},
			},
		}, nil)

	p.router.bar.Title = "HTTP"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *httpAddPage) Layout(gtx C, th *material.Theme) D {
	return p.list.Layout(gtx, 1, func(gtx C, _ int) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
				return component.Surface(th).Layout(gtx, func(gtx C) D {
					return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
						return p.layout(gtx, th)
					})
				})
			})
		})
	})
}

func (p *httpAddPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(material.Body1(th, "Expose local http tunnel to public network.").Layout),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Tunnel name").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return p.name.Layout(gtx, th, "Name")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Endpoint address").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if err := func() error {
				addr := strings.TrimSpace(p.addr.Text())
				if addr == "" {
					return nil
				}
				if _, _, err := net.SplitHostPort(addr); err != nil {
					return fmt.Errorf("invalid address format, should be [IP]:PORT or [HOST]:PORT")
				}
				return nil
			}(); err != nil {
				p.addr.SetError(err.Error())
			} else {
				p.addr.ClearError()
			}

			return p.addr.Layout(gtx, th, "Address")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, material.Body1(th, "Use custom hostname (rewrite HTTP Host header)").Layout),
				layout.Rigid(material.Switch(th, &p.bHost, "custom hostname").Layout),
			)
		}),
		layout.Rigid(func(gtx C) D {
			if !p.bHost.Value {
				p.hostname.SetText("")
				return layout.Dimensions{}
			}
			return p.hostname.Layout(gtx, th, "Hostname")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, material.Body1(th, "Use basic auth").Layout),
				layout.Rigid(material.Switch(th, &p.bBasicAuth, "use basic auth").Layout),
			)
		}),
		layout.Rigid(func(gtx C) D {
			if !p.bBasicAuth.Value {
				p.username.SetText("")
				return layout.Dimensions{}
			}
			return p.username.Layout(gtx, th, "Username")
		}),
		layout.Rigid(func(gtx C) D {
			if !p.bBasicAuth.Value {
				p.password.SetText("")
				return layout.Dimensions{}
			}
			return p.password.Layout(gtx, th, "Password")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, material.Body1(th, "Enable TLS").Layout),
				layout.Rigid(material.Switch(th, &p.bTLS, "enable TLS").Layout),
			)
		}),
	)
}

func (p *httpAddPage) createTunnel() error {
	var username, password string
	if p.bBasicAuth.Value {
		username = strings.TrimSpace(p.username.Text())
		password = strings.TrimSpace(p.password.Text())
	}
	var hostname string
	if p.bHost.Value {
		hostname = strings.TrimSpace(p.hostname.Text())
	}
	tun := tunnel.NewHTTPTunnel(
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.EndpointOption(strings.TrimSpace(p.addr.Text())),
		tunnel.UsernameOption(username),
		tunnel.PasswordOption(password),
		tunnel.HostnameOption(hostname),
		tunnel.EnableTLSOption(p.bTLS.Value),
	)

	tunnel.Add(tun)

	if err := tun.Run(); err != nil {
		tun.Close()
		return err
	}

	return nil
}

type httpEditPage struct {
	router *Router

	id string

	list       layout.List
	wgFavorite widget.Clickable
	wgState    widget.Clickable
	wgDelete   widget.Clickable
	wgDone     widget.Clickable

	name component.TextField
	addr component.TextField

	bHost    widget.Bool
	hostname component.TextField

	cbBasicAuth widget.Bool
	username    component.TextField
	password    component.TextField

	bTLS widget.Bool

	wgID         widget.Clickable
	wgEntrypoint widget.Clickable
}

func NewHTTPEditPage(r *Router) Page {
	return &httpEditPage{
		router: r,
		list: layout.List{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		},
		name: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		addr: component.TextField{
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
	}
}

func (p *httpEditPage) Init(opts ...PageOption) {
	var options PageOptions
	for _, opt := range opts {
		opt(&options)
	}

	p.id = options.ID
	s := tunnel.Get(p.id)
	if s != nil {
		sopts := s.Options()
		p.name.SetText(sopts.Name)
		p.addr.SetText(sopts.Endpoint)
		if sopts.Hostname != "" {
			p.bHost.Value = true
			p.hostname.SetText(sopts.Hostname)
		}
		if sopts.Username != "" {
			p.cbBasicAuth.Value = true
			p.username.SetText(sopts.Username)
			p.password.SetText(sopts.Password)
		}
		p.bTLS.Value = sopts.EnableTLS
	}

	actions := []component.AppBarAction{
		{
			OverflowAction: component.OverflowAction{
				Name: "Favorite",
				Tag:  &p.wgFavorite,
			},
			Layout: func(gtx C, bg, fg color.NRGBA) D {
				s := tunnel.Get(p.id)
				if s == nil {
					return D{}
				}

				if p.wgFavorite.Clicked(gtx) {
					s.Favorite(!s.IsFavorite())
					tunnel.SaveConfig()
				}

				btn := component.SimpleIconButton(bg, fg, &p.wgFavorite, icons.IconFavorite)
				if s.IsFavorite() {
					btn.Color = color.NRGBA(colornames.Red500)
				} else {
					btn.Color = fg
				}
				return btn.Layout(gtx)
			},
		},
		{
			OverflowAction: component.OverflowAction{
				Name: "Start/Stop",
				Tag:  &p.wgState,
			},
			Layout: func(gtx C, bg, fg color.NRGBA) D {
				s := tunnel.Get(p.id)
				if p.wgState.Clicked(gtx) && s != nil {
					if s.IsClosed() {
						s = p.createTunnel()
					} else {
						s.Close()
					}
					tunnel.SaveConfig()
				}

				if s != nil && !s.IsClosed() {
					return component.SimpleIconButton(bg, fg, &p.wgState, icons.IconStop).Layout(gtx)
				} else {
					return component.SimpleIconButton(bg, fg, &p.wgState, icons.IconStart).Layout(gtx)
				}
			},
		},
		{
			OverflowAction: component.OverflowAction{
				Name: "Delete",
				Tag:  &p.wgDelete,
			},
			Layout: func(gtx C, bg, fg color.NRGBA) D {
				if p.wgDelete.Clicked(gtx) {
					tunnel.Delete(p.id)
					tunnel.SaveConfig()
					p.router.SwitchTo(Route{Path: PageTunnel})
				}
				return component.SimpleIconButton(bg, fg, &p.wgDelete, icons.IconDelete).Layout(gtx)
			},
		},
		{
			OverflowAction: component.OverflowAction{
				Name: "Save",
				Tag:  &p.wgDone,
			},
			Layout: func(gtx C, bg, fg color.NRGBA) D {
				if p.wgDone.Clicked(gtx) {
					defer p.router.SwitchTo(Route{Path: PageTunnel})

					if s := tunnel.Get(p.id); s != nil {
						s.Close()
						p.createTunnel()
						tunnel.SaveConfig()
					}
				}
				return component.SimpleIconButton(bg, fg, &p.wgDone, icons.IconDone).Layout(gtx)
			},
		},
	}
	p.router.bar.SetActions(actions, nil)
	p.router.bar.Title = "HTTP"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *httpEditPage) createTunnel() tunnel.Tunnel {
	var username, password string
	if p.cbBasicAuth.Value {
		username = strings.TrimSpace(p.username.Text())
		password = strings.TrimSpace(p.password.Text())
	}
	var hostname string
	if p.bHost.Value {
		hostname = strings.TrimSpace(p.hostname.Text())
	}
	tun := tunnel.NewHTTPTunnel(
		tunnel.IDOption(p.id),
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.EndpointOption(strings.TrimSpace(p.addr.Text())),
		tunnel.UsernameOption(username),
		tunnel.PasswordOption(password),
		tunnel.HostnameOption(hostname),
		tunnel.EnableTLSOption(p.bTLS.Value),
	)

	tunnel.Set(tun)

	if err := tun.Run(); err != nil {
		tun.Close()
		log.Println(err)
	}

	return tun
}

func (p *httpEditPage) Layout(gtx C, th *material.Theme) D {
	return p.list.Layout(gtx, 1, func(gtx C, _ int) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
				return component.Surface(th).Layout(gtx, func(gtx C) D {
					return layout.UniformInset(10).Layout(gtx, func(gtx C) D {
						return p.layout(gtx, th)
					})
				})
			})
		})
	})
}

func (p *httpEditPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layoutHeader(gtx, th, tunnel.Get(p.id), &p.wgID, &p.wgEntrypoint)
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Tunnel name").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return p.name.Layout(gtx, th, "Name")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Endpoint address").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if err := func() error {
				addr := strings.TrimSpace(p.addr.Text())
				if addr == "" {
					return nil
				}
				if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
					return fmt.Errorf("invalid address format, should be [IP]:PORT or [HOST]:PORT")
				}
				return nil
			}(); err != nil {
				p.addr.SetError(err.Error())
			} else {
				p.addr.ClearError()
			}

			return p.addr.Layout(gtx, th, "Address")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: 10, Bottom: 10}.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(1, material.Body1(th, "Use custom hostname (rewrite HTTP Host header)").Layout),
					layout.Rigid(material.Switch(th, &p.bHost, "custom hostname").Layout),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			if !p.bHost.Value {
				p.hostname.SetText("")
				return layout.Dimensions{}
			}
			return p.hostname.Layout(gtx, th, "Hostname")
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: 10, Bottom: 10}.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(1, material.Body1(th, "Use basic auth").Layout),
					layout.Rigid(material.Switch(th, &p.cbBasicAuth, "use basic auth").Layout),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			if !p.cbBasicAuth.Value {
				p.username.SetText("")
				return layout.Dimensions{}
			}
			return p.username.Layout(gtx, th, "Username")
		}),
		layout.Rigid(func(gtx C) D {
			if !p.cbBasicAuth.Value {
				p.password.SetText("")
				return layout.Dimensions{}
			}
			return p.password.Layout(gtx, th, "Password")
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: 10, Bottom: 10}.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(1, material.Body1(th, "Enable TLS").Layout),
					layout.Rigid(material.Switch(th, &p.bTLS, "enable TLS").Layout),
				)
			})
		}),
	)
}
