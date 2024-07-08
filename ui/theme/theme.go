package theme

import (
	"image/color"
	"sync"

	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/colornames"
)

const (
	System string = "system"
	Light  string = "light"
	Dark   string = "dark"
)

type Palette struct {
	Material            material.Palette
	ContentSurfaceBg    color.NRGBA
	ListBg              color.NRGBA
	NavButtonBg         color.NRGBA
	NavButtonContrastBg color.NRGBA
	ItemBg              color.NRGBA
	NotificationBg      color.NRGBA
}

type Theme struct {
	Name string
	Palette
}

var (
	light = Theme{
		Name: Light,
		Palette: Palette{
			Material: material.Palette{
				Fg: color.NRGBA(colornames.Black),
				Bg: color.NRGBA(colornames.White),
				ContrastBg: color.NRGBA{
					R: 0x3f,
					G: 0x51,
					B: 0xb5,
					A: 0xff,
				},
				ContrastFg: color.NRGBA(colornames.White),
			},
			ContentSurfaceBg:    color.NRGBA(colornames.Grey50),
			ListBg:              color.NRGBA(colornames.BlueGrey50),
			NavButtonBg:         color.NRGBA(colornames.BlueGrey50),
			NavButtonContrastBg: color.NRGBA(colornames.BlueGrey100),
			ItemBg:              color.NRGBA(colornames.Grey300),
			NotificationBg:      color.NRGBA(colornames.Grey200),
		},
	}

	dark = Theme{
		Name: Dark,
		Palette: Palette{
			Material: material.Palette{
				Fg:         color.NRGBA(colornames.White),
				Bg:         color.NRGBA(colornames.Grey900),
				ContrastBg: color.NRGBA(colornames.Green500),
				ContrastFg: color.NRGBA(colornames.White),
			},
			ContentSurfaceBg:    color.NRGBA(colornames.Grey800),
			ListBg:              color.NRGBA(colornames.Grey700),
			NavButtonBg:         color.NRGBA(colornames.Grey800),
			NavButtonContrastBg: color.NRGBA(colornames.Grey600),
			ItemBg:              color.NRGBA(colornames.Grey600),
			NotificationBg:      color.NRGBA(colornames.Grey700),
		},
	}
)

var (
	theme Theme = light
	mux   sync.RWMutex
)

func Current() Theme {
	mux.RLock()
	defer mux.RUnlock()

	return theme
}

func UseLight() {
	mux.Lock()
	defer mux.Unlock()

	theme = light
}

func UseDark() {
	mux.Lock()
	defer mux.Unlock()

	theme = dark
}
