package ortfodb

// Functions to analyze media files.
// Used to go from a ParsedDescription struct to a Work struct.

import (
	"errors"
	"fmt"
	"image"
	"time"

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
	// "github.com/gen2brain/go-fitz" // FIXME requires cgo, which goreleaser has a hard time with
	"github.com/lafriks/go-svg"
	"github.com/metal3d/go-slugify"
	recurcopy "github.com/plus3it/gorecurcopy"
	ffmpeg "github.com/ssttevee/go-ffmpeg"
	"github.com/tcolgate/mp3"
)

func (p FilePathInsideMediaRoot) Absolute(ctx *RunContext) string {
	result, _ := filepath.Abs(filepath.Join(ctx.Config.Media.At, string(p)))
	return result
}

// RelativeToMediaRoot returns the path to the media file relative to the media root.
//
//	input:   ./[work id]/[scattered mode folder]/[file path]
//	                                             -----------
//	                                             part of the path
//	output:  ./[work id]/[scattered mode folder]/[file path]
//	         -----------------------------------------------
//	         part of the path
func (p FilePathInsidePortfolioFolder) RelativeToMediaRoot(ctx *RunContext, workID string) FilePathInsideMediaRoot {
	return FilePathInsideMediaRoot(filepath.Join(workID, ctx.Config.ScatteredModeFolder, string(p)))
}

// ImageDimensions represents metadata about a media as it's extracted from its file.
type ImageDimensions struct {
	Width       int     `json:"width"`       // Width in pixels
	Height      int     `json:"height"`      // Height in pixels
	AspectRatio float32 `json:"aspectRatio"` // width / height
}

