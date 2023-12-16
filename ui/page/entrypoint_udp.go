package page

import (
	"fmt"
	"image/color"
	"log"
	"net"
	"strconv"
	"strings"
	"unicode"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost-plus/tunnel"
	"github.com/go-gost/gost-plus/tunnel/entrypoint"
	"github.com/go-gost/gost-plus/ui/icons"
	"github.com/google/uuid"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type udpEntryPointAddPage struct {
	router *Router

	list   layout.List
	wgDone widget.Clickable

	name     component.TextField
	tunnelID component.TextField
	addr     component.TextField

	bKeepalive widget.Bool
	ttl        component.TextField
}

func NewUDPEntryPointAddPage(r *Router) Page {
	return &udpEntryPointAddPage{
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
		tunnelID: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		addr: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		ttl: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
	}
}

func (p *udpEntryPointAddPage) Init(opts ...PageOption) {
	p.name.SetText("")
	p.tunnelID.SetText("")
	p.addr.SetText("")
	p.bKeepalive.Value = false
	p.ttl.SetText("")

	p.router.bar.SetActions(
		[]component.AppBarAction{
			{
				OverflowAction: component.OverflowAction{
					Name: "Create",
					Tag:  &p.wgDone,
				},
				Layout: func(gtx C, bg, fg color.NRGBA) D {
					if !p.isValid() {
						gtx = gtx.Disabled()
					}
					if p.wgDone.Clicked(gtx) {
						defer p.router.SwitchTo(Route{Path: PageEntryPoint})
						p.createEntryPoint()
						entrypoint.SaveConfig()
					}
					return component.SimpleIconButton(bg, fg, &p.wgDone, icons.IconDone).Layout(gtx)
				},
			},
		}, nil)

	p.router.bar.Title = "UDP"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *udpEntryPointAddPage) isValid() bool {
	if p.tunnelID.Text() == "" || p.tunnelID.IsErrored() ||
		p.addr.Text() == "" || p.addr.IsErrored() {
		return false
	}
	return true
}

func (p *udpEntryPointAddPage) Layout(gtx C, th *material.Theme) D {
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

func (p *udpEntryPointAddPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(material.Body1(th, "Create an entrypoint to the specified UDP tunnel").Layout),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Entrypoint name").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return p.name.Layout(gtx, th, "Name")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Tunnel ID").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if err := func() error {
				tid := strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))
				if tid == "" {
					return nil
				}
				if _, err := uuid.Parse(tid); err != nil {
					return fmt.Errorf("invalid tunnel ID, should be a valid UUID")
				}
				if ep := entrypoint.Get(tid); ep != nil {
					return fmt.Errorf("the entrypoint for this tunnel exists")
				}
				return nil
			}(); err != nil {
				p.tunnelID.SetError(err.Error())
			} else {
				p.tunnelID.ClearError()
			}

			return p.tunnelID.Layout(gtx, th, "ID")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Entrypoint address").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if err := func() error {
				addr := strings.TrimSpace(p.addr.Text())
				if addr == "" {
					return nil
				}
				if _, err := net.ResolveUDPAddr("udp", addr); err != nil {
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
					layout.Flexed(1, material.Body1(th, "Enable keepalive").Layout),
					layout.Rigid(material.Switch(th, &p.bKeepalive, "keepalive").Layout),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			if !p.bKeepalive.Value {
				p.ttl.Clear()
				return layout.Dimensions{}
			}
			p.ttl.Suffix = func(gtx C) D {
				return material.Label(th, th.TextSize, "s").Layout(gtx)
			}

			if err := func() string {
				for _, r := range p.ttl.Text() {
					if !unicode.IsDigit(r) {
						return "Must contain only digits"
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
	)
}

func (p *udpEntryPointAddPage) createEntryPoint() error {
	ttl, _ := strconv.Atoi(strings.TrimSpace(p.ttl.Text()))
	ep := entrypoint.NewUDPEntryPoint(
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.IDOption(strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))),
		tunnel.EndpointOption(strings.TrimSpace(p.addr.Text())),
		tunnel.KeepaliveOption(p.bKeepalive.Value),
		tunnel.TTLOption(ttl),
	)

	entrypoint.Add(ep)

	if err := ep.Run(); err != nil {
		return err
	}

	return nil
}

type udpEntryPointEditPage struct {
	router *Router

	id string

	list       layout.List
	wgFavorite widget.Clickable
	wgState    widget.Clickable
	wgDelete   widget.Clickable
	wgDone     widget.Clickable

	name     component.TextField
	tunnelID component.TextField
	addr     component.TextField

	bKeepalive widget.Bool
	ttl        component.TextField
}

func NewUDPEntryPointEditPage(r *Router) Page {
	return &udpEntryPointEditPage{
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
		tunnelID: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
				ReadOnly:   true,
			},
		},
		addr: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
		ttl: component.TextField{
			Editor: widget.Editor{
				SingleLine: true,
			},
		},
	}
}

