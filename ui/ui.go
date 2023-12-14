package ui

import (
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/go-gost/gost-plus/ui/page"
)

type C = layout.Context
type D = layout.Dimensions

type UI struct {
	th     *material.Theme
	router *page.Router
}

func NewUI() *UI {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	// th.Bg = color.NRGBA(colornames.Brown800)
	// th.Fg = color.NRGBA(colornames.Grey50)

	ui := &UI{
		th:     th,
		router: page.NewRouter(),
	}

	return ui
}

func (ui *UI) Layout(gtx C) D {
	return ui.router.Layout(gtx, ui.th)
}
