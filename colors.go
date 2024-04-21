package ortfodb

import (
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	_ "golang.org/x/image/webp"
)

// ColorPalette reprensents the object in a Work's metadata.colors.
type ColorPalette struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Tertiary  string `json:"tertiary"`
}

func (colors ColorPalette) Empty() bool {
	return colors.Primary == "" && colors.Secondary == "" && colors.Tertiary == ""
}

// MergeWith merges the colors of the current palette with other: if a color is missing in the current palette, it is
// replaced by the one in other.
func (colors ColorPalette) MergeWith(other ColorPalette) ColorPalette {
	merged := colors
	if merged.Primary == "" {
		merged.Primary = other.Primary
	}
	if merged.Secondary == "" {
		merged.Secondary = other.Secondary
	}
	if merged.Tertiary == "" {
		merged.Tertiary = other.Tertiary
	}
	return merged
}

func canExtractColors(contentType string) bool {
	switch strings.Split(contentType, "/")[1] {
	case "jpeg", "png", "webp", "pbm", "ppm", "pgm":
		return true
	default:
		return false
	}
}

// ExtractColors extracts the 3 most proeminent colors from the given image-decodable file.
// See https://pkg.go.dev/image#Decode for what formats are decodable.
func ExtractColors(filename string) (ColorPalette, error) {
	file, err := os.Open(filename)
	if err != nil {
		return ColorPalette{}, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return ColorPalette{}, err
	}
	return kmeans(img)
}

// kmeans extracts colors from img.
func kmeans(img image.Image) (ColorPalette, error) {
	centroids, err := prominentcolor.Kmeans(img)
	if err != nil {
		return ColorPalette{}, err
	}
	colors := make([]string, 0)
	for _, centroid := range centroids {
		colors = append(colors, centroid.AsString())
	}
	if len(colors) == 0 {
		return ColorPalette{}, fmt.Errorf("no colors found in given image")
	}

	primary := "#" + colors[0]
	secondary := ""
	tertiary := ""
	if len(colors) > 1 {
		secondary = "#" + colors[1]
	}
	if len(colors) > 2 {
		tertiary = "#" + colors[2]
	}

	return ColorPalette{
		Primary:   primary,
		Secondary: secondary,
		Tertiary:  tertiary,
	}, nil
}
