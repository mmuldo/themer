package cmd

import (
	// "fmt"
	"github.com/jkl1337/go-chromath"
	"github.com/jkl1337/go-chromath/deltae"
	"github.com/spf13/cobra"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"sort"
)

const (
	de = 10
)

var (
	TargetIlluminant = &chromath.IlluminantRefD50
	RGB2Xyz          = chromath.NewRGBTransformer(&chromath.SpaceSRGB, &chromath.AdaptationBradford, TargetIlluminant, &chromath.Scaler8bClamping, 1.0, nil)
	Lab2Xyz          = chromath.NewLabTransformer(TargetIlluminant)
	klch             = &deltae.KLChDefault
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new theme from image",
	Long:  `Creates a new theme from image`,
	// Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		i, e := load("/home/matt/Downloads/Opal_-_Gen_1_With_Weapon.webp")
		if e != nil {
			log.Fatal(e)
		}

		m := getColors(i)
		cvs := map2ColorVolSlice(m)
		sort.Sort(byCount(cvs))

		// groups := groupColorVols(cvs)
		// consolidate(&groups)
		// for _, g := range groups {
		// 	fmt.Printf("%+v\n", g)
		// }
		// fmt.Println(len(groups))

		myImage := image.NewRGBA(image.Rect(0, 0, 800, 1200))
		outFile, e := os.Create("/home/matt/Downloads/test.png")
		if e != nil {
			log.Fatal(e)
		}
		defer outFile.Close()

		x := 0
		y := 0
		for i, cv := range cvs {
			if i >= 18 {
				break
			}
			for w := x; w-x < 200; w++ {
				for h := y; h-y < 200; h++ {
					myImage.Set(w, h, cv.RGB)
				}
			}
			x = (x + 200) % 800
			if x == 0 {
				y += 200
			}
		}

		png.Encode(outFile, myImage)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

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

type byDeviance []ColorVol

func (cvs byDeviance) Len() int { return len(cvs) }
func (cvs byDeviance) Less(i, j int) bool {
	average := average(cvs).Lab
	return deltae.CIE2000(cvs[i].Lab, average, klch) > deltae.CIE2000(cvs[j].Lab, average, klch)
}
func (cvs byDeviance) Swap(i, j int) { cvs[i], cvs[j] = cvs[j], cvs[i] }

func average(cvs []ColorVol) ColorVol {
	rt, gt, bt := 0.0, 0.0, 0.0
	t := 0

	for _, cv := range cvs {
		r, g, b, _ := cv.RGB.RGBA()
		rt += math.Pow(float64(uint8(r)), 2)
		gt += math.Pow(float64(uint8(g)), 2)
		bt += math.Pow(float64(uint8(b)), 2)

		t += cv.Count
	}

	rgb := chromath.RGB{
		math.Sqrt(rt / float64(len(cvs))),
		math.Sqrt(gt / float64(len(cvs))),
		math.Sqrt(bt / float64(len(cvs))),
	}

	lab := rgb2Lab(rgb)

	return ColorVol{
		color.RGBA{
			uint8(math.Round(rgb.R())),
			uint8(math.Round(rgb.G())),
			uint8(math.Round(rgb.B())),
			255,
		},
		lab,
		t,
	}
}

func consolidate(gs *[][]ColorVol) {
	for len(*gs) > 18 {
		var r float64
		var a, b int
		diff := deltae.CIE2000(average((*gs)[0]).Lab, average((*gs)[1]).Lab, klch)

		for i := 0; i < len(*gs); i++ {
			for j := i + 1; j < len(*gs); j++ {
				r = deltae.CIE2000(average((*gs)[i]).Lab, average((*gs)[j]).Lab, klch)
				if r < diff {
					diff = r
					a = i
					b = j
				}
			}
		}

		(*gs)[a] = append((*gs)[a], (*gs)[b]...)
		(*gs) = append((*gs)[:b], (*gs)[b+1:]...)
	}

	// TODO: split
	// for len(*gs) < 18 {
	// 	var a int
	// 	var r float64
	// 	dev := 0.0

	// 	for _, g := range *gs {
	// 		r =
	// 	}
	// }
}

// returns a map of an image's colors and the number of times each color occurs
func getColors(img *image.Image) map[color.Color]int {
	m := make(map[color.Color]int)

	w, h := (*img).Bounds().Max.X, (*img).Bounds().Max.Y
	for x := 0; x < w; x += 1 {
		for y := 0; y < h; y += 1 {
			c := (*img).At(x, y)
			if _, _, _, a := c.RGBA(); a != 0 {
				m[c]++
			}
		}
	}

	return m
}

// func getDifference(cvs []ColorVol) (color.Color, color.Color, float64) {

// }

func groupColorVols(cvs []ColorVol) [][]ColorVol {
	g := make([][]ColorVol, 0)
	done := make([]bool, len(cvs))

	k := -1
	for i := range cvs {
		if done[i] {
			continue
		}
		g = append(g, []ColorVol{(cvs)[i]})
		k++
		done[i] = true

		for j := i + 1; j < len(cvs); j++ {
			if done[j] {
				continue
			}

			if deltae.CIE2000((cvs)[i].Lab, (cvs)[j].Lab, klch) < de {
				g[k] = append(g[k], (cvs)[j])
				done[j] = true
			}
		}
	}

	return g
}

// loads an image for use given a file path
func load(path string) (*image.Image, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	i, _, e := image.Decode(f)
	if e != nil {
		return nil, e
	}

	return &i, nil
}

// converts a color-to-int map to a ColorVol slice
func map2ColorVolSlice(m map[color.Color]int) []ColorVol {
	cvs := make([]ColorVol, len(m))

	i := 0
	for k, v := range m {
		r, g, b, _ := k.RGBA()
		rgb := chromath.RGB{float64(byte(r)), float64(byte(g)), float64(byte(b))}
		lab := rgb2Lab(rgb)
		cvs[i] = ColorVol{k, lab, v}
		i++
	}

	return cvs
}

// converts an RGB color to its Lab equivalent
func rgb2Lab(rgb chromath.RGB) chromath.Lab {
	xyz := RGB2Xyz.Convert(rgb)
	return Lab2Xyz.Invert(xyz)
}

// func splitGroup(cvl *image.ColorVolList) (image.ColorVolList, image.ColorVolList) {

// }
