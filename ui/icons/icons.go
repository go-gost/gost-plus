package icons

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"gioui.org/op/paint"
	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

//go:embed icon.png
var iconAppData []byte

func init() {
	img, _, err := image.Decode(bytes.NewReader(iconAppData))
	if err != nil {
		log.Println("err")
		return
	}

	IconApp = &widget.Image{
		Src: paint.NewImageOp(img),
		Fit: widget.Unscaled,
	}
}

var (
	IconApp *widget.Image

	IconHome        = mustIcon(icons.ActionHome)
	IconFavorite    = mustIcon(icons.ActionFavorite)
	IconAdd         = mustIcon(icons.ContentAdd)
	IconSettings    = mustIcon(icons.ActionSettings)
	IconDone        = mustIcon(icons.ActionDone)
	IconTunnelState = mustIcon(icons.ToggleRadioButtonChecked)
	IconForward     = mustIcon(icons.NavigationChevronRight)
	IconEdit        = mustIcon(icons.EditorModeEdit)
	IconDelete      = mustIcon(icons.ActionDelete)
	IconStart       = mustIcon(icons.AVPlayArrow)
	IconStop        = mustIcon(icons.AVStop)
	IconBack        = mustIcon(icons.NavigationArrowBack)
	IconClose       = mustIcon(icons.ContentClear)
)

func mustIcon(data []byte) *widget.Icon {
	icon, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return icon
}
