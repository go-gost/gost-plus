package widget

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Switcher struct {
	clickable widget.Clickable
	b         widget.Bool
	Title     string
}

func (p *Switcher) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.clickable.Clicked(gtx) {
		p.b.Value = !p.b.Value
	}

	return material.Clickable(gtx, &p.clickable, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    8,
			Bottom: 8,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(1, material.Body1(th, p.Title).Layout),
				layout.Rigid(material.Switch(th, &p.b, "").Layout),
			)
		})
	})
}

func (p *Switcher) Value() bool {
	return p.b.Value
}

func (p *Switcher) SetValue(b bool) {
	p.b.Value = b
}
