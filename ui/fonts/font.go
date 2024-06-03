package fonts

import (
	_ "embed"
	"fmt"

	"gioui.org/font"
	"gioui.org/font/opentype"
)

var (

	//go:embed NotoSans-Regular.ttf
	notoSansRegular []byte
	//go:embed NotoSans-SemiBold.ttf
	notoSansSemiBold []byte

	//go:embed NotoSansSC-Regular.ttf
	notoSansSCRegular []byte
	//go:embed NotoSansSC-SemiBold.ttf
	notoSansSCSemiBold []byte
)

var (
	collection []font.FontFace
)

func init() {
	register(notoSansRegular)
	register(notoSansSemiBold)
	register(notoSansSCRegular)
	register(notoSansSCSemiBold)
}

func Collection() []font.FontFace {
	return collection
}

func register(ttf []byte) {
	faces, err := opentype.ParseCollection(ttf)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	collection = append(collection, faces...)
}
