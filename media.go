package ortfodb

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

// ImageDimensions represents metadata about a media as it's extracted from its file.
type ImageDimensions struct {
	Width       int
	Height      int
	AspectRatio float32
}

// Thumbnail represents a thumbnail.
type Thumbnail struct {
	Type        string
	ContentType string
	Format      string
	Source      string
	Dimensions  ImageDimensions
}

// Media represents a media object inserted in the work object's media array.
type Media struct {
	ID    string
	Alt   string
	Title string
	// Source is the media's path, verbatim from the embed declaration (what's actually written in the description file)
	Source string
	// AbsolutePath is the actual location of the file as an absolute path
	AbsolutePath string
	// Path is AbsolutePath with transformations applied, following the configuration.
	// See Configuration.ReplaceMediaSources
	Path        string
	ContentType string
	Size        uint64 // In bytes
	Dimensions  ImageDimensions
	Duration    uint // In seconds
	Online      bool // Whether the media is hosted online (referred to by an URL)
	Attributes  MediaAttributes
	HasSound    bool // The media is either an audio file or a video file that contains an audio stream
}

// GetImageDimensions returns an ImageDimensions object, given a pointer to a file.
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

// AnalyzeMediaFile analyzes the file at its absolute filepath filename and returns a Media struct, merging the analysis' results with information from the matching MediaEmbedDeclaration.
func (ctx *RunContext) AnalyzeMediaFile(filename string, embedDeclaration MediaEmbedDeclaration) (Media, error) {
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

	var dimensions ImageDimensions
	var duration uint
	var hasSound bool

	if isImage {
		dimensions, err = GetImageDimensions(file)
		if err != nil {
			return Media{}, err
		}
	}

	if isVideo {
		dimensions, duration, hasSound, err = AnalyzeVideo(filename)
		if err != nil {
			return Media{}, err
		}
	}

	if isAudio {
		duration = AnalyzeAudio(file)
		hasSound = true
	}

	return Media{
		ID:           slugify.Marshal(filepathBaseNoExt(filename)),
		Alt:          embedDeclaration.Alt,
		Title:        embedDeclaration.Title,
		Source:       embedDeclaration.Source,
		AbsolutePath: filename,
		Path:         ctx.TransformSource(filename),
		ContentType:  contentType,
		Dimensions:   dimensions,
		Duration:     duration,
		Size:         uint64(fileInfo.Size()),
		Attributes:   embedDeclaration.Attributes,
		HasSound:     hasSound,
	}, nil
}

// TransformSource returns the appropriate URI (HTTPS, local...), taking into account the configuration.
func (ctx *RunContext) TransformSource(source string) string {
	for _, replacement := range ctx.Config.ReplaceMediaSources {
		source = strings.ReplaceAll(source, replacement.Replace, replacement.With)
	}
	return source
}

// AnalyzeAudio takes in an os.File and returns the duration of the audio file in seconds. If any error occurs the duration will be 0.
func AnalyzeAudio(file *os.File) uint {
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

// AnalyzeVideo returns an ImageDimensions struct with the video's height, width and aspect ratio and a duration in seconds.
func AnalyzeVideo(filename string) (dimensions ImageDimensions, duration uint, hasSound bool, err error) {
	video, err := screengen.NewGenerator(filename)
	if err != nil {
		return
	}
	height := video.Height()
	width := video.Width()
	dimensions = ImageDimensions{
		Height:      height,
		Width:       width,
		AspectRatio: float32(width) / float32(height),
	}
	duration = uint(video.Duration) / 1000
	hasSound = video.AudioCodec != ""
	return
}

// AnalyzeAllMediae analyzes all the mediae from ParsedDescription's MediaEmbedDeclarations and returns analyzed mediae, ready for use as Work.Media.
func (ctx *RunContext) AnalyzeAllMediae(embedDeclarations map[string][]MediaEmbedDeclaration, currentDirectory string) (map[string][]Media, error) {
	if ctx.Flags.Scattered {
		currentDirectory = path.Join(currentDirectory, ".portfoliodb")
	}
	analyzedMediae := make(map[string][]Media)
	analyzedMediaeBySource := make(map[string]Media)
	for language, mediae := range embedDeclarations {
		analyzedMediae[language] = make([]Media, 0)
		for _, media := range mediae {
			// Handle sources which are URLs
			if isValidURL(media.Source) {
				analyzedMedia := Media{
					Alt:        media.Alt,
					Title:      media.Title,
					Source:     media.Source,
					Online:     true,
					Attributes: media.Attributes,
				}
				analyzedMediae[language] = append(analyzedMediae[language], analyzedMedia)
				continue
			}

			// Compute absolute filepath to media
			var filename string
			if !filepath.IsAbs(media.Source) {
				filename, _ = filepath.Abs(path.Join(currentDirectory, media.Source))
			} else {
				filename = media.Source
			}

			// Skip already-analyzed
			if alreadyAnalyzedMedia, ok := analyzedMediaeBySource[filename]; ok {
				// Update fields independent of media.Source
				alreadyAnalyzedMedia.Alt = media.Alt
				alreadyAnalyzedMedia.Attributes = media.Attributes
				alreadyAnalyzedMedia.Title = media.Title
				analyzedMediae[language] = append(analyzedMediae[language], alreadyAnalyzedMedia)
				continue
			}

			ctx.Status("Analyzing " + path.Base(filename))
			analyzedMedia, err := ctx.AnalyzeMediaFile(filename, media)
			if err != nil {
				return map[string][]Media{}, err
			}
			analyzedMediae[language] = append(analyzedMediae[language], analyzedMedia)
			analyzedMediaeBySource[filename] = analyzedMedia
		}
	}
	return analyzedMediae, nil
}
