package list

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

type Filter struct {
	Favorite bool
	Name     string
}

type List interface {
	Layout(gtx C, th *material.Theme) D
	Filter(f Filter)
}
