package ortfodb

// Functions to analyze media files.
// Used to go from a ParsedDescription struct to a Work struct.

import (
	"errors"
	"fmt"
	"image"

	// Supported formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/hullerob/go.farbfeld"
	_ "github.com/jbuchbinder/gopnm" // PBM, PGM and PPM
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/ccitt"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"

	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gen2brain/go-fitz"
	"github.com/lafriks/go-svg"
	"github.com/metal3d/go-slugify"
	recurcopy "github.com/plus3it/gorecurcopy"
	ffmpeg "github.com/ssttevee/go-ffmpeg"
	"github.com/tcolgate/mp3"
)

// ImageDimensions represents metadata about a media as it's extracted from its file.
type ImageDimensions struct {
	Width       int
	Height      int
	AspectRatio float32 `json:"aspect_ratio"`
}

// Media represents a media object inserted in the work object's media array.
type Media struct {
	ID    string
	Alt   string
	Title string
	// Source is the media's path, verbatim from the embed declaration (what's actually written in the description file).
	Source string
	// Path is the media's path, relative to (media directory)/(work ID).
	// See Configuration.Media.At.
	Path       string
	Attributes MediaAttributes
	// Analysis
	ContentType     string `json:"content_type"`
	Size            uint64 // In bytes
	Dimensions      ImageDimensions
	Online          bool            // Whether the media is hosted online (referred to by an URL)
	Duration        uint            // In seconds (except for PDFs, where it is in page count)
	HasSound        bool            `json:"has_sound"` // The media is either an audio file or a video file that contains an audio stream
	ExtractedColors ExtractedColors `json:"extracted_colors"`
	Thumbnails      map[uint16]string
}

