package palette

import (
	"github.com/jkl1337/go-chromath"
	// "github.com/jkl1337/go-chromath/deltae"
	// "image/color"
)

var (
	RGB2Xyz = chromath.NewRGBTransformer(&chromath.SpaceSRGB, nil, nil, nil, 1.0, nil)
	Lab2Xyz = chromath.NewLabTransformer(&chromath.IlluminantRefD65)
)

func RGB2Lab(rgb chromath.RGB) chromath.Lab {
	xyz := RGB2Xyz.Convert(rgb)
	return Lab2Xyz.Invert(xyz)
}
