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
	"github.com/go-gost/gost-plus/tunnel/entrypoint"
	"github.com/go-gost/gost-plus/ui/icons"
	"github.com/google/uuid"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type tcpEntryPointAddPage struct {
	router *Router

	list   layout.List
	wgDone widget.Clickable

	name     component.TextField
	tunnelID component.TextField
	addr     component.TextField
}

func NewTCPEntryPointAddPage(r *Router) Page {
	return &tcpEntryPointAddPage{
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
	}
}

func (p *tcpEntryPointAddPage) Init(opts ...PageOption) {
	p.name.SetText("")
	p.tunnelID.SetText("")
	p.addr.SetText("")

	p.router.bar.SetActions(
		[]component.AppBarAction{
			{
				OverflowAction: component.OverflowAction{
					Name: "Create",
					Tag:  &p.wgDone,
				},
				Layout: func(gtx C, bg, fg color.NRGBA) D {
					if p.wgDone.Clicked(gtx) {
						defer p.router.SwitchTo(Route{Path: PageEntryPoint})
						p.createEntryPoint()
						entrypoint.SaveConfig()
					}
					return component.SimpleIconButton(bg, fg, &p.wgDone, icons.IconDone).Layout(gtx)
				},
			},
		}, nil)

	p.router.bar.Title = "TCP"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *tcpEntryPointAddPage) Layout(gtx C, th *material.Theme) D {
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

func (p *tcpEntryPointAddPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(material.Body1(th, "Create an entrypoint to connect to the specified TCP tunnel").Layout),
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
	)
}

func (p *tcpEntryPointAddPage) createEntryPoint() error {
	tun := entrypoint.NewTCPEntryPoint(
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.IDOption(strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))),
		tunnel.EndpointOption(strings.TrimSpace(p.addr.Text())),
	)

	entrypoint.Add(tun)

	if err := tun.Run(); err != nil {
		tun.Close()
		return err
	}

	return nil
}

type tcpEntryPointEditPage struct {
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
}

func NewTCPEntryPointEditPage(r *Router) Page {
	return &tcpEntryPointEditPage{
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
	}
}

func (p *tcpEntryPointEditPage) Init(opts ...PageOption) {
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
						s = p.createEntryPoint()
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
	p.router.bar.Title = "TCP"
	p.router.bar.NavigationIcon = icons.IconClose
}

func (p *tcpEntryPointEditPage) createEntryPoint() entrypoint.EntryPoint {
	s := entrypoint.NewTCPEntryPoint(
		tunnel.IDOption(p.id),
		tunnel.NameOption(strings.TrimSpace(p.name.Text())),
		tunnel.IDOption(strings.ToLower(strings.TrimSpace(p.tunnelID.Text()))),
		tunnel.EndpointOption(strings.TrimSpace(p.addr.Text())),
	)

	if err := s.Run(); err != nil {
		log.Println(err)
	}
	entrypoint.Set(s)

	return s
}

func (p *tcpEntryPointEditPage) Layout(gtx C, th *material.Theme) D {
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

func (p *tcpEntryPointEditPage) layout(gtx C, th *material.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
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
	)
}