func (p *udpEntryPointEditPage) Init(opts ...PageOption) {
	var options PageOptions
	for _, opt := range opts {
		opt(&options)
	}

	p.id = options.ID

	s := entrypoint.Get(p.id)
	if s != nil {
		sopts := s.Options()
		p.name.SetText(sopts.Name)
		p.tunnelID.SetText(sopts.ID)
		p.addr.SetText(sopts.Endpoint)
		p.bKeepalive.Value = sopts.Keepalive
		p.ttl.SetText(strconv.Itoa(sopts.TTL))
	}

	actions := []component.AppBarAction{
		{
			OverflowAction: component.OverflowAction{
				Name: "Favorite",
				Tag:  &p.wgFavorite,
			},
			Layout: func(gtx C, bg, fg color.NRGBA) D {
				s := entrypoint.Get(p.id)
				if s == nil {
					return D{}
				}

				if p.wgFavorite.Clicked(gtx) {
					s.Favorite(!s.IsFavorite())
					entrypoint.SaveConfig()
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
				s := entrypoint.Get(p.id)
				if p.wgState.Clicked(gtx) && s != nil {
					if s.IsClosed() {
						opts := s.Options()
						s = p.createEntryPoint(
							tunnel.NameOption(opts.Name),
							tunnel.IDOption(opts.ID),
							tunnel.EndpointOption(opts.Endpoint),
							tunnel.KeepaliveOption(opts.Keepalive),
							tunnel.TTLOption(opts.TTL),
						)
					} else {
						s.Close()
					}
					entrypoint.SaveConfig()
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
					entrypoint.Delete(p.id)
					entrypoint.SaveConfig()
					p.router.SwitchTo(Route{Path: PageEntryPoint})
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
				if !p.isValid() {
					gtx = gtx.Disabled()
				}

				if p.wgDone.Clicked(gtx) {
					defer p.router.SwitchTo(Route{Path: PageEntryPoint})

					if s := entrypoint.Get(p.id); s != nil {
						s.Close()
						p.createEntryPoint()
						entrypoint.SaveConfig()
					}
				}
				return component.SimpleIconButton(bg, fg, &p.wgDone, icons.IconDone).Layout(gtx)
			},
		},
	}
	p.router.bar.SetActions(actions, nil)
	p.router.bar.Title = "UDP"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *udpEntryPointEditPage) isValid() bool {
	if p.tunnelID.Text() == "" || p.tunnelID.IsErrored() ||
		p.addr.Text() == "" || p.addr.IsErrored() {
		return false
	}
	return true
}

func (p *udpEntryPointEditPage) createEntryPoint(opts ...tunnel.Option) entrypoint.EntryPoint {
	if opts == nil {
		ttl, _ := strconv.Atoi(strings.TrimSpace(p.ttl.Text()))
		opts = []tunnel.Option{
			tunnel.NameOption(strings.TrimSpace(p.name.Text())),
			tunnel.IDOption(p.id),
			tunnel.EndpointOption(strings.TrimSpace(p.addr.Text())),
			tunnel.KeepaliveOption(p.bKeepalive.Value),
			tunnel.TTLOption(ttl),
		}
	}
	ep := entrypoint.NewUDPEntryPoint(opts...)

	entrypoint.Set(ep)

	if err := ep.Run(); err != nil {
		log.Println(err)
	}

	return ep
}

func (p *udpEntryPointEditPage) Layout(gtx C, th *material.Theme) D {
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

func (p *udpEntryPointEditPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Entrypoint name").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return p.name.Layout(gtx, th, "Name")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Tunnel ID").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if err := func() error {
				tid := strings.TrimSpace(p.tunnelID.Text())
				if tid == "" {
					return nil
				}
				if _, err := uuid.Parse(tid); err != nil {
					return fmt.Errorf("invalid tunnel ID, should be a valid UUID")
				}
				return nil
			}(); err != nil {
				p.tunnelID.SetError(err.Error())
			} else {
				p.tunnelID.ClearError()
			}

			return p.tunnelID.Layout(gtx, th, "ID")
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.Body1(th, "Entrypoint address").Layout(gtx)
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
					layout.Flexed(1, material.Body1(th, "Enable keepalive").Layout),
					layout.Rigid(material.Switch(th, &p.bKeepalive, "keepalive").Layout),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			if !p.bKeepalive.Value {
				p.ttl.Clear()
				return layout.Dimensions{}
			}
			p.ttl.Suffix = func(gtx C) D {
				return material.Label(th, th.TextSize, "s").Layout(gtx)
			}

			if err := func() string {
				for _, r := range p.ttl.Text() {
					if !unicode.IsDigit(r) {
						return "Must contain only digits"
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
	)
}
