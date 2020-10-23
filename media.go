package main

// Functions to analyze media files.
// Used to go from a ParsedDescription struct to a WorkObject.

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/metal3d/go-slugify"
	"github.com/tcolgate/mp3"
	"gitlab.com/opennota/screengen"
)

// ImageDimensions represents metadata about a media as it's extracted from its file
type ImageDimensions struct {
	Width       int
	Height      int
	AspectRatio float32
}

// Thumbnail represents a thumbnail
type Thumbnail struct {
	Type        string
	ContentType string
	Format      string
	Source      string
	dimensions  ImageDimensions
}

// Media represents a media object inserted in the work object's ``media`` array.
type Media struct {
	ID          string
	Alt         string
	Title       string
	Source      string
	ContentType string
	Size        uint64 // In bytes
	Dimensions  ImageDimensions
	Duration    uint // In seconds
	Online      bool // Whether the media is hosted online (referred to by an URL)
}

// TODO: support for pdf files.

// GetImageDimensions returns an ``ImageDimensions`` object, given a pointer to a file
func GetImageDimensions(file *os.File) (ImageDimensions, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return ImageDimensions{}, err
	}
	// Get height & width
	height := img.Bounds().Dy()
	width := img.Bounds().Dx()
	// Get aspect ratio
	ratio := float32(width) / float32(height)
	return ImageDimensions{width, height, ratio}, nil
}

// AnalyzeMediaFile analyzes the file at filename and returns a Media struct, merging the analysis' results with information from the matching MediaEmbedDeclaration
func AnalyzeMediaFile(filename string, embedDeclaration MediaEmbedDeclaration) Media {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	var contentType string
	mimeType, err := mimetype.DetectFile(filename)
	if err != nil {
		contentType = "application/octet-stream"
	} else {
		contentType = mimeType.String()
	}

	isAudio := strings.HasPrefix(contentType, "audio/")
	isVideo := strings.HasPrefix(contentType, "video/")
	isImage := strings.HasPrefix(contentType, "image/")

	dimensions := ImageDimensions{}
	var duration uint

	if isImage {
		dimensions, err = GetImageDimensions(file)
		if err != nil {
			panic(err)
		}
	}

	if isVideo {
		dimensions, duration = GetVideoDimensionsDuration(filename, dimensions, duration)
	}

	if isAudio {
		duration = GetAudioDuration(file, duration)
	}

	return Media{
		ID:          slugify.Marshal(filepath.Base(filename)),
		Alt:         embedDeclaration.Alt,
		Title:       embedDeclaration.Title,
		Source:      filename,
		ContentType: contentType,
		Dimensions:  dimensions,
		Duration:    duration,
		Size:        uint64(fileInfo.Size()),
	}
}

func GetAudioDuration(file *os.File, duration uint) uint {
	decoder := mp3.NewDecoder(file)
	skipped := 0
	var frame mp3.Frame
	for {
		err := decoder.Decode(&frame, &skipped)
		if err != nil {
			break
		} else {
			duration += uint(frame.Duration().Seconds())
		}
	}
	return duration
}

func GetVideoDimensionsDuration(filename string, dimensions ImageDimensions, duration uint) (ImageDimensions, uint) {
	video, err := screengen.NewGenerator(filename)
	if err != nil {
		panic(err)
	}
	height := video.Height()
	width := video.Width()
	dimensions = ImageDimensions{
		Height:      height,
		Width:       width,
		AspectRatio: float32(width) / float32(height),
	}
	duration = uint(video.Duration) / 1000
	return dimensions, duration
}

func AnalyzeAllMedia(embedDeclarations map[string][]MediaEmbedDeclaration, currentDirectory string) map[string][]Media {
	analyzedMediae := make(map[string][]Media, 0)
	analyzedMediaeBySource := make(map[string]Media, 0)
	for language, mediae := range embedDeclarations {
		analyzedMediae[language] = make([]Media, 0)
		for _, media := range mediae {
			filepath, _ := filepath.Abs(path.Join(currentDirectory, media.Source))
			if IsValidURL(media.Source) {
				analyzedMedia := Media{
					Alt:    media.Alt,
					Title:  media.Title,
					Source: media.Source,
					Online: true,
				}
				analyzedMediae[language] = append(analyzedMediae[language], analyzedMedia)
			} else if alreadyAnalyzedMedia, ok := analyzedMediaeBySource[filepath]; ok {

				analyzedMediae[language] = append(analyzedMediae[language], alreadyAnalyzedMedia)
			} else {
				analyzedMedia := AnalyzeMediaFile(filepath, media)
				analyzedMediae[language] = append(analyzedMediae[language], analyzedMedia)
				analyzedMediaeBySource[filepath] = analyzedMedia
			}
		}
	}
	return analyzedMediae
}
