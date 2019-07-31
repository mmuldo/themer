package image

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// Load loads an image for use given a file path
func Load(path string) (*image.Image, error) {
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
