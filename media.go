package main

// Functions to analyze media files.
// Used to go from a ParsedDescription struct to a Work struct.

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
func AnalyzeMediaFile(filename string, embedDeclaration MediaEmbedDeclaration) (Media, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Media{}, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return Media{}, err
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
			return Media{}, err
		}
	}

	if isVideo {
		dimensions, duration, err = GetVideoDimensionsDuration(filename, dimensions, duration)
	}

	if isAudio {
		duration = GetAudioDuration(file)
	}

	return Media{
		ID:          slugify.Marshal(FilepathBaseNoExt(filename)),
		Alt:         embedDeclaration.Alt,
		Title:       embedDeclaration.Title,
		Source:      filename,
		ContentType: contentType,
		Dimensions:  dimensions,
		Duration:    duration,
		Size:        uint64(fileInfo.Size()),
	}, nil
}

// GetAudioDuration takes in an os.File and returns the duration of the audio file in seconds. If any error occurs the duration will be 0.
func GetAudioDuration(file *os.File) uint {
	var duration uint
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

// GetVideoDimensionsDuration returns an ImageDimensions struct with the video's height, width and aspect ratio and a duration in seconds.
func GetVideoDimensionsDuration(filename string, dimensions ImageDimensions, duration uint) (ImageDimensions, uint, error) {
	video, err := screengen.NewGenerator(filename)
	if err != nil {
		return ImageDimensions{}, 0, err
	}
	height := video.Height()
	width := video.Width()
	dimensions = ImageDimensions{
		Height:      height,
		Width:       width,
		AspectRatio: float32(width) / float32(height),
	}
	duration = uint(video.Duration) / 1000
	return dimensions, duration, nil
}

// AnalyzeAllMediae analyzes all the mediae from ParsedDescription's MediaEmbedDeclarations and returns analyzed mediae, ready for use as Work.Media
func AnalyzeAllMediae(embedDeclarations map[string][]MediaEmbedDeclaration, currentDirectory string) (map[string][]Media, error) {
	analyzedMediae := make(map[string][]Media, 0)
	analyzedMediaeBySource := make(map[string]Media, 0)
	for language, mediae := range embedDeclarations {
		analyzedMediae[language] = make([]Media, 0)
		for _, media := range mediae {
			var filename string
			if !filepath.IsAbs(media.Source) {
				filename, _ = filepath.Abs(path.Join(currentDirectory, media.Source))
			} else {
				filename = media.Source
			}
			if IsValidURL(media.Source) {
				analyzedMedia := Media{
					Alt:    media.Alt,
					Title:  media.Title,
					Source: media.Source,
					Online: true,
				}
				analyzedMediae[language] = append(analyzedMediae[language], analyzedMedia)
			} else if alreadyAnalyzedMedia, ok := analyzedMediaeBySource[filename]; ok {

				analyzedMediae[language] = append(analyzedMediae[language], alreadyAnalyzedMedia)
			} else {
				analyzedMedia, err := AnalyzeMediaFile(filename, media)
				if err != nil {
					return map[string][]Media{}, err
				}
				analyzedMediae[language] = append(analyzedMediae[language], analyzedMedia)
				analyzedMediaeBySource[filename] = analyzedMedia
			}
		}
	}
	return analyzedMediae, nil
}
