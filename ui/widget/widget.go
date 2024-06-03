package widget

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions
type T = material.Theme

type Widget interface {
	Layout(gtx C, th *T) D
}
