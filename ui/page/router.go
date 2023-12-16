package page

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type C = layout.Context
type D = layout.Dimensions

type Route struct {
	Path string
	ID   string
}

type Router struct {
	modal *component.ModalLayer
	bar   *component.AppBar

	pages   map[string]Page
	current Route
}

func NewRouter() *Router {
	modal := component.NewModal()
	bar := component.NewAppBar(modal)

	r := &Router{
		modal: modal,
		bar:   bar,
		pages: make(map[string]Page),
	}

	r.Register(PageTunnel, NewTunnelPage(r))
	r.Register(PageMenu, NewMenuPage(r))
	r.Register(PageNewFile, NewFileAddPage(r))
	r.Register(PageEditFile, NewFileEditPage(r))
	r.Register(PageNewHTTP, NewHTTPAddPage(r))
	r.Register(PageEditHTTP, NewHTTPEditPage(r))
	r.Register(PageNewTCP, NewTCPAddPage(r))
	r.Register(PageEditTCP, NewTCPEditPage(r))
	r.Register(PageNewUDP, NewUDPAddPage(r))
	r.Register(PageEditUDP, NewUDPEditPage(r))

	r.Register(PageEntryPoint, NewEntryPointPage(r))
	r.Register(PageEditTCPEntryPoint, NewTCPEntryPointEditPage(r))
	r.Register(PageNewTCPEntryPoint, NewTCPEntryPointAddPage(r))

	r.Register(PageAbout, NewAboutPage(r))

	r.SwitchTo(Route{Path: PageTunnel})

	return r
}

func (r *Router) Register(path string, page Page) {
	if page != nil {
		r.pages[path] = page
	}
}

func (r *Router) SwitchTo(route Route) {
	p := r.pages[route.Path]
	if p == nil {
		return
	}

	p.Init(IDPageOption(route.ID))

	r.current = route
}

func (r *Router) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	for _, event := range r.bar.Events(gtx) {
		switch event := event.(type) {
		case component.AppBarNavigationClicked:
			// log.Printf("navigation clicked: %+v", event)
			if r.current.Path == PageTunnel {
				r.SwitchTo(Route{Path: PageEntryPoint})
			} else {
				r.SwitchTo(Route{Path: PageTunnel})
			}
		case component.AppBarContextMenuDismissed:
			// log.Printf("Context menu dismissed: %+v", event)
			r.SwitchTo(Route{Path: PageTunnel})
		case component.AppBarOverflowActionClicked:
			if event.Tag == OverflowActionAbout {
				r.SwitchTo(Route{Path: PageAbout})
			}
		}
	}

	defer r.modal.Layout(gtx, th)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return r.bar.Layout(gtx, th, "Menu", "Actions")
		}),
		layout.Flexed(1, func(gtx C) D {
			return r.pages[r.current.Path].Layout(gtx, th)
		}),
	)
}
