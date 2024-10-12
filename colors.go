package ortfodb

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"strings"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/zyedidia/generic/mapset"
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

// SortBySaturation sorts the palette by saturation. Primary will be the most saturated, tertiary the least.
// Empty or invalid colors are treated as having 0 saturation.
func (colors *ColorPalette) SortBySaturation() {
	primary := colors.Primary
	secondary := colors.Secondary
	tertiary := colors.Tertiary

	ll.Debug("sorting colors based on saturations: primary(%s) = %f, secondary(%s) = %f, tertiary(%s) = %f", primary, saturation(primary), secondary, saturation(secondary), tertiary, saturation(tertiary))

	if saturation(primary) < saturation(secondary) {
		primary, secondary = secondary, primary
	}
	if saturation(primary) < saturation(tertiary) {
		primary, tertiary = tertiary, primary
	}
	if saturation(secondary) < saturation(tertiary) {
		secondary, tertiary = tertiary, secondary
	}

	colors.Primary = primary
	colors.Secondary = secondary
	colors.Tertiary = tertiary
}

// mostSaturated returns at most n colors, sorted by descending saturation.
func paletteFromMostSaturated(colors mapset.Set[color.Color]) ColorPalette {
	bySaturation := make(map[string]float64, 0)
	colors.Each(func(color color.Color) {
		r, g, b, _ := color.RGBA()
		hex := colorful.Color{R: float64(r) / float64(0xffff), G: float64(g) / float64(0xffff), B: float64(b) / float64(0xffff)}.Hex()
		bySaturation[hex] = saturation(hex)
	})

	ll.Debug("paletteFromMostSaturated: bySaturation = %v", bySaturation)

	leastSaturatedSaturation := 0.0
	mostSaturateds := make([]string, 3)
	for hex, sat := range bySaturation {
		if sat > leastSaturatedSaturation {
			mostSaturateds[0], mostSaturateds[1], mostSaturateds[2] = hex, mostSaturateds[0], mostSaturateds[1]
			leastSaturatedSaturation = sat
		}
	}

	ll.Debug("paletteFromMostSaturated: mostSaturateds = %v", mostSaturateds)

	return ColorPalette{
		Primary:   mostSaturateds[0],
		Secondary: mostSaturateds[1],
		Tertiary:  mostSaturateds[2],
	}
}

// saturation returns the saturation of the given colorstring.
// invalid or empty colorstrings return 0.
func saturation(colorstring string) float64 {
	color, err := colorful.Hex(colorstring)
	if err != nil {
		return 0
	}

	_, sat, _ := color.Hsv()
	return sat
}

func canExtractColors(contentType string) bool {
	switch strings.Split(contentType, "/")[1] {
	case "jpeg", "png", "webp", "pbm", "ppm", "pgm", "gif":
		return true
	default:
		return false
	}
}

// ExtractColors extracts the 3 most proeminent colors from the given image-decodable file.
// See https://pkg.go.dev/image#Decode for what formats are decodable.
func ExtractColors(filename string, contentType string) (ColorPalette, error) {
	defer ll.TimeTrack(time.Now(), "ExtractColors", filename)
	file, err := os.Open(filename)
	if err != nil {
		return ColorPalette{}, err
	}
	defer file.Close()

	if contentType == "image/gif" {
		ll.Debug("extract colors from %s: decoding gif", filename)
		var decodedGif *gif.GIF
		decodedGif, err = gif.DecodeAll(file)
		if err != nil {
			return ColorPalette{}, fmt.Errorf("could not decode gif's config: %w", err)
		}

		gifColorsWithAppearanceCount := make(map[color.Color]int)
		for _, frame := range decodedGif.Image {
			for _, paletteIndex := range frame.Pix {
				gifColorsWithAppearanceCount[frame.Palette[paletteIndex]]++
			}
		}

		averageAppearanceCount := 0
		for _, count := range gifColorsWithAppearanceCount {
			averageAppearanceCount += count
		}
		averageAppearanceCount /= len(gifColorsWithAppearanceCount)

		gifColors := mapset.New[color.Color]()
		for color, count := range gifColorsWithAppearanceCount {
			if count > averageAppearanceCount/5 {
				gifColors.Put(color)
			}
		}

		ll.Debug("extract colors from %s: extracting most saturated colors from %d unique colors: %v", filename, gifColors.Size(), gifColorsWithAppearanceCount)

		return paletteFromMostSaturated(gifColors), nil

	}

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
		// retry without masking out backgrounds or cropping
		centroids, err = prominentcolor.KmeansWithAll(prominentcolor.DefaultK, img, prominentcolor.ArgumentNoCropping, prominentcolor.DefaultSize, []prominentcolor.ColorBackgroundMask{})
		if err != nil {
			return ColorPalette{}, err
		}
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

	palette := ColorPalette{
		Primary:   primary,
		Secondary: secondary,
		Tertiary:  tertiary,
	}
	palette.SortBySaturation()
	return palette, nil
}
