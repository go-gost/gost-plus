package widget

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/ui/icons"
)

type Selector struct {
	Title     string
	items     []string
	clickable widget.Clickable
}

func (p *Selector) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return material.Clickable(gtx, &p.clickable, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    8,
			Bottom: 8,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(material.Body1(th, p.Title).Layout),
				layout.Rigid(layout.Spacer{Width: 8}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Body2(th, strings.Join(p.items, ",")).Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return icons.IconNavRight.Layout(gtx, th.Fg)
				}),
			)
		})
	})
}

func (p *Selector) Clicked(gtx layout.Context) bool {
	return p.clickable.Clicked(gtx)
}

func (p *Selector) Select(items ...string) {
	for _, item := range items {
		if item == "" {
			continue
		}
		p.items = append(p.items, item)
	}
}

func (p *Selector) Unselect(item string) {
	for i := range p.items {
		if p.items[i] == item {
			p.items = append(p.items[:i], p.items[i+1:]...)
			return
		}
	}
}

func (p *Selector) Any(items ...string) bool {
	for _, item := range items {
		if p.contains(item) {
			return true
		}
	}
	return false
}

func (p *Selector) Value() string {
	if len(p.items) == 0 {
		return ""
	}
	return p.items[0]
}

func (p *Selector) Values() []string {
	return p.items
}

func (p *Selector) contains(item string) bool {
	if item == "" && len(p.items) == 0 {
		return true
	}

	for i := range p.items {
		if p.items[i] == item {
			return true
		}
	}
	return false
}

func (p *Selector) Clear() {
	p.items = nil
}
