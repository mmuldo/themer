package image

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"sort"
)

type ColorCount struct {
	Color color.Color
	Count int
}

type ColorCountList []ColorCount

func (ccl ColorCountList) Len() int           { return len(ccl) }
func (ccl ColorCountList) Less(i, j int) bool { return ccl[i].Count > ccl[j].Count }
func (ccl ColorCountList) Swap(i, j int)      { ccl[i], ccl[j] = ccl[j], ccl[i] }

// GetColors returns a map of an image's colors
// and the number of times each color occurs
func GetColors(img *image.Image) map[color.Color]int {
	m := make(map[color.Color]int)

	w, h := (*img).Bounds().Max.X, (*img).Bounds().Max.Y
	for x := 0; x < w; x += 1 {
		for y := 0; y < h; y += 1 {
			c := (*img).At(x, y)
			m[c]++
		}
	}

	return m
}

func RankColors(m map[color.Color]int) ColorCountList {
	cc := make(ColorCountList, len(m))

	i := 0
	for k, v := range m {
		cc[i] = ColorCount{k, v}
		i++
	}

	sort.Sort(cc)
	return cc
}
