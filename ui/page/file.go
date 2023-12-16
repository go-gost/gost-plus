package page

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost-plus/tunnel"
	"github.com/go-gost/gost-plus/ui/icons"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type fileAddPage struct {
	router *Router

	list   layout.List
	wgDone widget.Clickable

	name        component.TextField
	path        component.TextField
	cbBasicAuth widget.Bool
	username    component.TextField
	password    component.TextField

	wgPassword      widget.Clickable
	passwordVisible bool
}

func NewFileAddPage(r *Router) Page {
	return &fileAddPage{
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
		path: component.TextField{
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

func (p *fileAddPage) Init(opts ...PageOption) {
	p.name.Clear()
	p.path.Clear()
	p.cbBasicAuth.Value = false
	p.username.Clear()
	p.password.Clear()
	p.passwordVisible = false

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

	p.router.bar.Title = "File"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *fileAddPage) Layout(gtx C, th *material.Theme) D {
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

func (p *fileAddPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(material.Body1(th, "Expose local files to public network.").Layout),
		layout.Rigid(layout.Spacer{Height: 15}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Service name").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return p.name.Layout(gtx, th, "Name")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Root directory, default to the current working directory").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if err := func() error {
				dir := strings.TrimSpace(p.path.Text())
				if dir == "" {
					return nil
				}
				f, err := os.Open(dir)
				if err != nil {
					return err
				}
				defer f.Close()
				fs, err := f.Stat()
				if err != nil {
					return err
				}
				if !fs.IsDir() {
					return fmt.Errorf("%s is not a directory", dir)
				}
				return nil
			}(); err != nil {
				p.path.SetError(err.Error())
			} else {
				p.path.ClearError()
			}

			return p.path.Layout(gtx, th, "Path")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
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
				p.username.Clear()
				return layout.Dimensions{}
			}
			return p.username.Layout(gtx, th, "Username")
		}),
		layout.Rigid(func(gtx C) D {
			if !p.cbBasicAuth.Value {
				p.password.Clear()
				return layout.Dimensions{}
			}

			if p.wgPassword.Clicked(gtx) {
				p.passwordVisible = !p.passwordVisible
			}

			if p.passwordVisible {
				p.password.Suffix = func(gtx layout.Context) layout.Dimensions {
					return p.wgPassword.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return icons.IconVisibility.Layout(gtx, color.NRGBA(colornames.Grey500))
					})
				}
				p.password.Mask = 0
			} else {
				p.password.Suffix = func(gtx layout.Context) layout.Dimensions {
					return p.wgPassword.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return icons.IconVisibilityOff.Layout(gtx, color.NRGBA(colornames.Grey500))
					})
				}
				p.password.Mask = '*'
			}
			return p.password.Layout(gtx, th, "Password")
		}),
	)
}

func (p *fileAddPage) createTunnel() error {
	var username, password string
	if p.cbBasicAuth.Value {
		username = strings.TrimSpace(p.username.Text())
		password = strings.TrimSpace(p.password.Text())
	}
	tun := tunnel.NewFileTunnel(
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.EndpointOption(strings.TrimSpace(p.path.Text())),
		tunnel.UsernameOption(username),
		tunnel.PasswordOption(password),
	)

	tunnel.Add(tun)

	if err := tun.Run(); err != nil {
		tun.Close()
		return err
	}

	return nil
}

type fileEditPage struct {
	router *Router

	id string

	list       layout.List
	wgFavorite widget.Clickable
	wgState    widget.Clickable
	wgDelete   widget.Clickable
	wgDone     widget.Clickable

	wgID         widget.Clickable
	wgEntrypoint widget.Clickable

	name        component.TextField
	path        component.TextField
	cbBasicAuth widget.Bool
	username    component.TextField
	password    component.TextField

	wgPassword      widget.Clickable
	passwordVisible bool
}

