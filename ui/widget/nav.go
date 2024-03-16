package widget

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/ui/theme"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

type Nav struct {
	list    layout.List
	btns    []*NavButton
	current int
}

func NewNav(btns ...*NavButton) *Nav {
	return &Nav{
		btns: btns,
	}
}

func (p *Nav) Current() int {
	return p.current
}

func (p *Nav) SetCurrent(n int) {
	p.current = n
}

func (p *Nav) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	for i, btn := range p.btns {
		if btn.btn.Clicked(gtx) {
			p.current = i
			break
		}
	}

	return p.list.Layout(gtx, len(p.btns), func(gtx layout.Context, index int) layout.Dimensions {
		btn := p.btns[index]

		if p.current == index {
			btn.background = theme.Current().NavButtonContrastBg
			btn.borderWidth = 0
		} else {
			btn.background = theme.Current().NavButtonBg
			btn.borderWidth = 1
		}

		return layout.Inset{
			Top:    8,
			Bottom: 8,
			Left:   12,
			Right:  12,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return btn.Layout(gtx, th)
		})
	})
}

type NavButton struct {
	btn          widget.Clickable
	cornerRadius unit.Dp
	borderWidth  unit.Dp
	borderColor  color.NRGBA
	background   color.NRGBA
	text         string
}

func NewNavButton(text string) *NavButton {
	return &NavButton{
		cornerRadius: 18,
		borderWidth:  1,
		borderColor:  color.NRGBA(colornames.Grey200),
		text:         text,
	}
}

func (btn *NavButton) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return material.ButtonLayoutStyle{
		Background:   btn.background,
		CornerRadius: btn.cornerRadius,
		Button:       &btn.btn,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return widget.Border{
			Color:        btn.borderColor,
			Width:        btn.borderWidth,
			CornerRadius: btn.cornerRadius,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    8,
				Bottom: 8,
				Left:   20,
				Right:  20,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(th, btn.text)
				return label.Layout(gtx)
			})
		})
	})
}
