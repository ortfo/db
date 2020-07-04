package main

import (
	"image"
	"os"
	_ "image/png"
	_ "image/jpeg"
	
	"github.com/gabriel-vasile/mimetype"
)

// ReadImage reads an image at ``filepath``, decodes it, and returns an ``image.Image`` object as well as some metadata
func ReadImage(filepath string) (image.Image, ImageMetadata, error) {
	img, err := OpenImage(filepath)
	if err != nil {
		return nil, ImageMetadata{}, err
	}
	meta, err := GetImageMetadata(img, filepath)
	if err != nil {
		return img, ImageMetadata{}, err
	}
	return img, meta, nil
}

// ReadImage reads an image at ``filepath``, decodes it, and returns an ``image.Image`` object
func OpenImage(filepath string) (image.Image, error) {
	// Open the file
	reader, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(reader)
	if err !=  nil {
		return nil, err
	}
	return img, nil
}

type ImageMetadata struct {
	Width int
	Height int
	AspectRatio float32
	MIMEType *mimetype.MIME
}

// GetImageMetadata returns an ``ImageMetadata`` object, given the image object and the image's filepath
// TODO: pass a file object instead for improved performance
func GetImageMetadata(img image.Image, filepath string) (ImageMetadata, error) {
	// Get filetype
	mime, err := mimetype.DetectFile(filepath)
	if err != nil {
		return ImageMetadata{}, err
	}
	// Get height & width
	height := img.Bounds().Dy()
	width := img.Bounds().Dx()
	// Get aspect ratio
	ratio := float32(width) / float32(height)
	return ImageMetadata{width, height, ratio, mime}, nil
}
