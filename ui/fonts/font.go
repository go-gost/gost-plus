package fonts

import (
	_ "embed"
	"fmt"

	"gioui.org/font"
	"gioui.org/font/opentype"
)

//go:embed NotoSansSC-Regular.ttf
var fontNotoRegular []byte

//go:embed NotoSansSC-Bold.ttf
var fontNotoBold []byte

var (
	collection []font.FontFace
)

func init() {
	register(fontNotoRegular)
	register(fontNotoBold)
}

func Collection() []font.FontFace {
	return collection
}

func register(ttf []byte) {
	faces, err := opentype.ParseCollection(ttf)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	collection = append(collection, faces[0])
}