func (ctx *RunContext) AbsolutePathToMedia(media Media) string {
	return path.Join(ctx.Config.Media.At, ctx.CurrentWorkID, media.Path)
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

func removeUnit(s string) string {
	return strings.TrimSuffix(s, "px")
}

// GetSVGDimensions returns an ImageDimensions object, given a pointer to a SVG file.
// If neither viewBox nor width & height attributes are set, the resulting dimensions will be 0x0.
func GetSVGDimensions(file *os.File) (ImageDimensions, error) {
	svg, err := svg.Parse(file, svg.IgnoreErrorMode)
	if err != nil {
		return ImageDimensions{}, fmt.Errorf("while parsing SVG file: %w", err)
	}

	var height, width float64
	if svg.Width != "" && svg.Height != "" {
		height, err = strconv.ParseFloat(removeUnit(svg.Height), 32)
		if err != nil {
			return ImageDimensions{}, fmt.Errorf("cannot parse SVG height attribute as a number: %w", err)
		}
		width, err = strconv.ParseFloat(removeUnit(svg.Width), 32)
		if err != nil {
			return ImageDimensions{}, fmt.Errorf("cannot parse SVG width attribute as a number: %w", err)
		}
	} else if svg.ViewBox.H != 0 && svg.ViewBox.W != 0 {
		height = float64(svg.ViewBox.H)
		width = float64(svg.ViewBox.W)
	} else {
		return ImageDimensions{}, fmt.Errorf("cannot determine dimensions of SVG file")
	}

	return ImageDimensions{
		Width:       int(width),
		Height:      int(height),
		AspectRatio: float32(width) / float32(height),
	}, nil
}

// PathToWorkFolder returns the path to the work's folder, including the .portfoliodb part if --scattered.
func (ctx *RunContext) PathToWorkFolder(workID string) string {
	path := filepath.Join(ctx.DatabaseDirectory, workID)
	if ctx.Flags.Scattered {
		path = filepath.Join(path, ctx.Config.ScatteredModeFolder)
	}
	return path
}

// AnalyzeMediaFile analyzes the file at its absolute filepath filename and returns a Media struct, merging the analysis' results with information from the matching MediaEmbedDeclaration.
func (ctx *RunContext) AnalyzeMediaFile(workID string, embedDeclaration MediaEmbedDeclaration) (usedCache bool, media Media, err error) {
	ctx.Status(StepMediaAnalysis, ProgressDetails{
		Resolution: 0,
		File: embedDeclaration.Source,
	})
	// Compute absolute filepath to media
	var filename string
	if !filepath.IsAbs(embedDeclaration.Source) {
		filename, _ = filepath.Abs(filepath.Join(ctx.PathToWorkFolder(workID), embedDeclaration.Source))
	} else {
		filename = embedDeclaration.Source
	}
	file, err := os.Open(filename)
	if err != nil {
		return false, Media{}, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return false, Media{}, err
	}

	var contentType string

	if fileInfo.IsDir() {
		contentType = "directory"
	} else {
		usedCache, cachedAnalysis := ctx.UseCache(filename, embedDeclaration)
		if usedCache && cachedAnalysis.ContentType != "" {
			return true, cachedAnalysis, nil
		}

		mimeType, err := mimetype.DetectFile(filename)
		if err != nil {
			contentType = "application/octet-stream"
		} else {
			contentType = mimeType.String()
		}
	}

	isAudio := strings.HasPrefix(contentType, "audio/")
	isVideo := strings.HasPrefix(contentType, "video/")
	isImage := strings.HasPrefix(contentType, "image/")
	isPDF := contentType == "application/pdf"

	var dimensions ImageDimensions
	var duration uint
	var hasSound bool

	if isImage {
		if contentType == "image/svg" || contentType == "image/svg+xml" {
			dimensions, err = GetSVGDimensions(file)
		} else {
			dimensions, err = GetImageDimensions(file)
		}
		if err != nil {
			return false, Media{}, err
		}
	}

	if isVideo {
		dimensions, duration, hasSound, err = AnalyzeVideo(filename)
		if err != nil {
			return false, Media{}, err
		}
	}

	if isAudio {
		duration = AnalyzeAudio(file)
		hasSound = true
	}

	if isPDF {
		dimensions, duration, err = AnalyzePDF(filename)
		if err != nil {
			return false, Media{}, err
		}
	}

	analyzedMedia := Media{
		ID:          slugify.Marshal(filepathBaseNoExt(filename), true),
		Alt:         embedDeclaration.Alt,
		Title:       embedDeclaration.Title,
		Source:      embedDeclaration.Source,
		Path:        ctx.RelativePathToMedia(embedDeclaration),
		Attributes:  embedDeclaration.Attributes,
		ContentType: contentType,
		Dimensions:  dimensions,
		Duration:    duration,
		Size:        uint64(fileInfo.Size()),
		HasSound:    hasSound,
	}
	return false, analyzedMedia, nil
}

func (ctx *RunContext) UseCache(filename string, embedDeclaration MediaEmbedDeclaration) (used bool, media Media) {
	if ctx.Flags.NoCache {
		return
	}
	stat, err := os.Stat(filename)
	if err != nil {
		ctx.LogInfo("Cache miss for %s: file not found: %s", filename, err)
		return
	}

	if stat.ModTime().After(ctx.BuildMetadata.PreviousBuildDate) {
		ctx.LogInfo("Cache miss for %s: modification date is %v versus %v for date of building", filename, stat.ModTime(), ctx.BuildMetadata.PreviousBuildDate)
		return
	}

	if found, analyzedMedia := FindMedia(ctx.PreviousBuiltDatabase, embedDeclaration); found {
		return true, analyzedMedia
	}

	ctx.LogInfo("Cache miss for %s: media not found in previous database build", filename)
	return

}

func (ctx *RunContext) RelativePathToMedia(embedDeclaration MediaEmbedDeclaration) string {
	if ctx.Flags.Scattered {
		return path.Clean(path.Join(ctx.Config.ScatteredModeFolder, embedDeclaration.Source))
	} else {
		return path.Clean(embedDeclaration.Source)
	}
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

// AnalyzePDF returns an ImageDimensions struct for the first page of the PDF file at filename. It also returns the number of pages.
func AnalyzePDF(filename string) (dimensions ImageDimensions, pagesCount uint, err error) {
	document, err := fitz.New(filename)
	if err != nil {
		return dimensions, pagesCount, fmt.Errorf("while opening PDF: %w", err)
	}

	defer document.Close()

	firstPage, err := document.Image(1)
	if err != nil {
		return dimensions, pagesCount, fmt.Errorf("while getting an image of the PDF's first page: %w", err)
	}

	width := firstPage.Bounds().Size().X
	height := firstPage.Bounds().Size().Y

	return ImageDimensions{
		Width:       int(width),
		Height:      int(height),
		AspectRatio: float32(width) / float32(height),
	}, uint(document.NumPage()), nil
}

// AnalyzeVideo returns an ImageDimensions struct with the video's height, width and aspect ratio and a duration in seconds.
func AnalyzeVideo(filename string) (dimensions ImageDimensions, duration uint, hasSound bool, err error) {
	probe, err := ffmpeg.DefaultConfiguration()
	if err != nil {
		return
	}
	_, video, err := probe.Probe(filename)
	if err != nil {
		return
	}
	for _, stream := range video.Streams {
		if stream.CodecType == "audio" {
			hasSound = true
		}
		if stream.CodecType == "video" {
			dimensions = ImageDimensions{
				Height:      stream.Height,
				Width:       stream.Width,
				AspectRatio: float32(stream.Width) / float32(stream.Height),
			}
		}
	}
	durationFloat, parseErr := strconv.ParseFloat(video.Format.Duration, 64)
	if parseErr != nil {
		err = fmt.Errorf("couldn't convert media duration %#v to number: %w", video.Format.Duration, parseErr)
		return
	}
	duration = uint(durationFloat)
	return
}

func (ctx *RunContext) HandleMedia(workID string, embedDeclaration MediaEmbedDeclaration, language string) (Media, error) {
	usedCache, media, err := ctx.AnalyzeMediaFile(workID, embedDeclaration)
	if err != nil {
		return Media{}, fmt.Errorf("while analyzing media: %w", err)
	}

	// Copy over
	if ctx.Config.Media.At == "" {
		return Media{}, errors.New("please specify a destination for the media files in the configuration file (set media.at)")
	}

	var content []byte
	absolutePathSource := path.Join(ctx.DatabaseDirectory, workID, media.Path)

	if absolutePathSource != ctx.AbsolutePathToMedia(media) {
		err = os.MkdirAll(path.Dir(ctx.AbsolutePathToMedia(media)), 0o755)
		if err != nil {
			return Media{}, fmt.Errorf("could not create output directory for %s: %w", ctx.AbsolutePathToMedia(media), err)
		}
		if media.ContentType == "directory" {
			err = recurcopy.CopyDirectory(absolutePathSource, ctx.AbsolutePathToMedia(media))
		} else {
			content, err = os.ReadFile(absolutePathSource)
			err = os.WriteFile(ctx.AbsolutePathToMedia(media), content, 0777)
		}

		if err != nil {
			return Media{}, fmt.Errorf("while copying media over: %w", err)
		}
	}

	// Make thumbnail
	var filenameForColorExtraction string
	media.Thumbnails = make(map[uint16]string)
	if !media.Thumbnailable() || !ctx.Config.MakeThumbnails.Enabled {
		if strings.HasPrefix(media.ContentType, "image/") {
			filenameForColorExtraction = ctx.AbsolutePathToMedia(media)
		}
	} else {
		smallestThumbSize := ctx.Config.MakeThumbnails.Sizes[0]
		for _, size := range ctx.Config.MakeThumbnails.Sizes {
			saveTo := ctx.ComputeOutputThumbnailFilename(media, workID, size, language)

			if _, err := os.Stat(saveTo); err == nil && usedCache {
				media.Thumbnails[size] = saveTo
				continue
			}

			if size < smallestThumbSize {
				smallestThumbSize = size
				filenameForColorExtraction = saveTo
			}

			// Create potentially missing directories
			os.MkdirAll(filepath.Dir(saveTo), 0777)

			ctx.Status(StepThumbnails, ProgressDetails{
				Resolution: int(size),
				File:       ctx.AbsolutePathToMedia(media),
			})

			// Make the thumbnail
			err := ctx.MakeThumbnail(media, size, saveTo)
			if err != nil {
				return media, fmt.Errorf("while making thumbnail for %s: %w", workID, err)
			}
			media.Thumbnails[size] = saveTo

		}
	}

	// Extractor colors
	if ctx.Config.ExtractColors.Enabled && filenameForColorExtraction != "" && !usedCache {
		extracted, err := ExtractColors(filenameForColorExtraction)
		if err != nil {
			return media, fmt.Errorf("while extracting colors for %s: %w", workID, err)
		} else {
			media.ExtractedColors = extracted
		}
	}

	return media, nil
}
