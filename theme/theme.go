package theme

import (
	"fmt"
	"github.com/esimov/colorquant"
	"github.com/jkl1337/go-chromath"
	"github.com/jkl1337/go-chromath/deltae"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sort"
	"strconv"
)

var (
	// for RGB-to-Lab conversion
	targetIlluminant = &chromath.IlluminantRefD50
	rgb2Xyz          = chromath.NewRGBTransformer(
		&chromath.SpaceSRGB,
		&chromath.AdaptationBradford,
		targetIlluminant,
		&chromath.Scaler8bClamping,
		1.0,
		nil,
	)
	lab2Xyz = chromath.NewLabTransformer(targetIlluminant)
	klch    = &deltae.KLChDefault
)

// Palette represents a set of colors and their associated 'roles' (e.g. color0, color1, etc.).
type Palette map[int]ColorVol

// Theme represents a desktop theme.
type Theme map[string]interface{}

// ColorVol represents an RGB color, its Lab equivalent, and the number of pixels it takes up in a given image.
type ColorVol struct {
	RGB   color.Color
	Lab   chromath.Lab
	Count int
}

type byCount []ColorVol

func (cvs byCount) Len() int           { return len(cvs) }
func (cvs byCount) Less(i, j int) bool { return cvs[i].Count > cvs[j].Count }
func (cvs byCount) Swap(i, j int)      { cvs[i], cvs[j] = cvs[j], cvs[i] }

type byDarkness []ColorVol

func (cvs byDarkness) Len() int { return len(cvs) }
func (cvs byDarkness) Less(i, j int) bool {
	return cvs[i].Lab.L() < cvs[j].Lab.L()
}
func (cvs byDarkness) Swap(i, j int) { cvs[i], cvs[j] = cvs[j], cvs[i] }

//**exported functions**//
// Create creates a new desktop theme based a provided palette and other options
func Create(p *Palette, opts map[string]interface{}) (*Theme, error) {
	var r, g, b uint32
	var hex string
	var keys []int
	t := make(Theme)

	for k := range *p {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		r, g, b, _ = (*p)[k].RGB.RGBA()
		hex = rgb2Hex((*p)[k].RGB)

		t["color"+strconv.Itoa(k)] = hex
		fmt.Printf("\033[38;2;%d;%d;%dm color%d = %s\n", byte(r), byte(g), byte(b), k, hex)
	}

	for k, v := range opts {
		t[k] = v
	}

	setDefaults(&t)

	return &t, nil
}

// Delegate converts a ColorVol slice to a Palette.
func Delegate(cvs *[]ColorVol) (*Palette, error) {
	p := make(Palette) // Palette to return

	// group colors into darks and lights
	sort.Sort(byDarkness(*cvs))
	d := (*cvs)[:len(*cvs)/2]
	l := (*cvs)[len(*cvs)/2:]

	// assign roles by prevalence
	sort.Sort(byCount(d))
	sort.Sort(byCount(l))
	for i, c := range d {
		p[i] = c
	}
	for i, c := range l {
		p[len(d)+i] = c
	}

	return &p, nil
}

// GetColors retrieves a set of colors of size `num` that best represent the image located at `path`.
func GetColors(path string, num int) (*[]ColorVol, error) {
	m := make(map[color.Color]int) // image colors and the number of pixels they each occupy
	cvs := make([]ColorVol, 0)     // ColorVol slice to return

	// get image file at specified path
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	// decode image
	i, _, e := image.Decode(f)
	if e != nil {
		return nil, e
	}

	// quantize image
	b := i.Bounds()
	o := image.NewNRGBA(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y))
	colorquant.NoDither.Quantize(i, o, num, false, true)

	// map each image color to its prevalence
	w, h := o.Bounds().Max.X, o.Bounds().Max.Y
	for x := 0; x < w; x += 5 {
		for y := 0; y < h; y += 5 {
			m[o.At(x, y)]++
		}
	}
	if len(m) != num {
		return nil, fmt.Errorf("Image at %s does not have enough variation to support a base %d color palette", path, num)
	}

	// convert map to ColorVol slice
	for k, v := range m {
		r, g, b, _ := k.RGBA()
		rgb := chromath.RGB{float64(byte(r)), float64(byte(g)), float64(byte(b))}
		xyz := rgb2Xyz.Convert(rgb)
		lab := lab2Xyz.Invert(xyz)
		cvs = append(cvs, ColorVol{k, lab, v})
	}

	return &cvs, nil
}

//**helper functions**//
// base is the color used for comparison. return value is postive
// if c0 is more different; negative if c1 is more different; 0 if
// both colors are just as different from base.
func diff(base chromath.Lab, c0 chromath.Lab, c1 chromath.Lab) float64 {
	return deltae.CIE2000(c0, base, klch) - deltae.CIE2000(c1, base, klch)
}

func rgb2Hex(rgb color.Color) string {
	r, g, b, _ := rgb.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", byte(r), byte(g), byte(b))
}

func setDefaults(t *Theme) {
	if _, ok := (*t)["background"]; !ok {
		(*t)["background"] = (*t)["color0"]
	}

	if _, ok := (*t)["transparency"]; !ok {
		(*t)["transparency"] = 1.0
	}

	if _, ok := (*t)["foreground"]; !ok {
		(*t)["foreground"] = (*t)["color8"]
	}
}
