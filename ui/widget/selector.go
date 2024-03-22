package widget

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/go-gost/gost.plus/ui/i18n"
	"github.com/go-gost/gost.plus/ui/icons"
)

type SelectorItem struct {
	Name  i18n.Key
	Value string
}

type Selector struct {
	Title     i18n.Key
	items     []SelectorItem
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
				layout.Rigid(material.Body1(th, p.Title.Value()).Layout),
				layout.Rigid(layout.Spacer{Width: 8}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						var names []string
						for _, item := range p.items {
							if item.Name != "" {
								names = append(names, item.Name.Value())
							}
						}
						return material.Body2(th, strings.Join(names, ",")).Layout(gtx)
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

func (p *Selector) Select(items ...SelectorItem) {
	for _, item := range items {
		if item.Name == "" {
			continue
		}
		p.items = append(p.items, item)
	}
}

func (p *Selector) Unselect(name string) {
	for i := range p.items {
		if p.items[i].Name.Value() == name {
			p.items = append(p.items[:i], p.items[i+1:]...)
			return
		}
	}
}

func (p *Selector) Any(names ...string) bool {
	for _, name := range names {
		if p.contains(name) {
			return true
		}
	}
	return false
}

func (p *Selector) Item() SelectorItem {
	if len(p.items) == 0 {
		return SelectorItem{}
	}
	return p.items[0]
}

func (p *Selector) Items() []SelectorItem {
	return p.items
}

func (p *Selector) contains(name string) bool {
	if name == "" && len(p.items) == 0 {
		return true
	}

	for i := range p.items {
		if p.items[i].Name.Value() == name {
			return true
		}
	}
	return false
}

func (p *Selector) Clear() {
	p.items = nil
}
