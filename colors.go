package ortfodb

import (
	"image"
	"os"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/gabriel-vasile/mimetype"
)

// ExtractedColors reprensents the object in a Work's metadata.colors.
type ExtractedColors struct {
	Primary   string
	Secondary string
	Tertiary  string
}

// StepExtractColors executes the step "extract colors" and returns a metadata object with the colors entry modified accordingly.
func (ctx *RunContext) StepExtractColors(metadata map[string]interface{}, mediaPaths []string) map[string]interface{} {
	// Do not overwrite manually-set colors metadata entry
	if _, ok := metadata["colors"]; !ok {
		// Get only image filepaths
		imageFilepaths := filterSlice(mediaPaths, func(item string) bool {
			contentType, err := mimetype.DetectFile(item)
			return err == nil && strings.HasPrefix(contentType.String(), "image/")
		})
		// Extract colors from them
		extractFromFile := selectFileToExtractColorsFrom(imageFilepaths, ctx.Config.ExtractColors.DefaultFiles)
		if extractFromFile != "" {
			ctx.Status(StepColorExtraction, ProgressDetails{
				File: extractFromFile,
			})
			extractedColors, err := ExtractColors(extractFromFile)
			if err == nil {
				metadata["colors"] = extractedColors
			}
		}
	}
	return metadata
}

func selectFileToExtractColorsFrom(files []string, defaultFiles []string) string {
	if len(files) == 0 {
		return ""
	}
	if len(files) == 1 {
		return files[0]
	}
	for _, filename := range files {
		if stringInSlice(defaultFiles, filename) {
			return filename
		}
	}
	return files[0]
}

// ExtractColors extracts the 3 most proeminent colors from the given image-decodable file.
// See https://pkg.go.dev/image#Decode for what formats are decodable.
func ExtractColors(filename string) (ExtractedColors, error) {
	file, err := os.Open(filename)
	if err != nil {
		return ExtractedColors{}, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return ExtractedColors{}, err
	}
	return kmeans(img)
}

// kmeans extracts colors from img.
func kmeans(img image.Image) (ExtractedColors, error) {
	centroids, err := prominentcolor.Kmeans(img)
	if err != nil {
		return ExtractedColors{}, err
	}
	colors := make([]string, 3)
	for _, centroid := range centroids {
		colors = append(colors, centroid.AsString())
	}
	return ExtractedColors{
		Primary:   colors[0],
		Secondary: colors[1],
		Tertiary:  colors[2],
	}, nil
}
