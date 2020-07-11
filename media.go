package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

// ImageDimensions represents metadata about a media as it's extracted from its file
type ImageDimensions struct {
	Width       int
	Height      int
	AspectRatio float32 `json:"aspect_ratio"`
}

// Thumbnail represents a thumbnail
type Thumbnail struct {
	Type       string
	MIMEType   *mimetype.MIME `json:"mime_type"`
	Format     string
	Source     string
	dimensions ImageDimensions
}

// Media represents a media object inserted in the work object's ``media`` array.
type Media struct {
	ID         string
	Alt        string
	Title      string
	Source     string
	Type       string
	MIMEType   *mimetype.MIME `json:"mime_type"`
	thumbnais  []Thumbnail
	Size       uint
	dimensions ImageDimensions
	duration   uint // In seconds
}

// ReadImage reads an image at ``filepath``, decodes it, and returns an ``image.Image`` object
func ReadImage(filepath string) (image.Image, error) {
	// Open the file
	reader, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// GetImageDimensions returns an ``ImageDimensions`` object, given the image object
func GetImageDimensions(img image.Image) (ImageDimensions, error) {
	// Get height & width
	height := img.Bounds().Dy()
	width := img.Bounds().Dx()
	// Get aspect ratio
	ratio := float32(width) / float32(height)
	return ImageDimensions{width, height, ratio}, nil
}
