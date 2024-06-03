package page

import (
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/ui/theme"
	ui_widget "github.com/go-gost/gost.plus/ui/widget"
)

const (
	MaxWidth = 800
)

type Route struct {
	Path PagePath
	ID   string
}

type Router struct {
	w       *app.Window
	pages   map[PagePath]Page
	stack   routeStack
	current Route
	*material.Theme
	modal        *component.ModalLayer
	notification *ui_widget.Notification
	events       chan Event
}

func NewRouter(w *app.Window, th *T) *Router {
	r := &Router{
		w:     w,
		pages: make(map[PagePath]Page),
		Theme: th,
		modal: component.NewModal(),
		notification: ui_widget.NewNotification(3*time.Second, func() {
			w.Invalidate()
		}),
		events: make(chan Event, 16),
	}

	return r
}

func (r *Router) Register(path PagePath, page Page) {
	r.pages[path] = page
}

func (r *Router) Goto(route Route) {
	page := r.pages[route.Path]
	if page == nil {
		return
	}

	r.current = route
	r.stack.Push(route)

	page.Init(WithPageID(route.ID))
	logger.Default().WithFields(map[string]any{
		"kind":     "router",
		"route.id": route.ID,
	}).Debugf("go to %s", route.Path)
}

func (r *Router) Back() {
	r.stack.Pop()
	route := r.stack.Peek()

	page := r.pages[route.Path]
	if page == nil {
		return
	}
	r.current = route

	logger.Default().WithFields(map[string]any{
		"kind":     "router",
		"route.id": route.ID,
	}).Debugf("back to %s", route.Path)
}

func (r *Router) Layout(gtx C) D {
	if r.stack.Peek().Path != PageHome {
		event.Op(gtx.Ops, r.w)
		for {
			ev, ok := gtx.Event(
				key.Filter{Name: key.NameBack},
				key.Filter{Name: key.NameEscape},
			)
			if !ok {
				break
			}
			switch ev := ev.(type) {
			case key.Event:
				if ev.State == key.Press {
					r.Back()
				}
			}
		}
	}

	r.Theme.Palette = theme.Current().Material

	defer r.modal.Layout(gtx, r.Theme)

	return layout.Background{}.Layout(gtx,
		func(gtx C) D {
			defer clip.Rect{
				Max: gtx.Constraints.Max,
			}.Op().Push(gtx.Ops).Pop()

			paint.ColorOp{
				Color: r.Theme.Bg,
			}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)

			return layout.Dimensions{
				Size: gtx.Constraints.Max,
			}
		},
		func(gtx C) D {
			page := r.pages[r.current.Path]
			if page == nil {
				page = r.pages[PageHome]
			}

			inset := layout.Inset{}
			width := unit.Dp(MaxWidth)
			if x := gtx.Metric.PxToDp(gtx.Constraints.Max.X); x > width {
				inset.Left = (x - width) / 2
				inset.Right = inset.Left
			}

			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Stack{
					Alignment: layout.N,
				}.Layout(gtx,
					layout.Expanded(page.Layout),
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Top:    16,
							Bottom: 16,
							Right:  gtx.Metric.PxToDp(gtx.Constraints.Max.X) / 5,
							Left:   gtx.Metric.PxToDp(gtx.Constraints.Max.X) / 5,
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return r.notification.Layout(gtx, r.Theme)
						})
					}),
				)
			})
		},
	)
}

func (r *Router) ShowModal(gtx layout.Context, w func(gtx C, th *material.Theme) D) {
	r.modal.Widget = func(gtx C, th *material.Theme, anim *component.VisibilityAnimation) D {
		if gtx.Constraints.Max.X > gtx.Dp(MaxWidth) {
			gtx.Constraints.Max.X = gtx.Dp(MaxWidth)
		}
		gtx.Constraints.Max.X = gtx.Constraints.Max.X * 3 / 4

		var clk widget.Clickable
		return clk.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return w(gtx, th)
		})
	}
	r.modal.Appear(gtx.Now)
}

func (r *Router) HideModal(gtx C) {
	r.modal.Disappear(gtx.Now)
}

func (r *Router) Notify(message ui_widget.Message) {
	r.notification.Show(message)
}

func (r *Router) Emit(event Event) {
	select {
	case r.events <- event:
	default:
	}
}

func (r *Router) Event() <-chan Event {
	return r.events
}

type routeStack struct {
	routes []Route
}

func (p *routeStack) Push(route Route) {
	p.routes = append(p.routes, route)
}

func (p *routeStack) Pop() (route Route) {
	if len(p.routes) == 0 {
		return
	}

	n := len(p.routes) - 1
	route = p.routes[n]
	p.routes = p.routes[:n]

	return
}

func (p *routeStack) Peek() (route Route) {
	if len(p.routes) == 0 {
		return
	}

	return p.routes[len(p.routes)-1]
}
