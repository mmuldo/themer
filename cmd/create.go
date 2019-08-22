package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/esimov/colorquant"
	"github.com/jkl1337/go-chromath"
	"github.com/jkl1337/go-chromath/deltae"
	"github.com/spf13/cobra"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
)

var (
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

// ColorVol represents an RGB color, its Lab equivalent, and the number of pixels it
// takes up in a given image.
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

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new theme",
	Long: `Creates a new theme. Currently requires
an image to be specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		var r, g, b uint32
		var hex string
		theme := make(map[string]interface{})

		// quantize provide image
		i, e := quantize(imgFile, 16)
		if e != nil {
			log.Fatal(e)
		}

		// get colors from quantized image and convert to
		// a slice of ColorVols
		m := getColors(i)
		cvs := map2ColorVolSlice(m)

		// split colors into darks and lights
		d, l := splitDarkAndLight(&cvs)

		// assign colors to roles (color0, color1, etc.)
		p, e := delegate(&d, &l)
		if e != nil {
			log.Fatal(e)
		}

		var keys []int
		for k := range p {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			r, g, b, _ = p[k].RGB.RGBA()
			hex = rgb2Hex(p[k].RGB)

			theme["color"+strconv.Itoa(k)] = hex
			fmt.Printf("\033[38;2;%d;%d;%dm color%d = %s\n", byte(r), byte(g), byte(b), k, hex)
		}
		theme["background"] = theme["color0"]
		theme["foreground"] = theme["color8"]

		e = save(theme, name)
		if e != nil {
			log.Fatal(e)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func delegate(dark *[]ColorVol, light *[]ColorVol) (map[int]ColorVol, error) {
	m := make(map[int]ColorVol)

	sort.Sort(byCount(*dark))
	sort.Sort(byCount(*light))

	for i, d := range *dark {
		m[i] = d
	}

	for i, l := range *light {
		m[len(*dark)+i] = l
	}

	return m, nil
}

// base is the color used for comparison. return value is postive
// if c0 is more different; negative if c1 is more different; 0 if
// both colors are just as different from base.
func diff(base chromath.Lab, c0 chromath.Lab, c1 chromath.Lab) float64 {
	return deltae.CIE2000(c0, base, klch) - deltae.CIE2000(c1, base, klch)
}

// returns a map of an image's colors and the number of times each color occurs
func getColors(img image.Image) map[color.Color]int {
	m := make(map[color.Color]int)

	w, h := img.Bounds().Max.X, img.Bounds().Max.Y
	for x := 0; x < w; x += 5 {
		for y := 0; y < h; y += 5 {
			m[img.At(x, y)]++
		}
	}

	return m
}

// converts a color-to-int map to a ColorVol slice
func map2ColorVolSlice(m map[color.Color]int) []ColorVol {
	cvs := make([]ColorVol, 0)

	i := 0
	for k, v := range m {
		r, g, b, _ := k.RGBA()
		rgb := chromath.RGB{float64(byte(r)), float64(byte(g)), float64(byte(b))}
		lab := rgb2Lab(rgb)
		cvs = append(cvs, ColorVol{k, lab, v})
		i++
	}

	return cvs
}

// loads an image for use given a file path
func quantize(path string, num int) (image.Image, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	i, _, e := image.Decode(f)
	if e != nil {
		return nil, e
	}
	b := i.Bounds()
	o := image.NewNRGBA(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y))

	colorquant.NoDither.Quantize(i, o, num, false, true)

	return o, nil
}

func rgb2Hex(rgb color.Color) string {
	r, g, b, _ := rgb.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", byte(r), byte(g), byte(b))
}

// converts an RGB color to its Lab equivalent
func rgb2Lab(rgb chromath.RGB) chromath.Lab {
	xyz := rgb2Xyz.Convert(rgb)
	return lab2Xyz.Invert(xyz)
}

func save(theme map[string]interface{}, name string) error {
	p := path.Join(os.Getenv("HOME"), ".config", "themer", "themes")

	if i, e := os.Stat(p); os.IsNotExist(e) || !i.IsDir() {
		os.MkdirAll(p, os.ModePerm)
	}

	json, e := json.MarshalIndent(theme, "", "\t")
	if e != nil {
		return e
	}

	e = ioutil.WriteFile(path.Join(p, name), json, 0644)
	if e != nil {
		return e
	}

	return nil
}

func splitDarkAndLight(cvs *[]ColorVol) ([]ColorVol, []ColorVol) {
	sort.Sort(byDarkness(*cvs))

	d := (*cvs)[:len(*cvs)/2]
	l := (*cvs)[len(*cvs)/2:]

	return d, l
}
