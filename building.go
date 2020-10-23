package main

import (
	"github.com/EdlinOrg/prominentcolor"
	"image"
)

// ExtractColors returns (primary, secondary, tertiary) hexstrings of colors (without the # in front--just the hexstring)
func ExtractColors(img image.Image) (string, string, string, error) {
	centroids, err := prominentcolor.Kmeans(img)
	if err != nil {
		return "", "", "", err
	}
	colors := make([]string, 0, 3)
	for _, centroid := range centroids {
		colors = append(colors, centroid.AsString())
	}
	return colors[0], colors[1], colors[2], nil
}

//TODO: convert GIFs from `online: True` sources (YouTube, Dailymotion, Vimeo, you name it.). Might want to look at <https://github.com/hunterlong/gifs>
