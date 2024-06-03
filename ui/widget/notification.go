package widget

import (
	"image/color"
	"sync"
	"sync/atomic"
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/go-gost/gost.plus/ui/icons"
	"github.com/go-gost/gost.plus/ui/theme"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

const (
	Success string = "success"
	Info    string = "info"
	Warn    string = "warn"
	Error   string = "error"
)

type Message struct {
	Type    string
	Content string
}

type Notification struct {
	show     atomic.Bool
	messages chan Message
	current  Message
	duration time.Duration
	callback func()
	mu       sync.RWMutex
}

func NewNotification(d time.Duration, callback func()) *Notification {
	p := &Notification{
		messages: make(chan Message, 16),
		duration: d,
		callback: callback,
	}
	go p.run()
	return p
}

func (p *Notification) Layout(gtx C, th *T) D {
	if !p.show.Load() {
		return D{}
	}

	p.mu.RLock()
	message := p.current
	p.mu.RUnlock()

	return component.SurfaceStyle{
		Theme: th,
		ShadowStyle: component.ShadowStyle{
			CornerRadius: 8,
		},
		Fill: theme.Current().NotificationBg,
	}.Layout(gtx, func(gtx C) D {
		return layout.Inset{
			Top:    8,
			Bottom: 8,
			Left:   8,
			Right:  8,
		}.Layout(gtx, func(gtx C) D {
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {

					switch message.Type {
					case Success:
						return icons.IconInfo.Layout(gtx, color.NRGBA(colornames.Green500))
					case Warn:
						return icons.IconInfo.Layout(gtx, color.NRGBA(colornames.Orange500))
					case Error:
						return icons.IconAlert.Layout(gtx, color.NRGBA(colornames.Red500))
					case Info:
						fallthrough
					default:
						return icons.IconInfo.Layout(gtx, color.NRGBA(colornames.Blue500))
					}
				}),
				layout.Rigid(layout.Spacer{Width: 8}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Body2(th, message.Content).Layout(gtx)
				}),
			)
		})
	})
}

func (p *Notification) Show(message Message) {
	select {
	case p.messages <- message:
	default:
	}
}

func (p *Notification) run() {
	for m := range p.messages {
		p.mu.Lock()
		p.current = m
		p.mu.Unlock()

		p.show.Store(true)
		if p.callback != nil {
			p.callback()
		}

		<-time.After(p.duration)

		p.show.Store(false)
		if p.callback != nil {
			p.callback()
		}
	}
}
