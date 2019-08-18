package cmd

import (
	// "errors"
	"fmt"
	"github.com/esimov/colorquant"
	"github.com/jkl1337/go-chromath"
	"github.com/jkl1337/go-chromath/deltae"
	"github.com/spf13/cobra"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"sort"
)

var (
	TargetIlluminant = &chromath.IlluminantRefD50
	RGB2Xyz          = chromath.NewRGBTransformer(
		&chromath.SpaceSRGB,
		&chromath.AdaptationBradford,
		TargetIlluminant,
		&chromath.Scaler8bClamping,
		1.0,
		nil,
	)
	Lab2Xyz = chromath.NewLabTransformer(TargetIlluminant)
	klch    = &deltae.KLChDefault
	numPix  = 0.0
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
	Short: "Creates a new theme from image",
	Long:  `Creates a new theme from image`,
	// Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		i, e := quantize("/home/matt/Downloads/opal_fanart_su_by_urbietacreations_d8utzf1-pre.png", 18)
		// i, e := load("/home/matt/Downloads/Opal_-_Gen_1_With_Weapon.webp")
		if e != nil {
			log.Fatal(e)
		}

		m := getColors(i)
		cvs := map2ColorVolSlice(m)

		sort.Sort(byDarkness(cvs))

		d, l := splitDarkAndLight(&cvs)

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
			fmt.Println(k, ":", p[k])
		}

		myImage := image.NewRGBA(image.Rect(0, 0, 400, 900))
		outFile, e := os.Create("/home/matt/Downloads/test1.png")
		if e != nil {
			log.Fatal(e)
		}
		defer outFile.Close()

		x := 0
		y := 0
		for k, v := range p {
			if k < 0 {
				// for w := 150; w < 200; w++ {
				// 	for h := -k * 2; h < h+50; h++ {
				// 		myImage.Set(w, h, v.RGB)
				// 	}
				// }
				continue
			}
			for w := (k / 8) * 100; x < 100; w++ {
				x++
				for h := (k % 8) * 100; y < 100; h++ {
					y++
					myImage.Set(w, h, v.RGB)
				}
				y = 0
			}
			x = 0
			fmt.Println(k, ":", (k/9)*100, (k%9)*100)
		}

		// x := 0
		// y := 0
		// for _, cv := range cvs {
		// 	for w := x; w-x < 200; w++ {
		// 		for h := y; h-y < 200; h++ {
		// 			myImage.Set(w, h, cv.RGB)
		// 		}
		// 	}
		// 	x = (x + 200) % 800
		// 	if x == 0 {
		// 		y += 200
		// 	}
		// }

		png.Encode(outFile, myImage)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func delegate(dark *[]ColorVol, light *[]ColorVol) (map[int]ColorVol, error) {
	m := make(map[int]ColorVol)
	var bg ColorVol
	var fg ColorVol

	normal := []chromath.Lab{
		//black
		rgb2Lab(chromath.RGB{0.0, 0.0, 0.0}),
		//red
		rgb2Lab(chromath.RGB{128.0, 0.0, 0.0}),
		//green
		rgb2Lab(chromath.RGB{0.0, 128.0, 0.0}),
		//yellow
		rgb2Lab(chromath.RGB{128.0, 128.0, 0.0}),
		//blue
		rgb2Lab(chromath.RGB{0.0, 0.0, 128.0}),
		//magenta
		rgb2Lab(chromath.RGB{128.0, 0.0, 128.0}),
		//cyan
		rgb2Lab(chromath.RGB{0.0, 128.0, 128.0}),
		//white
		rgb2Lab(chromath.RGB{192.0, 192.0, 192.0}),
	}

	// bright colors
	bright := []chromath.Lab{
		//black
		rgb2Lab(chromath.RGB{128.0, 128.0, 128.0}),
		//red
		rgb2Lab(chromath.RGB{255.0, 0.0, 0.0}),
		//green
		rgb2Lab(chromath.RGB{0.0, 255.0, 0.0}),
		//yellow
		rgb2Lab(chromath.RGB{255.0, 255.0, 0.0}),
		//blue
		rgb2Lab(chromath.RGB{0.0, 0.0, 255.0}),
		//magenta
		rgb2Lab(chromath.RGB{255.0, 0.0, 255.0}),
		//cyan
		rgb2Lab(chromath.RGB{0.0, 255.0, 255.0}),
		//white
		rgb2Lab(chromath.RGB{255.0, 255.0, 255.0}),
	}

	sort.Sort(byCount(*dark))
	sort.Sort(byCount(*light))

	// most prominent color --> background
	bg = (*dark)[0]
	m[-2] = bg

	// contrast to prominent --> foreground
	fg = (*light)[0]
	for _, c := range *light {
		if diff(bg.Lab, c.Lab, fg.Lab) > 0 {
			fg = c
		}
	}
	m[-1] = fg

	for _, d := range *dark {
		if d == bg {
			continue
		}

		var sim int
		for i := range bright {
			if _, ok := m[8+i]; ok {
				continue
			}
			sim = i
			break
		}
		for i := range normal {
			if _, ok := m[i]; ok {
				continue
			}

			if diff(d.Lab, normal[i], normal[sim]) < 0 {
				sim = i
			}
		}
		fmt.Println(sim)

		m[sim] = d
	}

	for _, l := range *light {
		if l == fg {
			continue
		}

		var sim int
		for i := range bright {
			if _, ok := m[8+i]; ok {
				continue
			}
			sim = i
			break
		}
		for i := range bright {
			if _, ok := m[8+i]; ok {
				continue
			}

			if diff(l.Lab, bright[i], bright[sim]) < 0 {
				sim = i
			}
		}
		fmt.Println(sim)

		m[8+sim] = l
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
	for x := 0; x < w; x += 1 {
		for y := 0; y < h; y += 1 {
			c := img.At(x, y)
			if _, _, _, a := c.RGBA(); a != 0 {
				m[c]++
				numPix++
			}
		}
	}

	return m
}

// converts a color-to-int map to a ColorVol slice
func map2ColorVolSlice(m map[color.Color]int) []ColorVol {
	cvs := make([]ColorVol, 0)

	i := 0
	for k, v := range m {
		if float64(m[k])/numPix < 0.0005 {
			continue
		}

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
	o := image.NewRGBA(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y))

	colorquant.NoDither.Quantize(i, o, num, false, true)

	return o, nil
}

// converts an RGB color to its Lab equivalent
func rgb2Lab(rgb chromath.RGB) chromath.Lab {
	xyz := RGB2Xyz.Convert(rgb)
	return Lab2Xyz.Invert(xyz)
}

func splitDarkAndLight(cvs *[]ColorVol) ([]ColorVol, []ColorVol) {
	sort.Sort(byDarkness(*cvs))

	d := (*cvs)[:len(*cvs)/2]
	l := (*cvs)[len(*cvs)/2:]

	return d, l
}
