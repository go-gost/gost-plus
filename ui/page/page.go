package page

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost-plus/tunnel"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

const (
	PageHome     = "/"
	PageMenu     = "/menu"
	PageNewFile  = "/tunnel/file/create"
	PageNewHTTP  = "/tunnel/http/create"
	PageNewTCP   = "/tunnel/tcp/create"
	PageNewUDP   = "/tunnel/udp/create"
	PageEditFile = "/tunnel/file/edit"
	PageEditHTTP = "/tunnel/http/edit"
	PageEditTCP  = "/tunnel/tcp/edit"
	PageEditUDP  = "/tunnel/udp/edit"

	PageAbout = "/about"
)

type OverflowAction string

const (
	OverflowActionAbout OverflowAction = "about"
)

type PageOptions struct {
	ID string
}

type PageOption func(*PageOptions)

func IDPageOption(id string) PageOption {
	return func(opts *PageOptions) {
		opts.ID = id
	}
}

type Page interface {
	Init(opts ...PageOption)
	Layout(gtx layout.Context, th *material.Theme) layout.Dimensions
}

func layoutHeader(gtx C, th *material.Theme, tun tunnel.Tunnel, wgID, wgEntrypoint *widget.Clickable) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if tun == nil {
				return D{}
			}

			copied := false
			if wgID.Clicked(gtx) {
				copied = true
				clipboard.WriteOp{
					Text: tun.ID(),
				}.Add(gtx.Ops)
			}

			return wgID.Layout(gtx, func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						label := material.Body1(th, tun.ID())
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 10}.Layout),
					layout.Rigid(func(gtx C) D {
						label := material.Body1(th, "Copy")
						label.Color = color.NRGBA(colornames.Blue500)
						if copied {
							label = material.Body1(th, "Copied")
							label.Color = color.NRGBA(colornames.Green500)
						}
						return label.Layout(gtx)
					}),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			if tun == nil {
				return D{}
			}

			copied := false
			if wgEntrypoint.Clicked(gtx) {
				copied = true
				clipboard.WriteOp{
					Text: tun.Entrypoint(),
				}.Add(gtx.Ops)
			}

			return wgEntrypoint.Layout(gtx, func(gtx C) D {
				return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, func(gtx C) D {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							label := material.Body1(th, tun.Entrypoint())
							return label.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: 10}.Layout),
						layout.Rigid(func(gtx C) D {
							label := material.Body1(th, "Copy")
							label.Color = color.NRGBA(colornames.Blue500)
							if copied {
								label = material.Body1(th, "Copied")
								label.Color = color.NRGBA(colornames.Green500)
							}
							return label.Layout(gtx)
						}),
					)
				})
			})
		}),
	)
}
