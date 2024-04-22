package list

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/page"
	"github.com/go-gost/gost.plus/ui/theme"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type tunnelState struct {
	btn widget.Clickable
}

type tunnelList struct {
	router *page.Router
	list   layout.List
	states []tunnelState
	filter Filter
}

func Tunnel(r *page.Router) List {
	return &tunnelList{
		router: r,
		list: layout.List{
			Axis: layout.Vertical,
		},
		states: make([]tunnelState, 16),
	}
}

func (l *tunnelList) Filter(f Filter) {
	l.filter = f
}

func (l *tunnelList) Layout(gtx C, th *material.Theme) D {
	tn := tunnel.Count()
	if tn > len(l.states) {
		states := l.states
		l.states = make([]tunnelState, tn)
		copy(l.states, states)
	}

	return l.list.Layout(gtx, tn, func(gtx C, index int) D {
		t := tunnel.GetIndex(index)
		if t == nil {
			return D{}
		}

		if l.filter.Favorite && !t.IsFavorite() {
			return D{}
		}

		if l.states[index].btn.Clicked(gtx) {
			var path page.PagePath
			switch t.Type() {
			case tunnel.FileTunnel:
				path = page.PageTunnelFile
			case tunnel.HTTPTunnel:
				path = page.PageTunnelHTTP
			case tunnel.TCPTunnel:
				path = page.PageTunnelTCP
			case tunnel.UDPTunnel:
				path = page.PageTunnelUDP
			}
			l.router.Goto(page.Route{
				Path: path,
				ID:   t.ID(),
			})
		}

		stats := t.Stats()

		return layout.Inset{
			Top:    8,
			Bottom: 8,
			Left:   8,
			Right:  8,
		}.Layout(gtx, func(gtx C) D {
			return material.ButtonLayoutStyle{
				Background:   theme.Current().ListBg,
				CornerRadius: 12,
				Button:       &l.states[index].btn,
			}.Layout(gtx, func(gtx C) D {
				return layout.UniformInset(16).Layout(gtx, func(gtx C) D {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Alignment: layout.Middle,
								Spacing:   layout.SpaceBetween,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									label := material.Body1(th, t.Name())
									label.Font.Weight = font.SemiBold
									return label.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: 4}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = gtx.Dp(12)

									c := colornames.GreenA700
									if t.Err() != nil {
										c = colornames.Red600
									}
									if t.IsClosed() {
										c = colornames.Grey600
									}
									return icons.IconCircle.Layout(gtx, color.NRGBA(c))
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: 4}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Alignment: layout.Middle,
								Spacing:   layout.SpaceBetween,
							}.Layout(gtx,
								layout.Flexed(1, material.Body2(th, fmt.Sprintf("%s: %s", i18n.Type.Value(), strings.ToUpper(t.Type()))).Layout),
								layout.Rigid(layout.Spacer{Width: 4}.Layout),
								layout.Rigid(func(gtx C) D {
									if createdAt := t.Options().CreatedAt; !createdAt.IsZero() {
										v, unit := formatDuration(time.Since(createdAt))
										return material.Body2(th, fmt.Sprintf("%d%s", v, unit)).Layout(gtx)
									}
									return material.Body2(th, "N/A").Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: 4}.Layout),
						layout.Rigid(material.Body2(th, fmt.Sprintf("%s: %s", i18n.Endpoint.Value(), t.Endpoint())).Layout),
						layout.Rigid(layout.Spacer{Height: 4}.Layout),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{
								Alignment: layout.Middle,
								Spacing:   layout.SpaceBetween,
							}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return icons.IconActionCode.Layout(gtx, th.Fg)
								}),
								layout.Rigid(layout.Spacer{Width: 4}.Layout),
								layout.Flexed(1, func(gtx C) D {
									current, unitCurrent := format(int64(stats.CurrentConns), 1000)
									current = float64(int64(current*10)) / 10

									total, unitTotal := format(int64(stats.TotalConns), 1000)
									total = float64(int64(total*10)) / 10
									return material.Body2(th, fmt.Sprintf("%s%s / %s%s",
										strconv.FormatFloat(current, 'f', -1, 64), strings.ToLower(unitCurrent),
										strconv.FormatFloat(total, 'f', -1, 64), strings.ToLower(unitTotal))).Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									rate := stats.RequestRate
									rate = float64(int64(rate*100)) / 100
									return material.Body2(th, fmt.Sprintf("%s R/s", strconv.FormatFloat(rate, 'f', -1, 64))).Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: 4}.Layout),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{
								Alignment: layout.Middle,
								Spacing:   layout.SpaceBetween,
							}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return icons.IconNavExpandLess.Layout(gtx, th.Fg)
								}),
								layout.Rigid(layout.Spacer{Width: 4}.Layout),
								layout.Flexed(1, func(gtx C) D {
									v, unit := format(int64(stats.OutputBytes), 1024)
									v = float64(int64(v*100)) / 100
									return material.Body2(th, fmt.Sprintf("%s %sB", strconv.FormatFloat(v, 'f', -1, 64), unit)).Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									v, unit := format(int64(stats.OutputRateBytes), 1024)
									v = float64(int64(v*100)) / 100
									return material.Body2(th, fmt.Sprintf("%s %sB/s", strconv.FormatFloat(v, 'f', -1, 64), unit)).Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: 4}.Layout),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{
								Alignment: layout.Middle,
								Spacing:   layout.SpaceBetween,
							}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return icons.IconNavExpandMore.Layout(gtx, th.Fg)
								}),
								layout.Rigid(layout.Spacer{Width: 4}.Layout),
								layout.Flexed(1, func(gtx C) D {
									v, unit := format(int64(stats.InputBytes), 1024)
									v = float64(int64(v*100)) / 100
									return material.Body2(th, fmt.Sprintf("%s %sB", strconv.FormatFloat(v, 'f', -1, 64), unit)).Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									v, unit := format(int64(stats.InputRateBytes), 1024)
									v = float64(int64(v*100)) / 100
									return material.Body2(th, fmt.Sprintf("%s %sB/s", strconv.FormatFloat(v, 'f', -1, 64), unit)).Layout(gtx)
								}),
							)
						}),
					)
				})
			})
		})
	})
}

var (
	units = []string{"", "K", "M", "G", "T", "P", "E"}
)

func format(n int64, scale int64) (v float64, unit string) {
	var remain float64
	for i := range units {
		unit = units[i]

		r := n % scale
		if n = n / scale; n == 0 {
			v = float64(r) + remain/math.Pow(float64(scale), float64(i))
			return
		}
		remain += float64(r) * math.Pow(float64(scale), float64(i))
	}
	return
}

var (
	dunits = []string{"s", "m", "h"}
)

func formatDuration(d time.Duration) (v int64, unit string) {
	if d.Hours() >= 24 {
		v = int64(d.Hours() / 24)
		unit = "d"
		return
	}

	var scale int64 = 60
	n := int64(d.Seconds())
	for i := range dunits {
		v = n % scale
		unit = dunits[i]

		if n = n / scale; n == 0 {
			return
		}
	}
	return
}
