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
		Fit: widget.ScaleDown,
	}
}

var (
	IconApp *widget.Image

	IconHome                 = mustIcon(icons.ActionHome)
	IconFavorite             = mustIcon(icons.ActionFavorite)
	IconAdd                  = mustIcon(icons.ContentAdd)
	IconRemove               = mustIcon(icons.ContentRemove)
	IconSettings             = mustIcon(icons.ActionSettings)
	IconDone                 = mustIcon(icons.ActionDone)
	IconRadioButtonUnchecked = mustIcon(icons.ToggleRadioButtonUnchecked)
	IconRadioButtonChecked   = mustIcon(icons.ToggleRadioButtonChecked)
	IconForward              = mustIcon(icons.NavigationChevronRight)
	IconEdit                 = mustIcon(icons.EditorModeEdit)
	IconDelete               = mustIcon(icons.ActionDelete)
	IconDeleteForever        = mustIcon(icons.ActionDeleteForever)
	IconStart                = mustIcon(icons.AVPlayArrow)
	IconStop                 = mustIcon(icons.AVStop)
	IconBack                 = mustIcon(icons.NavigationArrowBack)
	IconClose                = mustIcon(icons.ContentClear)
	IconCopy                 = mustIcon(icons.ContentContentCopy)
	IconVisibility           = mustIcon(icons.ActionVisibility)
	IconVisibilityOff        = mustIcon(icons.ActionVisibilityOff)
	IconNavArrowForward      = mustIcon(icons.NavigationArrowForward)
	IconCircle               = mustIcon(icons.ImageLens)
	IconNavRight             = mustIcon(icons.NavigationChevronRight)
	IconNavExpandLess        = mustIcon(icons.NavigationExpandLess)
	IconNavExpandMore        = mustIcon(icons.NavigationExpandMore)
	IconActionCode           = mustIcon(icons.ActionCode)
	IconActionUpdate         = mustIcon(icons.ActionUpdate)
	IconActionHourGlassEmpty = mustIcon(icons.ActionHourglassEmpty)
	IconInfo                 = mustIcon(icons.ActionInfo)
)

func mustIcon(data []byte) *widget.Icon {
	icon, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return icon
}