// Media represents a media object inserted in the work object's media array.
type Media struct {
	Alt               string                        `json:"alt"`
	Caption           string                        `json:"caption"`
	RelativeSource    FilePathInsidePortfolioFolder `json:"relativeSource"`
	DistSource        FilePathInsideMediaRoot       `json:"distSource"`
	ContentType       string                        `json:"contentType"`
	Size              int                           `json:"size"` // in bytes
	Dimensions        ImageDimensions               `json:"dimensions"`
	Online            bool                          `json:"online"`
	Duration          float64                       `json:"duration"` // in seconds
	HasSound          bool                          `json:"hasSound"`
	Colors            ColorPalette                  `json:"colors"`
	Thumbnails        ThumbnailsMap                 `json:"thumbnails"`
	ThumbnailsBuiltAt string                        `json:"thumbnailsBuiltAt"`
	Attributes        MediaAttributes               `json:"attributes"`
	Analyzed          bool                          `json:"analyzed"` // whether the media has been analyzed
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
// TODO prevent duplicate analysis of the same file in the current session even when file was never analyzed on previous runs of the command
func (ctx *RunContext) AnalyzeMediaFile(workID string, embedDeclaration Media) (usedCache bool, analyzedMedia Media, anchor string, err error) {
	LogDebug("Analyzing media %#v", embedDeclaration)

	// Compute absolute filepath to media
	var filename string
	if !filepath.IsAbs(string(embedDeclaration.RelativeSource)) {
		filename, _ = filepath.Abs(filepath.Join(ctx.PathToWorkFolder(workID), string(embedDeclaration.RelativeSource)))
	} else {
		filename = string(embedDeclaration.RelativeSource)
	}
	anchor = slugify.Marshal(filepathBaseNoExt(filename), true)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return
	}

	var contentType string

	if fileInfo.IsDir() {
		contentType = "directory"
	} else {
		usedCache, cachedAnalysis := ctx.UseCache(filename, embedDeclaration, workID)
		if usedCache && cachedAnalysis.ContentType != "" {
			LogDebug("Reusing cached analysis %#v", cachedAnalysis)
			return true, cachedAnalysis, anchor, nil
		}

		ctx.Status(workID, PhaseMediaAnalysis, string(embedDeclaration.RelativeSource))
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
	var colors ColorPalette

	if isImage {
		if contentType == "image/svg" || contentType == "image/svg+xml" {
			dimensions, err = GetSVGDimensions(file)
		} else {
			dimensions, err = GetImageDimensions(file)
		}
		if err != nil {
			return
		}
		if ctx.Config.ExtractColors.Enabled && canExtractColors(contentType) {
			colors, err = ExtractColors(filename)
			if err != nil {
				DisplayError("Could not extract colors from %s", err, filename)
				err = nil
			}
		}
	}

	if isVideo {
		dimensions, duration, hasSound, err = AnalyzeVideo(filename)
		if err != nil {
			return
		}
	}

	if isAudio {
		duration = AnalyzeAudio(file)
		hasSound = true
	}

	if isPDF {
		dimensions, duration, err = AnalyzePDF(filename)
		if err != nil {
			return
		}
	}

	analyzedMedia = Media{
		Alt:            embedDeclaration.Alt,
		Caption:        embedDeclaration.Caption,
		RelativeSource: embedDeclaration.RelativeSource,
		DistSource:     FilePathInsideMediaRoot(embedDeclaration.RelativeSource.RelativeToMediaRoot(ctx, workID)),
		Attributes:     embedDeclaration.Attributes,
		ContentType:    contentType,
		Dimensions:     dimensions,
		Duration:       float64(duration),
		Size:           int(fileInfo.Size()),
		HasSound:       hasSound,
		Colors:         colors,
		Analyzed:       true,
	}
	LogDebug("Analyzed to %#v (no cache used)", analyzedMedia)
	return
}

func (ctx *RunContext) UseCache(filename string, embedDeclaration Media, workID string) (used bool, media Media) {
	if ctx.Flags.NoCache {
		return
	}
	stat, err := os.Stat(filename)
	if err != nil {
		LogDebug("Cache miss for %s: file not found: %s", filename, err)
		return
	}

	if found, analyzedMedia, builtAt := FindMedia(ctx.PreviousBuiltDatabase, embedDeclaration, workID); found {
		LogDebug("cache hit for %s: using cache from embed decl %#v", filename, embedDeclaration)
		if stat.ModTime().After(builtAt) {
			LogDebug("Cache miss for %s: modification date is %v versus %v for date of building", filename, stat.ModTime(), ctx.BuildMetadata.PreviousBuildDate)
			return
		}
		return true, analyzedMedia
	}

	LogDebug("Cache miss for %s: media not found in previous database build", filename)
	return

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
	return ImageDimensions{}, 0, errors.New("PDF analysis is disabled")
	/* fitz requires cgo, which goreleaser has a hard time with.
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
	*/
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

func (ctx *RunContext) HandleMedia(workID string, blockID string, embedDeclaration Media, language string) (media Media, anchor string, err error) {
	usedCache, media, anchor, err := ctx.AnalyzeMediaFile(workID, embedDeclaration)
	if err != nil {
		err = fmt.Errorf("while analyzing media: %w", err)
		return
	}

	// Copy over
	if ctx.Config.Media.At == "" {
		err = errors.New("please specify a destination for the media files in the configuration file (set media.at)")
		return
	}

	var content []byte
	absolutePathSource := media.RelativeSource.Absolute(ctx, workID)
	absolutePathDestination := media.DistSource.Absolute(ctx)

	if absolutePathDestination != absolutePathSource {
		err = os.MkdirAll(path.Dir(absolutePathDestination), 0o755)
		if err != nil {
			err = fmt.Errorf("could not create output directory for %s: %w", absolutePathSource, err)
			return
		}
		if media.ContentType == "directory" {
			err = recurcopy.CopyDirectory(absolutePathSource, absolutePathDestination)
		} else {
			content, err = os.ReadFile(absolutePathSource)
			if err != nil {
				err = fmt.Errorf("could not read file %s: %w", absolutePathSource, err)
				return
			}
			err = os.WriteFile(absolutePathDestination, content, 0777)
		}

		if err != nil {
			err = fmt.Errorf("while copying media over: %w", err)
			return
		}
	}

	// Make thumbnail
	media.Thumbnails = make(map[int]FilePathInsideMediaRoot)
	if media.Thumbnailable() && ctx.Config.MakeThumbnails.Enabled {
		type result struct {
			size    int
			err     error
			skipped bool
		}
		sizesChan := make(chan int, len(ctx.Config.MakeThumbnails.Sizes))
		resultsChan := make(chan result)
		builtSizes := 0
		smallestThumbSize := ctx.Config.MakeThumbnails.Sizes[0] // XXX what?? is it sorted?
		for _, size := range ctx.Config.MakeThumbnails.Sizes {
			sizesChan <- size
		}

		for i := 0; i < 2; i++ {
			go func() {
				for {
					size := <-sizesChan
					LogDebug("Making thumbnail for %s#%s", media.RelativeSource, blockID)
					saveTo := ctx.ComputeOutputThumbnailFilename(media, blockID, workID, size, language)

					if _, err := os.Stat(string(saveTo.Absolute(ctx))); err == nil && usedCache {
						LogDebug("Skipping thumbnail creation for %s#%s because it already exists", media.RelativeSource, blockID)
						resultsChan <- result{size: size, skipped: true}
						return
					}

					if size < smallestThumbSize {
						smallestThumbSize = size
					}

					// Create potentially missing directories
					os.MkdirAll(filepath.Dir(saveTo.Absolute(ctx)), 0777)

					ctx.Status(workID, PhaseThumbnails, string(media.RelativeSource), fmt.Sprintf("%dpx", size))

					// Make the thumbnail
					err := ctx.MakeThumbnail(media, size, saveTo.Absolute(ctx))
					if err != nil {
						resultsChan <- result{err: fmt.Errorf("while making thumbnail for %s: %w", workID, err)}
						return
					}
					LogDebug("Made thumbnail %s", saveTo)
					resultsChan <- result{size: size}
				}
			}()
		}

		for builtSizes < len(ctx.Config.MakeThumbnails.Sizes) {
			result := <-resultsChan
			if result.err != nil {
				return media, anchor, result.err
			}
			if !result.skipped {
				media.Thumbnails[result.size] = ctx.ComputeOutputThumbnailFilename(media, blockID, workID, result.size, language)
				media.ThumbnailsBuiltAt = time.Now().String()
			}
			builtSizes++
		}
	}

	return
}