func NewFileEditPage(r *Router) Page {
	return &fileEditPage{
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
		path: component.TextField{
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

func (p *fileEditPage) Init(opts ...PageOption) {
	p.passwordVisible = false

	var options PageOptions
	for _, opt := range opts {
		opt(&options)
	}

	p.id = options.ID
	s := tunnel.Get(p.id)
	if s != nil {
		sopts := s.Options()
		p.name.SetText(sopts.Name)
		p.path.SetText(sopts.Endpoint)
		if sopts.Username != "" {
			p.cbBasicAuth.Value = true
			p.username.SetText(sopts.Username)
			p.password.SetText(sopts.Password)
		}
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
				if s == nil {
					return D{}
				}

				if p.wgState.Clicked(gtx) {
					if s.IsClosed() {
						opts := s.Options()
						s = p.createTunnel(
							tunnel.NameOption(opts.Name),
							tunnel.IDOption(opts.ID),
							tunnel.EndpointOption(opts.Endpoint),
							tunnel.UsernameOption(opts.Username),
							tunnel.PasswordOption(opts.Password),
						)
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
	p.router.bar.Title = "File"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *fileEditPage) createTunnel(opts ...tunnel.Option) tunnel.Tunnel {
	if opts == nil {
		var username, password string
		if p.cbBasicAuth.Value {
			username = strings.TrimSpace(p.username.Text())
			password = strings.TrimSpace(p.password.Text())
		}
		opts = []tunnel.Option{
			tunnel.NameOption(strings.TrimSpace(p.name.Text())),
			tunnel.IDOption(p.id),
			tunnel.EndpointOption(strings.TrimSpace(p.path.Text())),
			tunnel.UsernameOption(username),
			tunnel.PasswordOption(password),
		}
	}
	tun := tunnel.NewFileTunnel(opts...)

	tunnel.Set(tun)

	if err := tun.Run(); err != nil {
		tun.Close()
		log.Println(err)
	}

	return tun
}

func (p *fileEditPage) Layout(gtx C, th *material.Theme) D {
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

func (p *fileEditPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layoutHeader(gtx, th, tunnel.Get(p.id), &p.wgID, &p.wgEntrypoint)
		}),
		layout.Rigid(func(gtx C) D {
			div := component.Divider(th)
			return div.Layout(gtx)
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
			return material.Body1(th, "Root directory, default to the current working directory").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if err := func() error {
				dir := strings.TrimSpace(p.path.Text())
				if dir == "" {
					return nil
				}
				f, err := os.Open(dir)
				if err != nil {
					return err
				}
				defer f.Close()
				fs, err := f.Stat()
				if err != nil {
					return err
				}
				if !fs.IsDir() {
					return fmt.Errorf("%s is not a directory", dir)
				}
				return nil
			}(); err != nil {
				p.path.SetError(err.Error())
			} else {
				p.path.ClearError()
			}

			return p.path.Layout(gtx, th, "Path")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, material.Body1(th, "Use basic auth").Layout),
				layout.Rigid(material.Switch(th, &p.cbBasicAuth, "use basic auth").Layout),
			)
		}),
		layout.Rigid(func(gtx C) D {
			if !p.cbBasicAuth.Value {
				p.username.Clear()
				return layout.Dimensions{}
			}
			return p.username.Layout(gtx, th, "Username")
		}),
		layout.Rigid(func(gtx C) D {
			if !p.cbBasicAuth.Value {
				p.password.Clear()
				return layout.Dimensions{}
			}

			if p.wgPassword.Clicked(gtx) {
				p.passwordVisible = !p.passwordVisible
			}

			if p.passwordVisible {
				p.password.Suffix = func(gtx layout.Context) layout.Dimensions {
					return p.wgPassword.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return icons.IconVisibility.Layout(gtx, color.NRGBA(colornames.Grey500))
					})
				}
				p.password.Mask = 0
			} else {
				p.password.Suffix = func(gtx layout.Context) layout.Dimensions {
					return p.wgPassword.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return icons.IconVisibilityOff.Layout(gtx, color.NRGBA(colornames.Grey500))
					})
				}
				p.password.Mask = '*'
			}

			return p.password.Layout(gtx, th, "Password")
		}),
	)
}
